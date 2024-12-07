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

package images

import (
	"context"

	k8scloud "github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/filter"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/forge-build/forge-provider-gcp/pkg/cloud"
	"google.golang.org/api/compute/v1"
)

type instanceInterface interface {
	Delete(project string, zone string, instance string) *compute.InstancesDeleteCall
	Get(project string, zone string, instance string) *compute.InstancesGetCall
	Insert(project string, zone string, instance *compute.Instance) *compute.InstancesInsertCall
	List(project string, zone string) *compute.InstancesListCall
	Start(project string, zone string, instance string) *compute.InstancesStartCall
	Stop(project string, zone string, instance string) *compute.InstancesStopCall
}

type imageInterface interface {
	Get(ctx context.Context, key *meta.Key, options ...k8scloud.Option) (*compute.Image, error)
	List(ctx context.Context, fl *filter.F, options ...k8scloud.Option) ([]*compute.Image, error)
	Insert(ctx context.Context, key *meta.Key, obj *compute.Image, options ...k8scloud.Option) error
	Delete(ctx context.Context, key *meta.Key, options ...k8scloud.Option) error
}

// Scope is an interface that holds methods used for reconciling images.
type Scope interface {
	cloud.Build
	InstanceImageSpec() *compute.AttachedDisk
	IsProvisionerReady() bool
	ImageName() string
	Name() string
	IsReady() bool
	GetComputeService() *compute.Service
	SetArtifactRef(artificatRef string)
}

// Service implements the reconcile logic for managing images in GCP.
type Service struct {
	scope    Scope
	instance instanceInterface
	images   imageInterface
}

// New returns a new instance of Service for image creation.
func New(scope Scope) *Service {
	return &Service{
		scope:    scope,
		instance: compute.NewInstancesService(scope.GetComputeService()),
		images:   scope.Cloud().Images(),
	}
}

var _ cloud.Reconciler = &Service{}
