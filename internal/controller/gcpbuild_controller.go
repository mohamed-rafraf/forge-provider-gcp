/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/forge-build/forge/pkg/ssh"

	"github.com/forge-build/forge-provider-gcp/cloud/services/compute/instances"
	"github.com/forge-build/forge-provider-gcp/cloud/services/compute/machineimages"

	"github.com/forge-build/forge-provider-gcp/cloud"
	"github.com/forge-build/forge-provider-gcp/cloud/scope"
	"github.com/forge-build/forge-provider-gcp/cloud/services/compute/firewalls"
	"github.com/forge-build/forge-provider-gcp/cloud/services/compute/networks"
	"github.com/forge-build/forge-provider-gcp/cloud/services/compute/subnets"
	"github.com/pkg/errors"
	"sigs.k8s.io/cluster-api/util/record"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	buildv1 "github.com/forge-build/forge/api/v1alpha1"
	forgeutil "github.com/forge-build/forge/util"
	"github.com/forge-build/forge/util/annotations"
	"github.com/forge-build/forge/util/predicates"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/forge-build/forge-provider-gcp/api/v1alpha1"
)

// GCPBuildReconciler reconciles a GCPBuild object
type GCPBuildReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.forge.build,resources=gcpbuilds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.forge.build,resources=gcpbuilds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.forge.build,resources=gcpbuilds/finalizers,verbs=update
// +kubebuilder:rbac:groups=forge.build,resources=builds,verbs=get;list;watch;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *GCPBuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	logger := log.FromContext(ctx)
	gcpBuild := &infrav1.GCPBuild{}
	err := r.Get(ctx, req.NamespacedName, gcpBuild)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("GCPBuild resource not found or already deleted")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Unable to fetch GCPBuild resource")
		return ctrl.Result{}, err
	}

	// Fetch the Build.
	build, err := forgeutil.GetOwnerBuild(ctx, r.Client, gcpBuild.ObjectMeta)
	if err != nil {
		logger.Error(err, "Failed to get owner build")
		return ctrl.Result{}, err
	}
	if build == nil {
		logger.Info("Build Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(build, gcpBuild) {
		logger.Info("GCPBuild of linked Build is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	buildScope, err := scope.NewBuildScope(ctx, scope.BuildScopeParams{
		Client:   r.Client,
		Build:    build,
		GCPBuild: gcpBuild,
	})
	if err != nil {
		return ctrl.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any GCPBuild changes.
	defer func() {
		if err := buildScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted clusters
	if !gcpBuild.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.reconcileDelete(ctx, buildScope)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, buildScope)
}

func (r *GCPBuildReconciler) reconcileDelete(ctx context.Context, buildScope *scope.BuildScope) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Delete GCPBuild")

	reconcilers := []cloud.Reconciler{
		instances.New(buildScope),
		subnets.New(buildScope),
		firewalls.New(buildScope),
		networks.New(buildScope),
	}

	for _, r := range reconcilers {
		if err := r.Delete(ctx); err != nil {
			logger.Error(err, "Reconcile error")
			record.Warnf(buildScope.GCPBuild, "GCPBuildReconciler", "Reconcile error - %v", err)
			return err
		}
	}

	controllerutil.RemoveFinalizer(buildScope.GCPBuild, infrav1.BuildFinalizer)
	record.Event(buildScope.GCPBuild, "GCPBuildReconciler", "Reconciled")
	buildScope.SetCleanUpReady()
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GCPBuildReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1.GCPBuild{}).
		WithEventFilter(predicates.ResourceNotPaused(ctrl.LoggerFrom(ctx))).
		Watches(&buildv1.Build{},
			handler.EnqueueRequestsFromMapFunc(forgeutil.BuildToInfrastructureMapFunc(ctx, infrav1.GroupVersion.WithKind(infrav1.GCPBuildKind), mgr.GetClient(), &infrav1.GCPBuild{})),
			builder.WithPredicates(predicates.BuildUnpaused(ctrl.LoggerFrom(ctx)))).
		Complete(r)
}

