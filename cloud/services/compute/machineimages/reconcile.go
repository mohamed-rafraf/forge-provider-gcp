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
	"context"
	"fmt"

	"google.golang.org/api/compute/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/forge-build/forge-provider-gcp/cloud/gcperrors"
)

// Reconcile ensures that an image is created from an instance.
func (s *Service) Reconcile(ctx context.Context) error {
	logger := log.FromContext(ctx)

	if !s.scope.IsProvisionerReady() {
		logger.Info("Not ready for building the image")
		return nil
	}

	logger.Info("Reconciling image creation")

	// Get project, zone, and image name from scope
	imageName := s.scope.ImageName()

	// First, check if the image already exists
	imageExists, err := s.checkIfImageExists(ctx, imageName)
	if err != nil {
		return err
	}

	// If the image doesn't exist, create it
	if !imageExists {
		// Create image from the instance
		if err := s.createImageFromInstance(ctx, s.scope.Name(), imageName); err != nil {
			return err
		}
	}

	// Optionally log the successful reconciliation
	logger.Info("Image reconciliation successful", "image", imageName)
	s.scope.SetBuildReady()

	if !s.scope.IsBuildReady() || s.scope.IsReady() {
		logger.Info("Not ready for checking Image")
		return nil
	}

	logger.Info("Reconciling image validation")
	// Get the MachineImagesGetCall object
	getCall := s.images.Get(s.scope.Project(), imageName)

	// Call Do() to execute the request
	image, err := getCall.Do()
	if err != nil {
		if gcperrors.IsNotFound(err) {
			// Image doesn't exist
			return err
		}
		// Handle other errors (network, permissions)
		logger.Error(err, "Error checking image existence", "image", imageName)
	}

	if image.Status == "READY" {
		artificatRef := fmt.Sprintf("projects/%s/zones/%s/machineImages/%s", s.scope.Project(), s.scope.Zone(), imageName)
		s.scope.SetArtifactRef(artificatRef)
	}
	return nil
}

// checkIfImageExists checks if an image with the specified name already exists.
func (s *Service) checkIfImageExists(ctx context.Context, imageName string) (bool, error) {
	logger := log.FromContext(ctx)
	logger.V(2).Info("Checking if image exists", "image", imageName)

	// Get the MachineImagesGetCall object
	getCall := s.images.Get(s.scope.Project(), imageName)

	// Call Do() to execute the request
	_, err := getCall.Do()
	if err != nil {
		if gcperrors.IsNotFound(err) {
			// Image doesn't exist
			return false, nil
		}
		// Handle other errors (network, permissions)
		logger.Error(err, "Error checking image existence", "image", imageName)
		return false, err
	}

	// If the Do() call succeeds, the image exists
	logger.V(2).Info("Image exists", "image", imageName)
	return true, nil
}

// createImageFromInstance creates a GCP image from an existing instance.
func (s *Service) createImageFromInstance(ctx context.Context, instanceName, imageName string) error {
	logger := log.FromContext(ctx)
	logger.Info("Creating image from instance", "instance", instanceName, "image", imageName)

	// Define image creation parameters
	image := &compute.MachineImage{
		Name:           imageName,
		SourceInstance: fmt.Sprintf("projects/%s/zones/%s/instances/%s", s.scope.Project(), s.scope.Zone(), instanceName),
		Description:    fmt.Sprintf("Custom image created from instance: %s", instanceName),
	}

	// Create the MachineImagesInsertCall
	insertCall := s.images.Insert(s.scope.Project(), image)

	// Call Do() to execute the request
	op, err := insertCall.Do()
	if err != nil {
		return fmt.Errorf("failed to create image: %v", err)
	}
	// we can check the operation's status here to ensure the image creation completes before returning.
	if op.Error != nil {
		return fmt.Errorf("image creation failed: %v", op.Error)
	}

	// Log success and return
	logger.Info("Image creation initiated", "image", imageName, "operation", op.Name)
	return nil
}

func (s *Service) Delete(ctx context.Context) error {
	return nil
}
