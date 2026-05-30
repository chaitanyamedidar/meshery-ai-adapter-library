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
	"errors"
	"fmt"
	"strings"
)

const (
	// AICapabilityMetadataProperty is the ComponentInfo properties key used to
	// expose AI capability metadata without changing the current gRPC operation
	// discovery response.
	AICapabilityMetadataProperty = "ai.capability_metadata"
)

type AIAssistantResponseMode string

const (
	AIAssistantResponseModeExplanation    AIAssistantResponseMode = "explanation"
	AIAssistantResponseModeRecommendation AIAssistantResponseMode = "recommendation"
	AIAssistantResponseModeRedirect       AIAssistantResponseMode = "redirect"
	AIAssistantResponseModeError          AIAssistantResponseMode = "structured_error"
)

type AIPrivacyMode string

const (
	AIPrivacyModeCloud AIPrivacyMode = "cloud"
	AIPrivacyModeLocal AIPrivacyMode = "local"
)

type AISchemaEnforcementMode string

const (
	AISchemaEnforcementStrictJSONSchema AISchemaEnforcementMode = "strict_json_schema"
	AISchemaEnforcementJSONObject       AISchemaEnforcementMode = "json_object"
	AISchemaEnforcementPromptOnly       AISchemaEnforcementMode = "prompt_only"
	AISchemaEnforcementNone             AISchemaEnforcementMode = "none"
)

// AICapabilityMetadata describes AI adapter discovery metadata that does not
// fit in meshes.SupportedOperation's key/value/category fields.
type AICapabilityMetadata struct {
	SupportedOperations []string                  `json:"supported_operations,omitempty"`
	ResponseModes       []AIAssistantResponseMode `json:"response_modes,omitempty"`
	PrivacyModes        []AIPrivacyMode           `json:"privacy_modes,omitempty"`
	ContextLimits       AIContextLimits           `json:"context_limits,omitempty"`
	Providers           []AIProviderCapability    `json:"providers,omitempty"`
}

type AIContextLimits struct {
	MaxRequestBytes  int `json:"max_request_bytes,omitempty"`
	MaxInputTokens   int `json:"max_input_tokens,omitempty"`
	MaxOutputTokens  int `json:"max_output_tokens,omitempty"`
	ReservedResponse int `json:"reserved_response,omitempty"`
}

type AIProviderCapability struct {
	Name                  string                  `json:"name"`
	Type                  string                  `json:"type,omitempty"`
	PrivacyMode           AIPrivacyMode           `json:"privacy_mode,omitempty"`
	SchemaEnforcementMode AISchemaEnforcementMode `json:"schema_enforcement_mode,omitempty"`
	Models                []string                `json:"models,omitempty"`
}

// NewReadOnlyAICapabilityMetadata returns the standard metadata shape for
// adapters implementing the read-only Kanvas AI Assistant contract.
func NewReadOnlyAICapabilityMetadata() AICapabilityMetadata {
	return AICapabilityMetadata{
		SupportedOperations: []string{AIAssistantOperation},
		ResponseModes: []AIAssistantResponseMode{
			AIAssistantResponseModeExplanation,
			AIAssistantResponseModeRecommendation,
			AIAssistantResponseModeRedirect,
			AIAssistantResponseModeError,
		},
	}
}

// ToProperties serializes AI capability metadata for ComponentInfo.Properties.
func (m AICapabilityMetadata) ToProperties() (map[string]string, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}

	metadata, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		AICapabilityMetadataProperty: string(metadata),
	}, nil
}

func (m AICapabilityMetadata) Validate() error {
	if len(m.SupportedOperations) == 0 {
		return errors.New("supported_operations is required")
	}
	if len(m.ResponseModes) == 0 {
		return errors.New("response_modes is required")
	}
	if err := validateStringList("supported_operations", m.SupportedOperations); err != nil {
		return err
	}
	if err := validateResponseModes(m.ResponseModes); err != nil {
		return err
	}
	if err := validatePrivacyModes(m.PrivacyModes); err != nil {
		return err
	}
	if err := m.ContextLimits.Validate(); err != nil {
		return err
	}
	for i, provider := range m.Providers {
		if err := provider.Validate(); err != nil {
			return fmt.Errorf("providers[%d]: %w", i, err)
		}
	}
	return nil
}

func (l AIContextLimits) Validate() error {
	if l.MaxRequestBytes < 0 {
		return errors.New("context_limits.max_request_bytes cannot be negative")
	}
	if l.MaxInputTokens < 0 {
		return errors.New("context_limits.max_input_tokens cannot be negative")
	}
	if l.MaxOutputTokens < 0 {
		return errors.New("context_limits.max_output_tokens cannot be negative")
	}
	if l.ReservedResponse < 0 {
		return errors.New("context_limits.reserved_response cannot be negative")
	}
	return nil
}

func (p AIProviderCapability) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("name is required")
	}
	if p.PrivacyMode != "" {
		if err := validatePrivacyModes([]AIPrivacyMode{p.PrivacyMode}); err != nil {
			return err
		}
	}
	if p.SchemaEnforcementMode != "" {
		if err := validateSchemaEnforcementModes([]AISchemaEnforcementMode{p.SchemaEnforcementMode}); err != nil {
			return err
		}
	}
	return nil
}

func validateStringList(field string, values []string) error {
	for i, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s[%d] is required", field, i)
		}
	}
	return nil
}

func validateResponseModes(modes []AIAssistantResponseMode) error {
	allowed := map[AIAssistantResponseMode]struct{}{
		AIAssistantResponseModeExplanation:    {},
		AIAssistantResponseModeRecommendation: {},
		AIAssistantResponseModeRedirect:       {},
		AIAssistantResponseModeError:          {},
	}
	for i, mode := range modes {
		if _, ok := allowed[mode]; !ok {
			return fmt.Errorf("response_modes[%d] is invalid", i)
		}
	}
	return nil
}

func validatePrivacyModes(modes []AIPrivacyMode) error {
	allowed := map[AIPrivacyMode]struct{}{
		AIPrivacyModeCloud: {},
		AIPrivacyModeLocal: {},
	}
	for i, mode := range modes {
		if _, ok := allowed[mode]; !ok {
			return fmt.Errorf("privacy_modes[%d] is invalid", i)
		}
	}
	return nil
}

func validateSchemaEnforcementModes(modes []AISchemaEnforcementMode) error {
	allowed := map[AISchemaEnforcementMode]struct{}{
		AISchemaEnforcementStrictJSONSchema: {},
		AISchemaEnforcementJSONObject:       {},
		AISchemaEnforcementPromptOnly:       {},
		AISchemaEnforcementNone:             {},
	}
	for i, mode := range modes {
		if _, ok := allowed[mode]; !ok {
			return fmt.Errorf("schema_enforcement_modes[%d] is invalid", i)
		}
	}
	return nil
}