func (r *GCPBuildReconciler) reconcileNormal(ctx context.Context, buildScope *scope.BuildScope) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling GCPBuild")

	controllerutil.AddFinalizer(buildScope.GCPBuild, infrav1.BuildFinalizer)
	if err := buildScope.PatchObject(); err != nil {
		return ctrl.Result{}, err
	}

	// get ssh key
	sshKey, err := r.GetSSHKey(ctx, buildScope)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to get an ssh-key")
	}
	buildScope.SetSSHKey(sshKey)

	reconcilers := []cloud.Reconciler{
		networks.New(buildScope),
		firewalls.New(buildScope),
		subnets.New(buildScope),
		// TODO, make sure to provide a way of connecting to this machine.
		instances.New(buildScope),
		machineimages.New(buildScope),
		// TODO, createConnectionSecret.
		// TODO exportImage,
		// TODO cleanup GCP resources.
	}
	if !buildScope.IsReady() {
		for _, r := range reconcilers {
			if err := r.Reconcile(ctx); err != nil {
				logger.Error(err, "Reconcile error")
				record.Warnf(buildScope.GCPBuild, "GCPBuildReconciler", "Reconcile error - %v", err)
				return ctrl.Result{}, err
			}
		}
	}

	if buildScope.IsReady() && !buildScope.IsCleanedUp() {
		return ctrl.Result{}, r.reconcileDelete(ctx, buildScope)
	}

	if buildScope.GetInstanceID() == nil {
		logger.Info("GCPBuild does not started the build yet. Reconciling")
		record.Event(buildScope.GCPBuild, "GCPBuildReconciler", "Waiting for build to start")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	record.Eventf(buildScope.GCPBuild, "GCPBuildReconciler", "Machine is created, Got an instance ID - %s", *buildScope.GetInstanceID())
	buildScope.SetMachineReady()

	if buildScope.GCPBuild.Status.ArtifactRef == nil {
		logger.Info("Artifact is not available yet. Reconciling")
		record.Event(buildScope.GCPBuild, "GCPBuildReconciler", "Waiting for image build")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	record.Eventf(buildScope.GCPBuild, "GCPBuildReconciler", "Got an available Artifact - %s", *buildScope.GCPBuild.Status.ArtifactRef)
	buildScope.SetReady()
	record.Event(buildScope.GCPBuild, "GCPClusterReconcile", "Reconciled")
	return ctrl.Result{}, nil
}

func (r *GCPBuildReconciler) GetSSHKey(ctx context.Context, buildScope *scope.BuildScope) (key scope.SSHKey, err error) {
	if buildScope.GCPBuild.Spec.GenerateSSHKey {
		sshKey, err := ssh.NewKeyPair()
		if err != nil {
			return key, errors.Wrap(err, "cannot generate ssh key")
		}

		return scopeSSHKey(buildScope.GCPBuild.Spec.Username, string(sshKey.PrivateKey), string(sshKey.PublicKey)), nil
	}

	if buildScope.GCPBuild.Spec.SSHCredentialsRef != nil {
		secret, err := forgeutil.GetSecretFromSecretReference(ctx, r.Client, *buildScope.GCPBuild.Spec.SSHCredentialsRef)
		if err != nil {
			return scope.SSHKey{}, errors.Wrap(err, "unable to get ssh credentials secret")
		}

		_, _, privKey := ssh.GetCredentialsFromSecret(secret)

		pubKey, err := ssh.GetPublicKeyFromPrivateKey(privKey)
		if err != nil {
			return scope.SSHKey{}, errors.Wrap(err, "unable to get public key")
		}

		return scopeSSHKey(buildScope.GCPBuild.Spec.Username, privKey, pubKey), nil
	}

	return scope.SSHKey{}, errors.New("no ssh key provided, consider using spec.generateSSHKey or provide a private key")
}

func scopeSSHKey(username, privateKey, pubKey string) scope.SSHKey {
	sshPublicKey := strings.TrimSuffix(pubKey, "\n")

	return scope.SSHKey{
		MetadataSSHKeys: fmt.Sprintf("%s:%s %s", username, sshPublicKey, username),
		PrivateKey:      privateKey,
		PublicKey:       pubKey,
	}
}
