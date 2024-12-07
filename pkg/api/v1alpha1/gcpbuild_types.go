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

package v1alpha1

import (
	buildv1 "github.com/forge-build/forge/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// BuildFinalizer allows ReconcileGCPBuild to clean up GCP resources associated with GCPBuild before
	// removing it from the apiserver.
	BuildFinalizer = "gcpbuild.infrastructure.forge.build"

	// GCPBuildKind the kind of a GCPBuild Object.
	GCPBuildKind string = "GCPBuild"
)

// DiskType is a type to use to define with disk type will be used.
type DiskType string

const (
	// PdStandardDiskType defines the name for the standard disk.
	PdStandardDiskType DiskType = "pd-standard"
	// PdSsdDiskType defines the name for the ssd disk.
	PdSsdDiskType DiskType = "pd-ssd"
	// LocalSsdDiskType defines the name for the local ssd disk.
	LocalSsdDiskType DiskType = "local-ssd"
)

// AttachedDiskSpec degined GCP machine disk.
type AttachedDiskSpec struct {
	// DeviceType is a device type of the attached disk.
	// Supported types of non-root attached volumes:
	// 1. "pd-standard" - Standard (HDD) persistent disk
	// 2. "pd-ssd" - SSD persistent disk
	// 3. "local-ssd" - Local SSD disk (https://cloud.google.com/compute/docs/disks/local-ssd).
	// 4. "pd-balanced" - Balanced Persistent Disk
	// 5. "hyperdisk-balanced" - Hyperdisk Balanced
	// Default is "pd-standard".
	// +optional
	DeviceType *DiskType `json:"deviceType,omitempty"`
	// Size is the size of the disk in GBs.
	// Defaults to 30GB. For "local-ssd" size is always 375GB.
	// +optional
	Size *int64 `json:"size,omitempty"`
	// EncryptionKey defines the KMS key to be used to encrypt the disk.
	// +optional
	EncryptionKey *CustomerEncryptionKey `json:"encryptionKey,omitempty"`
}

// CustomerEncryptionKey supports both Customer-Managed or Customer-Supplied encryption keys .
type CustomerEncryptionKey struct {
	// KeyType is the type of encryption key. Must be either Managed, aka Customer-Managed Encryption Key (CMEK) or
	// Supplied, aka Customer-Supplied EncryptionKey (CSEK).
	// +kubebuilder:validation:Enum=Managed;Supplied
	KeyType KeyType `json:"keyType"`
	// KMSKeyServiceAccount is the service account being used for the encryption request for the given KMS key.
	// If absent, the Compute Engine default service account is used. For example:
	// "kmsKeyServiceAccount": "name@project_id.iam.gserviceaccount.com.
	// The maximum length is based on the Service Account ID (max 30), Project (max 30), and a valid gcloud email
	// suffix ("iam.gserviceaccount.com").
	// +kubebuilder:validation:MaxLength=85
	// +kubebuilder:validation:Pattern=`[-_[A-Za-z0-9]+@[-_[A-Za-z0-9]+.iam.gserviceaccount.com`
	// +optional
	KMSKeyServiceAccount *string `json:"kmsKeyServiceAccount,omitempty"`
	// ManagedKey references keys managed by the Cloud Key Management Service. This should be set when KeyType is Managed.
	// +optional
	ManagedKey *ManagedKey `json:"managedKey,omitempty"`
	// SuppliedKey provides the key used to create or manage a disk. This should be set when KeyType is Managed.
	// +optional
	SuppliedKey *SuppliedKey `json:"suppliedKey,omitempty"`
}

// KeyType is a type for disk encryption.
type KeyType string

const (
	// CustomerManagedKey (CMEK) references an encryption key stored in Google Cloud KMS.
	CustomerManagedKey KeyType = "Managed"
	// CustomerSuppliedKey (CSEK) specifies an encryption key to use.
	CustomerSuppliedKey KeyType = "Supplied"
)

