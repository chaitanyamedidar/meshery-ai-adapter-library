// Copyright Meshery Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc

import (
	"context"
	"testing"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/meshes"
)

type componentInfoHandler struct{}

func (componentInfoHandler) GetName() string {
	return "AI"
}

func (componentInfoHandler) GetComponentInfo(svc interface{}) error {
	service := svc.(*Service)
	service.Name = "AI"
	service.Type = "adapter"
	service.Version = "edge"
	service.GitSHA = "abc123"
	service.Properties = map[string]string{
		adapter.AICapabilityMetadataProperty: "{}",
	}
	return nil
}

func (componentInfoHandler) ApplyOperation(context.Context, adapter.OperationRequest) error {
	return nil
}

func (componentInfoHandler) ListOperations() (adapter.Operations, error) {
	return nil, nil
}

func (componentInfoHandler) ProcessOAM(context.Context, adapter.OAMRequest) (string, error) {
	return "", nil
}

func (componentInfoHandler) StreamErr(*meshes.EventsResponse, error) {}

func (componentInfoHandler) StreamInfo(*meshes.EventsResponse) {}

func TestComponentInfoReturnsProperties(t *testing.T) {
	service := &Service{Handler: componentInfoHandler{}}

	resp, err := service.ComponentInfo(context.Background(), &meshes.ComponentInfoRequest{})
	if err != nil {
		t.Fatalf("expected component info, got %v", err)
	}
	if resp.Properties[adapter.AICapabilityMetadataProperty] != "{}" {
		t.Fatalf("expected properties in component info, got %v", resp.Properties)
	}
}
