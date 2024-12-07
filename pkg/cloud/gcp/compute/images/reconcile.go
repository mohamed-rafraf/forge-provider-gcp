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
	"fmt"
	"time"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	"google.golang.org/api/compute/v1"
	"sigs.k8s.io/cluster-api-provider-gcp/cloud/gcperrors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Reconcile ensures that a disk image is created from an instance.
func (s *Service) Reconcile(ctx context.Context) error {
	logger := log.FromContext(ctx)

	if !s.scope.IsProvisionerReady() || s.scope.IsReady() {
		logger.Info("Not ready for exporting the image")
		return nil
	}

	logger.Info("Reconciling image creation")

	imageName := s.scope.ImageName()
	instanceName := s.scope.Name()

	// Stop the instance
	if err := s.stopInstance(ctx, instanceName); err != nil {
		return err
	}

	// Wait for the instance to be stopped.
	instance, err := s.instance.Get(s.scope.Project(), s.scope.Zone(), instanceName).Do()
	if err != nil {
		return fmt.Errorf("failed to get instance status: %v", err)
	}
	if instance.Status != "TERMINATED" {
		logger.Info("The instance is not stopped yet", "status", instance.Status)
		return nil
	}
	// Delete the existing disk image if it exists
	logger.Info("Ensuring no existing image conflicts", "image", imageName)
	if err := s.ensureImageDoesNotExist(ctx, imageName); err != nil {
		return err
	}

	// Create the disk image from the instance's boot disk
	logger.Info("Creating disk image from instance", "instance", instanceName, "image", imageName)
	if err := s.createDiskImage(ctx, instanceName, imageName); err != nil {
		return err
	}

	// Wait for the disk image to be ready
	logger.Info("Waiting for image readiness", "image", imageName)
	key := &meta.Key{Name: imageName}
	image, err := s.images.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get image status: %v", err)
	}

	if image.Status != "READY" {
		logger.Info("Disk image is not ready yet", "image", imageName)
		return nil
	}

	artifactRef := fmt.Sprintf("projects/%s/global/images/%s", s.scope.Project(), imageName)
	s.scope.SetArtifactRef(artifactRef)
	logger.Info("Disk image reconciliation successful", "image", imageName)
	return nil
}

// stopInstance stops the specified instance and waits for it to stop.
func (s *Service) stopInstance(ctx context.Context, instanceName string) error {
	logger := log.FromContext(ctx)
	logger.Info("Stopping instance", "instance", instanceName)

	stopCall := s.instance.Stop(s.scope.Project(), s.scope.Zone(), instanceName)
	_, err := stopCall.Do()
	if err != nil {
		return fmt.Errorf("failed to stop instance: %v", err)
	}
	return nil
}

// ensureImageDoesNotExist deletes the disk image if it already exists.
func (s *Service) ensureImageDoesNotExist(ctx context.Context, imageName string) error {
	logger := log.FromContext(ctx)
	key := &meta.Key{Name: imageName}

	err := s.images.Delete(ctx, key)
	if err != nil {
		if gcperrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete existing image: %v", err)
	}

	logger.Info("Waiting for existing image deletion", "image", imageName)
	time.Sleep(5 * time.Second)
	return nil
}

// createDiskImage creates a new disk image from the instance's boot disk.
func (s *Service) createDiskImage(ctx context.Context, instanceName, imageName string) error {
	logger := log.FromContext(ctx)
	image := &compute.Image{
		Name: imageName,
		SourceDisk: fmt.Sprintf("projects/%s/zones/%s/disks/%s",
			s.scope.Project(), s.scope.Zone(), instanceName),
		Description: fmt.Sprintf("Custom disk image created from instance: %s", instanceName),
	}

	key := &meta.Key{Name: imageName}
	err := s.images.Insert(ctx, key, image)
	if err != nil {
		return fmt.Errorf("failed to create disk image: %v", err)
	}

	logger.Info("Disk image creation initiated", "image", imageName)
	return nil
}

func (s *Service) Delete(ctx context.Context) error {
	return nil
}