// ManagedKey is a reference to a key managed by the Cloud Key Management Service.
type ManagedKey struct {
	// KMSKeyName is the name of the encryption key that is stored in Google Cloud KMS. For example:
	// "kmsKeyName": "projects/kms_project_id/locations/region/keyRings/key_region/cryptoKeys/key
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`projects\/[-_[A-Za-z0-9]+\/locations\/[-_[A-Za-z0-9]+\/keyRings\/[-_[A-Za-z0-9]+\/cryptoKeys\/[-_[A-Za-z0-9]+`
	// +kubebuilder:validation:MaxLength=160
	KMSKeyName string `json:"kmsKeyName,omitempty"`
}

// SuppliedKey contains a key for disk encryption. Either RawKey or RSAEncryptedKey must be provided.
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type SuppliedKey struct {
	// RawKey specifies a 256-bit customer-supplied encryption key, encoded in RFC 4648
	// base64 to either encrypt or decrypt this resource. You can provide either the rawKey or the rsaEncryptedKey.
	// For example: "rawKey": "SGVsbG8gZnJvbSBHb29nbGUgQ2xvdWQgUGxhdGZvcm0="
	// +optional
	RawKey []byte `json:"rawKey,omitempty"`
	// RSAEncryptedKey specifies an RFC 4648 base64 encoded, RSA-wrapped 2048-bit customer-supplied encryption
	// key to either encrypt or decrypt this resource. You can provide either the rawKey or the
	// rsaEncryptedKey.
	// For example: "rsaEncryptedKey": "ieCx/NcW06PcT7Ep1X6LUTc/hLvUDYyzSZPPVCVPTVEohpeHASqC8uw5TzyO9U+Fka9JFHi
	// z0mBibXUInrC/jEk014kCK/NPjYgEMOyssZ4ZINPKxlUh2zn1bV+MCaTICrdmuSBTWlUUiFoDi
	// D6PYznLwh8ZNdaheCeZ8ewEXgFQ8V+sDroLaN3Xs3MDTXQEMMoNUXMCZEIpg9Vtp9x2oe=="
	// The key must meet the following requirements before you can provide it to Compute Engine:
	// 1. The key is wrapped using a RSA public key certificate provided by Google.
	// 2. After being wrapped, the key must be encoded in RFC 4648 base64 encoding.
	// Gets the RSA public key certificate provided by Google at: https://cloud-certs.storage.googleapis.com/google-cloud-csek-ingress.pem
	// +optional
	RSAEncryptedKey []byte `json:"rsaEncryptedKey,omitempty"`
}

