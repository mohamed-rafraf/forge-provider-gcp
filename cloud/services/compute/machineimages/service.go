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

package machineimages

import (
	"github.com/forge-build/forge-provider-gcp/cloud"
	"google.golang.org/api/compute/v1"
)

// imagesInterface defines the interface for interacting with GCP images.
type imagesInterface interface {
	Delete(project string, machineImage string) *compute.MachineImagesDeleteCall
	Get(project string, machineImage string) *compute.MachineImagesGetCall
	GetIamPolicy(project string, resource string) *compute.MachineImagesGetIamPolicyCall
	Insert(project string, machineimage *compute.MachineImage) *compute.MachineImagesInsertCall
	List(project string) *compute.MachineImagesListCall
}

// Scope is an interface that holds methods used for reconciling images.
type Scope interface {
	cloud.Build
	InstanceImageSpec() *compute.AttachedDisk
	IsProvisionerReady() bool
	ImageName() string
	Name() string
	SetBuildReady()
	IsBuildReady() bool
	IsReady() bool
	GetComputeService() *compute.Service
	SetArtifactRef(artificatRef string)
}

// Service implements the reconcile logic for managing images in GCP.
type Service struct {
	scope  Scope
	images imagesInterface
}

// New returns a new instance of Service for image creation.
func New(scope Scope) *Service {
	return &Service{
		scope:  scope,
		images: compute.NewMachineImagesService(scope.GetComputeService()),
	}
}

var _ cloud.Reconciler = &Service{}
