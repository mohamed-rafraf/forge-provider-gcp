/*
Copyright 2024 The Forge contributors.

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

package gcpbuild

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/forge-build/forge/pkg/ssh"
	"go.uber.org/zap"

	"github.com/forge-build/forge-provider-gcp/pkg/cloud/gcp/compute/images"
	"github.com/forge-build/forge-provider-gcp/pkg/cloud/gcp/compute/instances"

	"github.com/forge-build/forge-provider-gcp/pkg/cloud"
	"github.com/forge-build/forge-provider-gcp/pkg/cloud/gcp/compute/firewalls"
	"github.com/forge-build/forge-provider-gcp/pkg/cloud/gcp/compute/networks"
	"github.com/forge-build/forge-provider-gcp/pkg/cloud/gcp/compute/subnets"
	"github.com/forge-build/forge-provider-gcp/pkg/cloud/scope"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	buildv1 "github.com/forge-build/forge/api/v1alpha1"
	forgeutil "github.com/forge-build/forge/util"
	"github.com/forge-build/forge/util/annotations"
	"github.com/forge-build/forge/util/predicates"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	infrav1 "github.com/forge-build/forge-provider-gcp/pkg/api/v1alpha1"
)

const ControllerName = "gcpbuild-controller"

// GCPBuildReconciler reconciles a GCPBuild object
type GCPBuildReconciler struct {
	client.Client
	log      *zap.SugaredLogger
	recorder record.EventRecorder
}

func (r *GCPBuildReconciler) recordEvent(gcpBuild *infrav1.GCPBuild, eventType, reason, message string) {
	r.log.Info(message)
	r.recorder.Event(gcpBuild, eventType, reason, message)
}

// +kubebuilder:rbac:groups=infrastructure.forge.build,resources=gcpbuilds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.forge.build,resources=gcpbuilds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.forge.build,resources=gcpbuilds/finalizers,verbs=update
// +kubebuilder:rbac:groups=forge.build,resources=builds,verbs=get;list;watch;patch

func (r *GCPBuildReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	r.log = r.log.With("GCPBuild", req.NamespacedName)
	r.log.Info("Reconciling")
	gcpBuild := &infrav1.GCPBuild{}
	err := r.Get(ctx, req.NamespacedName, gcpBuild)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.log.Info("GCPBuild resource not found or already deleted")
			return ctrl.Result{}, nil
		}

		r.log.Error(err, "Unable to fetch GCPBuild resource")
		return ctrl.Result{}, err
	}

	// Fetch the Build.
	build, err := forgeutil.GetOwnerBuild(ctx, r.Client, gcpBuild.ObjectMeta)
	if err != nil {
		r.log.Error(err, "Failed to get owner build")
		return ctrl.Result{}, err
	}
	if build == nil {
		r.log.Info("Build Controller has not yet set OwnerRef")
		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(build, gcpBuild) {
		r.log.Info("GCPBuild of linked Build is marked as paused. Won't reconcile")
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
	r.log.Info("Reconciling Delete GCPBuild")

	reconcilers := []cloud.Reconciler{
		instances.New(buildScope),
		subnets.New(buildScope),
		firewalls.New(buildScope),
		networks.New(buildScope),
	}

	for _, reconcile := range reconcilers {
		if err := reconcile.Delete(ctx); err != nil {
			r.log.Error(err, "Reconcile error")
			r.recordEvent(buildScope.GCPBuild, "Warning", "Cleaning Up Failed", fmt.Sprintf("Reconcile error - %v ", err))
			return err
		}
	}

	controllerutil.RemoveFinalizer(buildScope.GCPBuild, infrav1.BuildFinalizer)
	r.recordEvent(buildScope.GCPBuild, "Normal", "Reconciled", fmt.Sprintf("%s is reconciled successfully ", buildScope.GCPBuild.Name))
	buildScope.SetCleanedUP()
	return nil
}
func (r *GCPBuildReconciler) reconcileNormal(ctx context.Context, buildScope *scope.BuildScope) (ctrl.Result, error) {
	reconcilers := []cloud.Reconciler{
		networks.New(buildScope),
		firewalls.New(buildScope),
		subnets.New(buildScope),
		instances.New(buildScope),
		images.New(buildScope),
	}

	if !buildScope.IsReady() {
		for _, reconciler := range reconcilers {
			if err := reconciler.Reconcile(ctx); err != nil {
				r.log.Error(err, "Reconcile error")
				r.recordEvent(buildScope.GCPBuild, "Warning", "Building Failed", fmt.Sprintf("Reconcile error - %v ", err))
				return ctrl.Result{}, err
			}
		}
	}

	r.log.Info("Reconciling GCPBuild")

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

	if buildScope.IsReady() && !buildScope.IsCleanedUP() {
		return ctrl.Result{}, r.reconcileDelete(ctx, buildScope)
	}

	if buildScope.GetInstanceID() == nil {
		r.recordEvent(buildScope.GCPBuild, "Normal", "GCPBuildReconciler", "GCPBuild does not started the build yet ")

		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	r.recordEvent(buildScope.GCPBuild, "Normal", "InstanceCreated", fmt.Sprintf("Machine is created, Got an instance ID - %s ", *buildScope.GetInstanceID()))

	buildScope.SetMachineReady()

	if buildScope.GCPBuild.Status.ArtifactRef == nil {
		r.recordEvent(buildScope.GCPBuild, "Normal", "WaitBuilding", "Artifact is not available yet ")

		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	r.recordEvent(buildScope.GCPBuild, "Normal", "ImageReady", fmt.Sprintf("Got an available Artifact - %s ", *buildScope.GCPBuild.Status.ArtifactRef))

	buildScope.SetReady()
	r.recordEvent(buildScope.GCPBuild, "Normal", "Reconciled", "GCP Build is reconciled successfully ")

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

// Add creates a new GCPBuild controller and adds it to the Manager.
func Add(ctx context.Context, mgr ctrl.Manager, numWorkers int, log *zap.SugaredLogger) error {
	// Create the reconciler instance
	reconciler := &GCPBuildReconciler{
		Client:   mgr.GetClient(),
		log:      log,
		recorder: mgr.GetEventRecorderFor(ControllerName),
	}

	// Set up the controller with custom predicates
	_, err := builder.ControllerManagedBy(mgr).
		Named(ControllerName).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: numWorkers,
		}).
		For(&infrav1.GCPBuild{}).
		WithEventFilter(predicates.ResourceNotPaused(ctrl.LoggerFrom(ctx))).
		Watches(&buildv1.Build{},
			handler.EnqueueRequestsFromMapFunc(forgeutil.BuildToInfrastructureMapFunc(ctx, infrav1.GroupVersion.WithKind(infrav1.GCPBuildKind), mgr.GetClient(), &infrav1.GCPBuild{})),
			builder.WithPredicates(predicates.BuildUnpaused(ctrl.LoggerFrom(ctx)))).
		Build(reconciler)

	return err
}