// GCPBuildSpec defines the desired state of GCPBuild
type GCPBuildSpec struct {
	// Embedded ConnectionSpec to define default connection credentials.
	buildv1.ConnectionSpec `json:",inline"`

	// Project is the name of the project to deploy the cluster to.
	Project string `json:"project"`

	// The GCP Region the cluster lives in.
	Region string `json:"region"`

	// The GCP Region the cluster lives in.
	Zone string `json:"zone"`

	// InstanceType is the type of instance to create. Example: n1.standard-2
	InstanceType string `json:"instanceType"`

	// NetworkSpec encapsulates all things related to GCP network.
	// +optional
	Network NetworkSpec `json:"network"`

	// FailureDomains is an optional field which is used to assign selected availability zones to a cluster
	// FailureDomains if empty, defaults to all the zones in the selected region and if specified would override
	// the default zones.
	// +optional
	FailureDomains []string `json:"failureDomains,omitempty"`

	// Subnet is a reference to the subnetwork to use for this instance. If not specified,
	// the first subnetwork retrieved from the Cluster Region and Network is picked.
	// +optional
	Subnet *string `json:"subnet,omitempty"`

	// InstanceID is the unique identifier as specified by the cloud provider.
	// +optional
	InstanceID *string `json:"InstanceID,omitempty"`

	// Bootstrap is a reference to a local struct which encapsulates
	// fields to configure the Machine’s bootstrapping mechanism.
	// +optional
	Bootstrap clusterv1.Bootstrap `json:"bootstrap,omitempty"`

	// ImageFamily is the full reference to a valid image family to be used for this machine.
	// +optional
	ImageFamily *string `json:"imageFamily,omitempty"`

	// Image is the full reference to a valid image to be used for this machine.
	// Takes precedence over ImageFamily.
	// +optional
	Image *string `json:"image,omitempty"`

	// AdditionalLabels is an optional set of tags to add to an instance, in addition to the ones added by default by the
	// GCP provider. If both the GcpBuild and the GCPMachine specify the same tag name with different values, the
	// GCPMachine's value takes precedence.
	// +optional
	AdditionalLabels Labels `json:"additionalLabels,omitempty"`

	// AdditionalMetadata is an optional set of metadata to add to an instance, in addition to the ones added by default by the
	// GCP provider.
	// +listType=map
	// +listMapKey=key
	// +optional
	AdditionalMetadata []MetadataItem `json:"additionalMetadata,omitempty"`

	// IAMInstanceProfile is a name of an IAM instance profile to assign to the instance
	// +optional
	// IAMInstanceProfile string `json:"iamInstanceProfile,omitempty"`

	// PublicIP specifies whether the instance should get a public IP.
	// Set this to true if you don't have a NAT instances or Cloud Nat setup.
	// +optional
	PublicIP *bool `json:"publicIP,omitempty"`

	// AdditionalNetworkTags is a list of network tags that should be applied to the
	// instance. These tags are set in addition to any network tags defined
	// at the cluster level or in the actuator.
	// +optional
	AdditionalNetworkTags []string `json:"additionalNetworkTags,omitempty"`

	// RootDeviceSize is the size of the root volume in GB.
	// Defaults to 30.
	// +optional
	RootDeviceSize int64 `json:"rootDeviceSize,omitempty"`

	// RootDeviceType is the type of the root volume.
	// Supported types of root volumes:
	// 1. "pd-standard" - Standard (HDD) persistent disk
	// 2. "pd-ssd" - SSD persistent disk
	// 3. "pd-balanced" - Balanced Persistent Disk
	// 4. "hyperdisk-balanced" - Hyperdisk Balanced
	// Default is "pd-standard".
	// +optional
	RootDeviceType *DiskType `json:"rootDeviceType,omitempty"`

	// AdditionalDisks are optional non-boot attached disks.
	// +optional
	AdditionalDisks []AttachedDiskSpec `json:"additionalDisks,omitempty"`

	// ServiceAccount specifies the service account email and which scopes to assign to the machine.
	// Defaults to: email: "default", scope: []{compute.CloudPlatformScope}
	// +optional
	ServiceAccount *ServiceAccount `json:"serviceAccounts,omitempty"`

	// Preemptible defines if instance is preemptible
	// +optional
	Preemptible bool `json:"preemptible,omitempty"`

	// CredentialsRef is a reference to a Secret that contains the credentials to use for provisioning this cluster. If not
	// supplied then the credentials of the controller will be used.
	// +optional
	CredentialsRef *corev1.SecretReference `json:"credentialsRef,omitempty"`
}

// ServiceAccount describes compute.serviceAccount.
type ServiceAccount struct {
	// Email: Email address of the service account.
	Email string `json:"email,omitempty"`

	// Scopes: The list of scopes to be made available for this service
	// account.
	Scopes []string `json:"scopes,omitempty"`
}

// MetadataItem defines a single piece of metadata associated with an instance.
type MetadataItem struct {
	// Key is the identifier for the metadata entry.
	Key string `json:"key"`
	// Value is the value of the metadata entry.
	Value *string `json:"value,omitempty"`
}

