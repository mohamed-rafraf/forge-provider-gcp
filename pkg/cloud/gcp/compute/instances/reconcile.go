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
package instances

import (
	"context"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/cluster-api-provider-gcp/cloud/gcperrors"
	"sigs.k8s.io/controller-runtime/pkg/log"

	infrav1 "github.com/forge-build/forge-provider-gcp/pkg/api/v1alpha1"
)

// Reconcile reconcile machine instance.
func (s *Service) Reconcile(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling instance resources")
	instance, err := s.createOrGetInstance(ctx)
	if err != nil {
		return err
	}

	//machineName := s.scope.Name()
	//zone := s.scope.Zone()
	//project := s.scope.Project()

	var publicIP string
	for _, iface := range instance.NetworkInterfaces {
		for _, ac := range iface.AccessConfigs {
			publicIP = ac.NatIP
			break
		}
	}

	//TODO things, after the machine is created
	err = s.scope.EnsureCredentialsSecret(ctx, publicIP)
	if err != nil {
		return err
	}
	s.scope.SetInstanceID(instance.Name)
	s.scope.SetInstanceStatus(infrav1.InstanceStatus(instance.Status))

	return nil
}

// Delete delete machine instance.
func (s *Service) Delete(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Info("Deleting instance resources")
	instanceSpec := s.scope.InstanceSpec(logger)
	instanceName := instanceSpec.Name
	instanceKey := meta.ZonalKey(instanceName, s.scope.Zone())
	logger.V(2).Info("Looking for instance before deleting", "name", instanceName, "zone", s.scope.Zone())
	_, err := s.instances.Get(ctx, instanceKey)
	if err != nil {
		if !gcperrors.IsNotFound(err) {
			logger.Error(err, "Error looking for instance before deleting", "name", instanceName)
			return err
		}

		return nil
	}

	logger.V(2).Info("Deleting instance", "name", instanceName, "zone", s.scope.Zone())
	return gcperrors.IgnoreNotFound(s.instances.Delete(ctx, instanceKey))
}

func (s *Service) createOrGetInstance(ctx context.Context) (*compute.Instance, error) {
	logger := log.FromContext(ctx)
	logger.V(2).Info("Getting bootstrap data for machine")
	bootstrapData, err := s.scope.GetBootstrapData()
	if err != nil {
		logger.Error(err, "Error getting bootstrap data for machine")
		return nil, errors.Wrap(err, "failed to retrieve bootstrap data")
	}

	instanceSpec := s.scope.InstanceSpec(logger)
	instanceName := instanceSpec.Name
	instanceKey := meta.ZonalKey(instanceName, s.scope.Zone())
	if bootstrapData != "" {
		instanceSpec.Metadata.Items = append(instanceSpec.Metadata.Items, &compute.MetadataItems{
			Key:   "user-data",
			Value: ptr.To[string](bootstrapData),
		})
	}

	logger.V(2).Info("Looking for instance", "name", instanceName, "zone", s.scope.Zone())
	instance, err := s.instances.Get(ctx, instanceKey)
	if err != nil {
		if !gcperrors.IsNotFound(err) {
			logger.Error(err, "Error looking for instance", "name", instanceName, "zone", s.scope.Zone())
			return nil, err
		}

		logger.V(2).Info("Creating an instance", "name", instanceName, "zone", s.scope.Zone())
		if err := s.instances.Insert(ctx, instanceKey, instanceSpec); err != nil {
			logger.Error(err, "Error creating an instance", "name", instanceName, "zone", s.scope.Zone())
			return nil, err
		}

		instance, err = s.instances.Get(ctx, instanceKey)
		if err != nil {
			return nil, err
		}
	}

	return instance, nil
}
