/*
Copyright 2024 The Forge Authors.

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

package scope

import (
	"context"
	"fmt"
	"path"
	"strconv"

	"github.com/forge-build/forge/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"

	buildv1 "github.com/forge-build/forge/api/v1alpha1"
	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/forge-build/forge-provider-gcp/pkg/api/v1alpha1"
	"github.com/forge-build/forge-provider-gcp/pkg/cloud"
)

// BuildScopeParams defines the input parameters used to create a new Scope.
type BuildScopeParams struct {
	GCPServices
	Client      client.Client
	Build       *buildv1.Build
	GCPBuild    *infrav1.GCPBuild
	BuildGetter cloud.BuildGetter
}

// NewBuildScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewBuildScope(ctx context.Context, params BuildScopeParams) (*BuildScope, error) {
	if params.Build == nil {
		return nil, errors.New("failed to generate new scope from nil Build")
	}
	if params.GCPBuild == nil {
		return nil, errors.New("failed to generate new scope from nil GCPBuild")
	}

	if params.GCPServices.Compute == nil {
		computeSvc, err := newComputeService(ctx, params.GCPBuild.Spec.CredentialsRef, params.Client)
		if err != nil {
			return nil, errors.Errorf("failed to create gcp compute client: %v", err)
		}

		params.GCPServices.Compute = computeSvc
	}

	helper, err := patch.NewHelper(params.GCPBuild, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &BuildScope{
		client:      params.Client,
		Build:       params.Build,
		GCPBuild:    params.GCPBuild,
		GCPServices: params.GCPServices,
		patchHelper: helper,
	}, nil
}

// BuildScope defines the basic context for an actuator to operate upon.
type BuildScope struct {
	client      client.Client
	patchHelper *patch.Helper

	Build    *buildv1.Build
	GCPBuild *infrav1.GCPBuild
	//BuildGetter cloud.BuildGetter
	GCPServices

	sshKEy SSHKey
}

const sshMetaKey = "ssh-keys"

type SSHKey struct {
	MetadataSSHKeys string
	PrivateKey      string
	PublicKey       string
}

// Cloud returns initialized cloud.
func (s *BuildScope) Cloud() cloud.Cloud {
	return newCloud(s.Project(), s.GCPServices)
}

// NetworkCloud returns initialized cloud.
func (s *BuildScope) NetworkCloud() cloud.Cloud {
	return newCloud(s.NetworkProject(), s.GCPServices)
}

// Project returns the current project name.
func (s *BuildScope) Project() string {
	return s.GCPBuild.Spec.Project
}

// NetworkProject returns the project name where network resources should exist.
// The network project defaults to the Project when one is not supplied.
func (s *BuildScope) NetworkProject() string {
	return ptr.Deref(s.GCPBuild.Spec.Network.HostProject, s.Project())
}

// IsSharedVpc returns true If sharedVPC used else , returns false.
func (s *BuildScope) IsSharedVpc() bool {
	return s.NetworkProject() != s.Project()
}

// Region returns the cluster region.
func (s *BuildScope) Region() string {
	return s.GCPBuild.Spec.Region
}

// Name returns the cluster name.
func (s *BuildScope) Name() string {
	return s.Build.Name
}

// Namespace returns the cluster namespace.
func (s *BuildScope) Namespace() string {
	return s.Build.Namespace
}

// NetworkName returns the cluster network unique identifier.
func (s *BuildScope) NetworkName() string {
	return ptr.Deref(s.GCPBuild.Spec.Network.Name, "default")
}

// NetworkLink returns the partial URL for the network.
func (s *BuildScope) NetworkLink() string {
	return fmt.Sprintf("projects/%s/global/networks/%s", s.NetworkProject(), s.NetworkName())
}

// Network returns the cluster network object.
func (s *BuildScope) Network() *infrav1.Network {
	return &s.GCPBuild.Status.Network
}

// AdditionalLabels returns the cluster additional labels.
func (s *BuildScope) AdditionalLabels() infrav1.Labels {
	return s.GCPBuild.Spec.AdditionalLabels
}

// GetInstanceID returns the build instanceID
func (s *BuildScope) GetInstanceID() *string {
	return s.GCPBuild.Spec.InstanceID
}

// SetInstanceID sets the build InstanceID.
func (s *BuildScope) SetInstanceID(instanceID string) {
	s.GCPBuild.Spec.InstanceID = ptr.To(instanceID)
}

// GetInstanceStatus returns the GCPBuild instance status.
func (s *BuildScope) GetInstanceStatus() *infrav1.InstanceStatus {
	return s.GCPBuild.Status.InstanceStatus
}

// SetInstanceStatus sets the GCPMachine instance status.
func (s *BuildScope) SetInstanceStatus(v infrav1.InstanceStatus) {
	s.GCPBuild.Status.InstanceStatus = &v
}

// GetSSHKey returns the ssh key.
func (s *BuildScope) GetSSHKey() SSHKey {
	return s.sshKEy
}

// SetSSHKey sets ssh key.
func (s *BuildScope) SetSSHKey(key SSHKey) {
	s.sshKEy = key
}

// SetReady sets cluster ready status.
func (s *BuildScope) SetReady() {
	s.GCPBuild.Status.Ready = true
}

// SetMachineReady sets build machine ready status.
func (s *BuildScope) SetMachineReady() {
	s.GCPBuild.Status.MachineReady = true
}

// SetBuildReady sets cleanup ready status.
func (s *BuildScope) SetCleanedUP() {
	s.GCPBuild.Status.CleanedUP = true
}

func (s *BuildScope) SetArtifactRef(reference string) {
	s.GCPBuild.Status.ArtifactRef = &reference
}

// NetworkSpec returns google compute network spec.
func (s *BuildScope) NetworkSpec() *compute.Network {
	createSubnet := ptr.Deref(s.GCPBuild.Spec.Network.AutoCreateSubnetworks, true)
	network := &compute.Network{
		Name:                  s.NetworkName(),
		Description:           infrav1.BuildTagKey(s.Name()),
		AutoCreateSubnetworks: createSubnet,
		ForceSendFields:       []string{"AutoCreateSubnetworks"},
	}

	return network
}

// NatRouterSpec returns google compute nat router spec.
func (s *BuildScope) NatRouterSpec() *compute.Router {
	networkSpec := s.NetworkSpec()
	return &compute.Router{
		Name: fmt.Sprintf("%s-%s", networkSpec.Name, "router"),
		Nats: []*compute.RouterNat{
			{
				Name:                          fmt.Sprintf("%s-%s", networkSpec.Name, "nat"),
				NatIpAllocateOption:           "AUTO_ONLY",
				SourceSubnetworkIpRangesToNat: "ALL_SUBNETWORKS_ALL_IP_RANGES",
			},
		},
	}
}

// SubnetSpecs returns google compute subnets spec.
func (s *BuildScope) SubnetSpecs() []*compute.Subnetwork {
	subnets := []*compute.Subnetwork{}
	for _, subnetwork := range s.GCPBuild.Spec.Network.Subnets {
		secondaryIPRanges := []*compute.SubnetworkSecondaryRange{}
		for rangeName, secondaryCidrBlock := range subnetwork.SecondaryCidrBlocks {
			secondaryIPRanges = append(secondaryIPRanges, &compute.SubnetworkSecondaryRange{RangeName: rangeName, IpCidrRange: secondaryCidrBlock})
		}
		subnets = append(subnets, &compute.Subnetwork{
			Name:                  subnetwork.Name,
			Region:                subnetwork.Region,
			EnableFlowLogs:        ptr.Deref(subnetwork.EnableFlowLogs, false),
			PrivateIpGoogleAccess: ptr.Deref(subnetwork.PrivateGoogleAccess, false),
			IpCidrRange:           subnetwork.CidrBlock,
			SecondaryIpRanges:     secondaryIPRanges,
			Description:           ptr.Deref(subnetwork.Description, infrav1.BuildTagKey(s.Name())),
			Network:               s.NetworkLink(),
			Purpose:               ptr.Deref(subnetwork.Purpose, "PRIVATE_RFC_1918"),
			Role:                  "ACTIVE",
		})
	}

	return subnets
}

// FirewallRulesSpec returns google compute firewall spec.
func (s *BuildScope) FirewallRulesSpec() []*compute.Firewall {
	firewallRules := []*compute.Firewall{
		{
			Name:    fmt.Sprintf("allow-%s-healthchecks", s.Name()),
			Network: s.NetworkLink(),
			Allowed: []*compute.FirewallAllowed{
				{
					IPProtocol: "TCP",
					Ports: []string{
						strconv.FormatInt(6443, 10),
					},
				},
			},
			Direction: "INGRESS",
			SourceRanges: []string{
				"35.191.0.0/16",
				"130.211.0.0/22",
			},
			TargetTags: []string{
				s.Name() + "-control-plane",
			},
		},
		{
			Name:    fmt.Sprintf("allow-%s-cluster", s.Name()),
			Network: s.NetworkLink(),
			Allowed: []*compute.FirewallAllowed{
				{
					IPProtocol: "all",
				},
			},
			Direction: "INGRESS",
			SourceTags: []string{
				s.Name() + "-control-plane",
				s.Name() + "-node",
			},
			TargetTags: []string{
				s.Name() + "-control-plane",
				s.Name() + "-node",
			},
		},
	}

	return firewallRules
}

// Zone returns the FailureDomain for the GCPBuild.
func (s *BuildScope) Zone() string {
	return s.GCPBuild.Spec.Zone
}

// InstanceImageSpec returns compute instance image attched-disk spec.
func (s *BuildScope) InstanceImageSpec() *compute.AttachedDisk {
	sourceImage := path.Join("global", "images", "ubuntu-2204-jammy-v20240904")
	if s.GCPBuild.Spec.Image != nil {
		sourceImage = *s.GCPBuild.Spec.Image
	} else if s.GCPBuild.Spec.ImageFamily != nil {
		sourceImage = *s.GCPBuild.Spec.ImageFamily
	}

	diskType := infrav1.PdStandardDiskType
	if t := s.GCPBuild.Spec.RootDeviceType; t != nil {
		diskType = *t
	}

	disk := &compute.AttachedDisk{
		AutoDelete: true,
		Boot:       true,
		InitializeParams: &compute.AttachedDiskInitializeParams{
			DiskSizeGb:  s.GCPBuild.Spec.RootDeviceSize,
			DiskType:    path.Join("zones", s.Zone(), "diskTypes", string(diskType)),
			SourceImage: sourceImage,
			Labels:      s.AdditionalLabels().AddLabels(s.GCPBuild.Spec.AdditionalLabels),
		},
	}

	return disk
}

// InstanceNetworkInterfaceSpec returns compute network interface spec.
func (s *BuildScope) InstanceNetworkInterfaceSpec() *compute.NetworkInterface {
	networkInterface := &compute.NetworkInterface{
		Network: path.Join("projects", s.NetworkProject(), "global", "networks", s.NetworkName()),
	}

	if s.GCPBuild.Spec.PublicIP != nil && *s.GCPBuild.Spec.PublicIP {
		networkInterface.AccessConfigs = []*compute.AccessConfig{
			{
				Type: "ONE_TO_ONE_NAT",
				Name: "External NAT",
			},
		}
	}

	if s.GCPBuild.Spec.Subnet != nil {
		networkInterface.Subnetwork = path.Join("projects", s.NetworkProject(), "regions", s.Region(), "subnetworks", *s.GCPBuild.Spec.Subnet)
	}

	return networkInterface
}

// InstanceServiceAccountsSpec returns service-account spec.
func (s *BuildScope) InstanceServiceAccountsSpec() *compute.ServiceAccount {
	serviceAccount := &compute.ServiceAccount{
		Email: "default",
		Scopes: []string{
			compute.CloudPlatformScope,
		},
	}

	if s.GCPBuild.Spec.ServiceAccount != nil {
		serviceAccount.Email = s.GCPBuild.Spec.ServiceAccount.Email
		serviceAccount.Scopes = s.GCPBuild.Spec.ServiceAccount.Scopes
	}

	return serviceAccount
}

// InstanceAdditionalMetadataSpec returns additional metadata spec.
func (s *BuildScope) InstanceAdditionalMetadataSpec() *compute.Metadata {
	metadata := new(compute.Metadata)
	for _, additionalMetadata := range s.GCPBuild.Spec.AdditionalMetadata {
		metadata.Items = append(metadata.Items, &compute.MetadataItems{
			Key:   additionalMetadata.Key,
			Value: additionalMetadata.Value,
		})
	}

	// Add the ssh keys.
	metadata.Items = append(metadata.Items, &compute.MetadataItems{
		Key:   sshMetaKey,
		Value: &s.sshKEy.MetadataSSHKeys,
	})

	return metadata
}

// InstanceSpec returns instance spec.
func (s *BuildScope) InstanceSpec(log logr.Logger) *compute.Instance {
	instance := &compute.Instance{
		Name:        s.Name(),
		Zone:        s.Zone(),
		MachineType: path.Join("zones", s.Zone(), "machineTypes", s.GCPBuild.Spec.InstanceType),
		Tags: &compute.Tags{
			Items: append(
				s.GCPBuild.Spec.AdditionalNetworkTags,
				fmt.Sprintf("%s-%s", s.Name(), "forge-builder"),
				s.Name(),
			),
		},
		Labels: infrav1.Build(infrav1.BuildParams{
			BuildName: s.Name(),
			Lifecycle: infrav1.ResourceLifecycleOwned,
			//nolint: godox
			// TODO: Check what needs to be added for the cloud provider label.
			Additional: s.AdditionalLabels().AddLabels(s.GCPBuild.Spec.AdditionalLabels),
		}),
		Scheduling: &compute.Scheduling{
			Preemptible: s.GCPBuild.Spec.Preemptible,
		},
	}

	instance.Disks = append(instance.Disks, s.InstanceImageSpec())
	instance.Metadata = s.InstanceAdditionalMetadataSpec()
	instance.ServiceAccounts = append(instance.ServiceAccounts, s.InstanceServiceAccountsSpec())
	instance.NetworkInterfaces = append(instance.NetworkInterfaces, s.InstanceNetworkInterfaceSpec())
	return instance
}

// ANCHOR_END: MachineInstanceSpec

// GetBootstrapData returns the bootstrap data from the secret in the Machine's bootstrap.dataSecretName.
func (s *BuildScope) GetBootstrapData() (string, error) {
	if s.GCPBuild.Spec.Bootstrap.DataSecretName == nil {
		return "", nil
	}

	secret := &corev1.Secret{}
	key := types.NamespacedName{Namespace: s.Namespace(), Name: *s.GCPBuild.Spec.Bootstrap.DataSecretName}
	if err := s.client.Get(context.TODO(), key, secret); err != nil {
		return "", errors.Wrapf(err, "failed to retrieve bootstrap data secret for GCPBuild %s/%s", s.Namespace(), s.Name())
	}

	value, ok := secret.Data["value"]
	if !ok {
		return "", errors.New("error retrieving bootstrap data: secret value key is missing")
	}

	return string(value), nil
}

func (s *BuildScope) EnsureCredentialsSecret(ctx context.Context, host string) error {
	err := util.EnsureCredentialsSecret(ctx, s.client, s.Build, util.SSHCredentials{
		Host:       host,
		Username:   s.GCPBuild.Spec.Username,
		PrivateKey: s.sshKEy.PrivateKey,
		PublicKey:  s.sshKEy.PublicKey,
	}, "gcp")
	if err != nil {
		return err
	}
	return nil
}

// PatchObject persists the cluster configuration and status.
func (s *BuildScope) PatchObject() error {
	return s.patchHelper.Patch(context.TODO(), s.GCPBuild)
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *BuildScope) Close() error {
	return s.PatchObject()
}

func (s *BuildScope) ImageName() string {
	return fmt.Sprintf("%s-%s", "forge", s.Name())
}

func (s *BuildScope) IsProvisionerReady() bool {
	return s.Build.Status.ProvisionersReady
}

func (s *BuildScope) IsReady() bool {
	return s.GCPBuild.Status.Ready
}

func (s *BuildScope) IsCleanedUP() bool {
	return s.GCPBuild.Status.CleanedUP
}

// Implement the method to return the Compute service
func (b *BuildScope) GetComputeService() *compute.Service {
	return b.GCPServices.Compute
}
