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

/*
Package gcpbuild implements a Kubernetes controller for managing GCPBuild custom resources within the Forge project.
This controller is responsible for building machine images in Google Cloud Platform (GCP) as part of Forge's infrastructure provisioning process.
Usage:
- Deploy the controller with Forge's controller manager.
- Define GCPBuild resources to specify image build configurations.
*/
package gcpbuild
