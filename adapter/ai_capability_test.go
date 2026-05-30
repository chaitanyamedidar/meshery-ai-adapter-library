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

package adapter

import (
	"encoding/json"
	"testing"
)

func TestNewReadOnlyAICapabilityMetadata(t *testing.T) {
	metadata := NewReadOnlyAICapabilityMetadata()
	if err := metadata.Validate(); err != nil {
		t.Fatalf("expected valid metadata, got %v", err)
	}
	if len(metadata.SupportedOperations) != 1 || metadata.SupportedOperations[0] != AIAssistantOperation {
		t.Fatalf("expected assistant operation, got %v", metadata.SupportedOperations)
	}
	if len(metadata.ResponseModes) != 4 {
		t.Fatalf("expected four response modes, got %v", metadata.ResponseModes)
	}
}

func TestAICapabilityMetadataToProperties(t *testing.T) {
	metadata := NewReadOnlyAICapabilityMetadata()
	metadata.PrivacyModes = []AIPrivacyMode{AIPrivacyModeLocal}
	metadata.ContextLimits = AIContextLimits{
		MaxRequestBytes: 1024,
		MaxInputTokens:  512,
	}
	metadata.Providers = []AIProviderCapability{{
		Name:                  "ollama",
		Type:                  "local",
		PrivacyMode:           AIPrivacyModeLocal,
		SchemaEnforcementMode: AISchemaEnforcementPromptOnly,
		Models:                []string{"llama3.1"},
	}}

	properties, err := metadata.ToProperties()
	if err != nil {
		t.Fatalf("expected properties, got %v", err)
	}

	raw := properties[AICapabilityMetadataProperty]
	if raw == "" {
		t.Fatal("expected capability metadata property")
	}

	var got AICapabilityMetadata
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("expected JSON metadata, got %v", err)
	}
	if got.Providers[0].Name != "ollama" {
		t.Fatalf("expected provider metadata, got %v", got.Providers)
	}
}

func TestAICapabilityMetadataValidate(t *testing.T) {
	tests := []struct {
		name     string
		metadata AICapabilityMetadata
	}{
		{
			name:     "missing supported operations",
			metadata: AICapabilityMetadata{ResponseModes: []AIAssistantResponseMode{AIAssistantResponseModeExplanation}},
		},
		{
			name:     "missing response modes",
			metadata: AICapabilityMetadata{SupportedOperations: []string{AIAssistantOperation}},
		},
		{
			name: "empty supported operation",
			metadata: AICapabilityMetadata{
				SupportedOperations: []string{""},
				ResponseModes:       []AIAssistantResponseMode{AIAssistantResponseModeExplanation},
			},
		},
		{
			name: "invalid response mode",
			metadata: AICapabilityMetadata{
				SupportedOperations: []string{AIAssistantOperation},
				ResponseModes:       []AIAssistantResponseMode{"mutation"},
			},
		},
		{
			name: "invalid privacy mode",
			metadata: AICapabilityMetadata{
				SupportedOperations: []string{AIAssistantOperation},
				ResponseModes:       []AIAssistantResponseMode{AIAssistantResponseModeExplanation},
				PrivacyModes:        []AIPrivacyMode{"hybrid"},
			},
		},
		{
			name: "negative context limit",
			metadata: AICapabilityMetadata{
				SupportedOperations: []string{AIAssistantOperation},
				ResponseModes:       []AIAssistantResponseMode{AIAssistantResponseModeExplanation},
				ContextLimits:       AIContextLimits{MaxRequestBytes: -1},
			},
		},
		{
			name: "provider without name",
			metadata: AICapabilityMetadata{
				SupportedOperations: []string{AIAssistantOperation},
				ResponseModes:       []AIAssistantResponseMode{AIAssistantResponseModeExplanation},
				Providers:           []AIProviderCapability{{Type: "local"}},
			},
		},
		{
			name: "invalid provider schema mode",
			metadata: AICapabilityMetadata{
				SupportedOperations: []string{AIAssistantOperation},
				ResponseModes:       []AIAssistantResponseMode{AIAssistantResponseModeExplanation},
				Providers: []AIProviderCapability{{
					Name:                  "ollama",
					SchemaEnforcementMode: "unknown",
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.metadata.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
