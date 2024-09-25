//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/api/v1beta1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AttachedDiskSpec) DeepCopyInto(out *AttachedDiskSpec) {
	*out = *in
	if in.DeviceType != nil {
		in, out := &in.DeviceType, &out.DeviceType
		*out = new(DiskType)
		**out = **in
	}
	if in.Size != nil {
		in, out := &in.Size, &out.Size
		*out = new(int64)
		**out = **in
	}
	if in.EncryptionKey != nil {
		in, out := &in.EncryptionKey, &out.EncryptionKey
		*out = new(CustomerEncryptionKey)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AttachedDiskSpec.
func (in *AttachedDiskSpec) DeepCopy() *AttachedDiskSpec {
	if in == nil {
		return nil
	}
	out := new(AttachedDiskSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BuildParams) DeepCopyInto(out *BuildParams) {
	*out = *in
	if in.Additional != nil {
		in, out := &in.Additional, &out.Additional
		*out = make(Labels, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BuildParams.
func (in *BuildParams) DeepCopy() *BuildParams {
	if in == nil {
		return nil
	}
	out := new(BuildParams)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomerEncryptionKey) DeepCopyInto(out *CustomerEncryptionKey) {
	*out = *in
	if in.KMSKeyServiceAccount != nil {
		in, out := &in.KMSKeyServiceAccount, &out.KMSKeyServiceAccount
		*out = new(string)
		**out = **in
	}
	if in.ManagedKey != nil {
		in, out := &in.ManagedKey, &out.ManagedKey
		*out = new(ManagedKey)
		**out = **in
	}
	if in.SuppliedKey != nil {
		in, out := &in.SuppliedKey, &out.SuppliedKey
		*out = new(SuppliedKey)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomerEncryptionKey.
func (in *CustomerEncryptionKey) DeepCopy() *CustomerEncryptionKey {
	if in == nil {
		return nil
	}
	out := new(CustomerEncryptionKey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GCPBuild) DeepCopyInto(out *GCPBuild) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GCPBuild.
func (in *GCPBuild) DeepCopy() *GCPBuild {
	if in == nil {
		return nil
	}
	out := new(GCPBuild)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GCPBuild) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GCPBuildList) DeepCopyInto(out *GCPBuildList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GCPBuild, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GCPBuildList.
func (in *GCPBuildList) DeepCopy() *GCPBuildList {
	if in == nil {
		return nil
	}
	out := new(GCPBuildList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GCPBuildList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GCPBuildSpec) DeepCopyInto(out *GCPBuildSpec) {
	*out = *in
	in.Network.DeepCopyInto(&out.Network)
	if in.FailureDomains != nil {
		in, out := &in.FailureDomains, &out.FailureDomains
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Subnet != nil {
		in, out := &in.Subnet, &out.Subnet
		*out = new(string)
		**out = **in
	}
	if in.InstanceID != nil {
		in, out := &in.InstanceID, &out.InstanceID
		*out = new(string)
		**out = **in
	}
	in.Bootstrap.DeepCopyInto(&out.Bootstrap)
	if in.ImageFamily != nil {
		in, out := &in.ImageFamily, &out.ImageFamily
		*out = new(string)
		**out = **in
	}
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(string)
		**out = **in
	}
	if in.AdditionalLabels != nil {
		in, out := &in.AdditionalLabels, &out.AdditionalLabels
		*out = make(Labels, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.AdditionalMetadata != nil {
		in, out := &in.AdditionalMetadata, &out.AdditionalMetadata
		*out = make([]MetadataItem, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.PublicIP != nil {
		in, out := &in.PublicIP, &out.PublicIP
		*out = new(bool)
		**out = **in
	}
	if in.AdditionalNetworkTags != nil {
		in, out := &in.AdditionalNetworkTags, &out.AdditionalNetworkTags
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.RootDeviceType != nil {
		in, out := &in.RootDeviceType, &out.RootDeviceType
		*out = new(DiskType)
		**out = **in
	}
	if in.AdditionalDisks != nil {
		in, out := &in.AdditionalDisks, &out.AdditionalDisks
		*out = make([]AttachedDiskSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ServiceAccount != nil {
		in, out := &in.ServiceAccount, &out.ServiceAccount
		*out = new(ServiceAccount)
		(*in).DeepCopyInto(*out)
	}
	if in.CredentialsRef != nil {
		in, out := &in.CredentialsRef, &out.CredentialsRef
		*out = new(v1.SecretReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GCPBuildSpec.
func (in *GCPBuildSpec) DeepCopy() *GCPBuildSpec {
	if in == nil {
		return nil
	}
	out := new(GCPBuildSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GCPBuildStatus) DeepCopyInto(out *GCPBuildStatus) {
	*out = *in
	in.Network.DeepCopyInto(&out.Network)
	if in.InstanceStatus != nil {
		in, out := &in.InstanceStatus, &out.InstanceStatus
		*out = new(InstanceStatus)
		**out = **in
	}
	if in.ArtifactRef != nil {
		in, out := &in.ArtifactRef, &out.ArtifactRef
		*out = new(string)
		**out = **in
	}
	if in.FailureReason != nil {
		in, out := &in.FailureReason, &out.FailureReason
		*out = new(string)
		**out = **in
	}
	if in.FailureMessage != nil {
		in, out := &in.FailureMessage, &out.FailureMessage
		*out = new(string)
		**out = **in
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make(v1beta1.Conditions, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GCPBuildStatus.
func (in *GCPBuildStatus) DeepCopy() *GCPBuildStatus {
	if in == nil {
		return nil
	}
	out := new(GCPBuildStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Labels) DeepCopyInto(out *Labels) {
	{
		in := &in
		*out = make(Labels, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Labels.
func (in Labels) DeepCopy() Labels {
	if in == nil {
		return nil
	}
	out := new(Labels)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadBalancer) DeepCopyInto(out *LoadBalancer) {
	*out = *in
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.Subnet != nil {
		in, out := &in.Subnet, &out.Subnet
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadBalancer.
func (in *LoadBalancer) DeepCopy() *LoadBalancer {
	if in == nil {
		return nil
	}
	out := new(LoadBalancer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadBalancerSpec) DeepCopyInto(out *LoadBalancerSpec) {
	*out = *in
	if in.APIServerInstanceGroupTagOverride != nil {
		in, out := &in.APIServerInstanceGroupTagOverride, &out.APIServerInstanceGroupTagOverride
		*out = new(string)
		**out = **in
	}
	if in.LoadBalancerType != nil {
		in, out := &in.LoadBalancerType, &out.LoadBalancerType
		*out = new(LoadBalancerType)
		**out = **in
	}
	if in.InternalLoadBalancer != nil {
		in, out := &in.InternalLoadBalancer, &out.InternalLoadBalancer
		*out = new(LoadBalancer)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadBalancerSpec.
func (in *LoadBalancerSpec) DeepCopy() *LoadBalancerSpec {
	if in == nil {
		return nil
	}
	out := new(LoadBalancerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManagedKey) DeepCopyInto(out *ManagedKey) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManagedKey.
func (in *ManagedKey) DeepCopy() *ManagedKey {
	if in == nil {
		return nil
	}
	out := new(ManagedKey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetadataItem) DeepCopyInto(out *MetadataItem) {
	*out = *in
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetadataItem.
func (in *MetadataItem) DeepCopy() *MetadataItem {
	if in == nil {
		return nil
	}
	out := new(MetadataItem)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Network) DeepCopyInto(out *Network) {
	*out = *in
	if in.SelfLink != nil {
		in, out := &in.SelfLink, &out.SelfLink
		*out = new(string)
		**out = **in
	}
	if in.FirewallRules != nil {
		in, out := &in.FirewallRules, &out.FirewallRules
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Router != nil {
		in, out := &in.Router, &out.Router
		*out = new(string)
		**out = **in
	}
	if in.APIServerAddress != nil {
		in, out := &in.APIServerAddress, &out.APIServerAddress
		*out = new(string)
		**out = **in
	}
	if in.APIServerHealthCheck != nil {
		in, out := &in.APIServerHealthCheck, &out.APIServerHealthCheck
		*out = new(string)
		**out = **in
	}
	if in.APIServerInstanceGroups != nil {
		in, out := &in.APIServerInstanceGroups, &out.APIServerInstanceGroups
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.APIServerBackendService != nil {
		in, out := &in.APIServerBackendService, &out.APIServerBackendService
		*out = new(string)
		**out = **in
	}
	if in.APIServerTargetProxy != nil {
		in, out := &in.APIServerTargetProxy, &out.APIServerTargetProxy
		*out = new(string)
		**out = **in
	}
	if in.APIServerForwardingRule != nil {
		in, out := &in.APIServerForwardingRule, &out.APIServerForwardingRule
		*out = new(string)
		**out = **in
	}
	if in.APIInternalAddress != nil {
		in, out := &in.APIInternalAddress, &out.APIInternalAddress
		*out = new(string)
		**out = **in
	}
	if in.APIInternalHealthCheck != nil {
		in, out := &in.APIInternalHealthCheck, &out.APIInternalHealthCheck
		*out = new(string)
		**out = **in
	}
	if in.APIInternalBackendService != nil {
		in, out := &in.APIInternalBackendService, &out.APIInternalBackendService
		*out = new(string)
		**out = **in
	}
	if in.APIInternalForwardingRule != nil {
		in, out := &in.APIInternalForwardingRule, &out.APIInternalForwardingRule
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Network.
func (in *Network) DeepCopy() *Network {
	if in == nil {
		return nil
	}
	out := new(Network)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkSpec) DeepCopyInto(out *NetworkSpec) {
	*out = *in
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.AutoCreateSubnetworks != nil {
		in, out := &in.AutoCreateSubnetworks, &out.AutoCreateSubnetworks
		*out = new(bool)
		**out = **in
	}
	if in.Subnets != nil {
		in, out := &in.Subnets, &out.Subnets
		*out = make(Subnets, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LoadBalancerBackendPort != nil {
		in, out := &in.LoadBalancerBackendPort, &out.LoadBalancerBackendPort
		*out = new(int32)
		**out = **in
	}
	if in.HostProject != nil {
		in, out := &in.HostProject, &out.HostProject
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkSpec.
func (in *NetworkSpec) DeepCopy() *NetworkSpec {
	if in == nil {
		return nil
	}
	out := new(NetworkSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceAccount) DeepCopyInto(out *ServiceAccount) {
	*out = *in
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceAccount.
func (in *ServiceAccount) DeepCopy() *ServiceAccount {
	if in == nil {
		return nil
	}
	out := new(ServiceAccount)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubnetSpec) DeepCopyInto(out *SubnetSpec) {
	*out = *in
	if in.Description != nil {
		in, out := &in.Description, &out.Description
		*out = new(string)
		**out = **in
	}
	if in.SecondaryCidrBlocks != nil {
		in, out := &in.SecondaryCidrBlocks, &out.SecondaryCidrBlocks
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.PrivateGoogleAccess != nil {
		in, out := &in.PrivateGoogleAccess, &out.PrivateGoogleAccess
		*out = new(bool)
		**out = **in
	}
	if in.EnableFlowLogs != nil {
		in, out := &in.EnableFlowLogs, &out.EnableFlowLogs
		*out = new(bool)
		**out = **in
	}
	if in.Purpose != nil {
		in, out := &in.Purpose, &out.Purpose
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubnetSpec.
func (in *SubnetSpec) DeepCopy() *SubnetSpec {
	if in == nil {
		return nil
	}
	out := new(SubnetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Subnets) DeepCopyInto(out *Subnets) {
	{
		in := &in
		*out = make(Subnets, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Subnets.
func (in Subnets) DeepCopy() Subnets {
	if in == nil {
		return nil
	}
	out := new(Subnets)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SuppliedKey) DeepCopyInto(out *SuppliedKey) {
	*out = *in
	if in.RawKey != nil {
		in, out := &in.RawKey, &out.RawKey
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.RSAEncryptedKey != nil {
		in, out := &in.RSAEncryptedKey, &out.RSAEncryptedKey
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SuppliedKey.
func (in *SuppliedKey) DeepCopy() *SuppliedKey {
	if in == nil {
		return nil
	}
	out := new(SuppliedKey)
	in.DeepCopyInto(out)
	return out
}