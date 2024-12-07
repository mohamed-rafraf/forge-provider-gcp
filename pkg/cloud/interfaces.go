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

package cloud

import (
	"context"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	infrav1 "github.com/forge-build/forge-provider-gcp/pkg/api/v1alpha1"
)

// Cloud alias for cloud.Cloud interface.
type Cloud = cloud.Cloud

// Reconciler is a generic interface used by components offering a type of service.
type Reconciler interface {
	Reconcile(ctx context.Context) error
	Delete(ctx context.Context) error
}

// Client is an interface which can get cloud client.
type Client interface {
	Cloud() Cloud
	NetworkCloud() Cloud
}

// BuildGetter is an interface which can get build information.
type BuildGetter interface {
	Client
	Project() string
	Region() string
	Name() string
	Namespace() string
	Zone() string
	NetworkName() string
	NetworkProject() string
	IsSharedVpc() bool
	Network() *infrav1.Network
	AdditionalLabels() infrav1.Labels
	GetBootstrapData() (string, error)
}

// BuildSetter is an interface which can set cluster information.
type BuildSetter interface {
	SetInstanceID(instanceID string)
	SetInstanceStatus(v infrav1.InstanceStatus)
	EnsureCredentialsSecret(ctx context.Context, host string) error
}

// Build is an interface which can get and set build information.
type Build interface {
	BuildGetter
	BuildSetter
}