// GCPBuildStatus defines the observed state of GCPBuild
type GCPBuildStatus struct {
	// Ready indicates that the GCPBuild is ready.
	// +optional
	// +kubebuilder:default=false
	Ready bool `json:"ready"`

	// MachineReady indicates that the associated machine is ready to accept connection.
	// +optional
	// +kubebuilder:default=false
	MachineReady bool `json:"machineReady"`

	// CleanUpReady indicates that the Infrastructure is cleaned up or not.
	// +optional
	// +kubebuilder:default=false
	CleanedUP bool `json:"cleanedUP,omitempty"`

	// Network status of network.
	Network Network `json:"network,omitempty"`

	// InstanceStatus is the status of the GCP instance for this machine.
	// +optional
	InstanceStatus *InstanceStatus `json:"instanceState,omitempty"`

	// ArtifactRef The Reference of image that has been built.
	// +optional
	ArtifactRef *string `json:"artifactRef,omitempty"`

	// FailureReason will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a succinct value suitable
	// for machine interpretation.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of ProxmoxCluster
	// can be added as events to the ProxmoxCluster object and/or logged in the
	// controller's output.
	// +optional
	FailureReason *string `json:"failureReason,omitempty"`

	// FailureMessage will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a more verbose string suitable
	// for logging and human consumption.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of ProxmoxMachines
	// can be added as events to the ProxmoxCluster object and/or logged in the
	// controller's output.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// Conditions defines current service state of the ProxmoxCluster.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

// Network encapsulates GCP networking resources.
type Network struct {
	// SelfLink is the link to the Network used for this cluster.
	SelfLink *string `json:"selfLink,omitempty"`

	// FirewallRules is a map from the name of the rule to its full reference.
	// +optional
	FirewallRules map[string]string `json:"firewallRules,omitempty"`

	// Router is the full reference to the router created within the network
	// it'll contain the cloud nat gateway
	// +optional
	Router *string `json:"router,omitempty"`

	// APIServerAddress is the IPV4 global address assigned to the load balancer
	// created for the API Server.
	// +optional
	APIServerAddress *string `json:"apiServerIpAddress,omitempty"`

	// APIServerHealthCheck is the full reference to the health check
	// created for the API Server.
	// +optional
	APIServerHealthCheck *string `json:"apiServerHealthCheck,omitempty"`

	// APIServerInstanceGroups is a map from zone to the full reference
	// to the instance groups created for the control plane nodes created in the same zone.
	// +optional
	APIServerInstanceGroups map[string]string `json:"apiServerInstanceGroups,omitempty"`

	// APIServerBackendService is the full reference to the backend service
	// created for the API Server.
	// +optional
	APIServerBackendService *string `json:"apiServerBackendService,omitempty"`

	// APIServerTargetProxy is the full reference to the target proxy
	// created for the API Server.
	// +optional
	APIServerTargetProxy *string `json:"apiServerTargetProxy,omitempty"`

	// APIServerForwardingRule is the full reference to the forwarding rule
	// created for the API Server.
	// +optional
	APIServerForwardingRule *string `json:"apiServerForwardingRule,omitempty"`

	// APIInternalAddress is the IPV4 regional address assigned to the
	// internal Load Balancer.
	// +optional
	APIInternalAddress *string `json:"apiInternalIpAddress,omitempty"`

	// APIInternalHealthCheck is the full reference to the health check
	// created for the internal Load Balancer.
	// +optional
	APIInternalHealthCheck *string `json:"apiInternalHealthCheck,omitempty"`

	// APIInternalBackendService is the full reference to the backend service
	// created for the internal Load Balancer.
	// +optional
	APIInternalBackendService *string `json:"apiInternalBackendService,omitempty"`

	// APIInternalForwardingRule is the full reference to the forwarding rule
	// created for the internal Load Balancer.
	// +optional
	APIInternalForwardingRule *string `json:"apiInternalForwardingRule,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=gcpbuilds,scope=Namespaced,categories=forge;gcp,singular=gcpbuild
// +kubebuilder:printcolumn:name="Build",type="string",JSONPath=".metadata.labels['forge\\.build/build-name']",description="Build"
// +kubebuilder:printcolumn:name="Machine Ready",type="string",JSONPath=".status.machineReady",description="Machine Ready"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Build is ready"

// GCPBuild is the Schema for the gcpbuilds API
type GCPBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPBuildSpec   `json:"spec,omitempty"`
	Status GCPBuildStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GCPBuildList contains a list of GCPBuild
type GCPBuildList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPBuild `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GCPBuild{}, &GCPBuildList{})
}
