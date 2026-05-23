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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/layer5io/meshery-adapter-library/meshes"
)

const (
	// AIAssistantOperation is the shared operation key for read-only assistant requests.
	AIAssistantOperation = "ai-assistant-query"
)

// AIAssistantHandler is implemented by adapters that support read-only AI assistant queries.
type AIAssistantHandler interface {
	QueryAssistant(context.Context, AIAssistantRequest) (AIAssistantResponse, error)
}

// AIAssistantRequest is the provider-neutral request contract for read-only assistant queries.
type AIAssistantRequest struct {
	UserIntent              string          `json:"user_intent"`
	ProviderConnectionRef   string          `json:"provider_connection_ref,omitempty"`
	CurrentDesignRef        string          `json:"current_design_ref,omitempty"`
	CurrentDesignContext    json.RawMessage `json:"current_design_context,omitempty"`
	SelectedComponentIDs    []string        `json:"selected_component_ids,omitempty"`
	SchemaScope             []string        `json:"schema_scope,omitempty"`
	RequestID               string          `json:"request_id,omitempty"`
	AdditionalContextFields json.RawMessage `json:"additional_context_fields,omitempty"`
}

// AIAssistantResponse is the provider-neutral response contract for read-only assistant queries.
type AIAssistantResponse struct {
	Explanation     string                      `json:"explanation,omitempty"`
	Recommendations []AIAssistantRecommendation `json:"recommendations,omitempty"`
	Redirects       []AIAssistantRedirect       `json:"redirects,omitempty"`
	Errors          []AIAssistantError          `json:"errors,omitempty"`
	RequestID       string                      `json:"request_id,omitempty"`
	Metadata        map[string]string           `json:"metadata,omitempty"`
}

// AIAssistantRecommendation describes a read-only recommendation for the user.
type AIAssistantRecommendation struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// AIAssistantRedirect points the user to a Meshery resource or documentation page.
type AIAssistantRedirect struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type,omitempty"`
}

// AIAssistantError is a structured assistant error safe to return to Meshery Server.
type AIAssistantError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// NewAIAssistantOperation returns the standard custom operation for AI-capable adapters.
func NewAIAssistantOperation() *Operation {
	return &Operation{
		Type:        int32(meshes.OpCategory_CUSTOM),
		Description: "Read-only AI assistant query",
		Versions:    NoneVersion,
		Templates:   NoneTemplate,
	}
}

// Validate checks that the assistant request has the minimum required input.
func (r AIAssistantRequest) Validate() error {
	if strings.TrimSpace(r.UserIntent) == "" {
		return errors.New("user_intent is required")
	}
	if err := validateRawJSON("current_design_context", r.CurrentDesignContext); err != nil {
		return err
	}
	if err := validateRawJSON("additional_context_fields", r.AdditionalContextFields); err != nil {
		return err
	}
	return nil
}

// Validate checks that the assistant response has exactly one useful output class.
func (r AIAssistantResponse) Validate() error {
	outputClasses := 0
	if strings.TrimSpace(r.Explanation) != "" {
		outputClasses++
	}
	if len(r.Recommendations) > 0 {
		outputClasses++
	}
	if len(r.Redirects) > 0 {
		outputClasses++
	}
	if len(r.Errors) > 0 {
		outputClasses++
	}
	if outputClasses == 0 {
		return errors.New("assistant response must include explanation, recommendation, redirect, or structured error")
	}
	if outputClasses > 1 {
		return errors.New("assistant response must include only one output class")
	}
	if len(r.Recommendations) > 0 {
		return validateRecommendations(r.Recommendations)
	}
	if len(r.Redirects) > 0 {
		return validateRedirects(r.Redirects)
	}
	if len(r.Errors) > 0 {
		return validateAssistantErrors(r.Errors)
	}
	return nil
}

func validateRawJSON(field string, raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	if !json.Valid(raw) {
		return fmt.Errorf("%s must contain valid JSON", field)
	}
	return nil
}

func validateRecommendations(recommendations []AIAssistantRecommendation) error {
	for i, recommendation := range recommendations {
		if strings.TrimSpace(recommendation.Title) == "" {
			return fmt.Errorf("recommendations[%d].title is required", i)
		}
	}
	return nil
}

func validateRedirects(redirects []AIAssistantRedirect) error {
	for i, redirect := range redirects {
		if strings.TrimSpace(redirect.Title) == "" {
			return fmt.Errorf("redirects[%d].title is required", i)
		}
		if strings.TrimSpace(redirect.URL) == "" {
			return fmt.Errorf("redirects[%d].url is required", i)
		}
	}
	return nil
}

func validateAssistantErrors(assistantErrors []AIAssistantError) error {
	for i, assistantError := range assistantErrors {
		if strings.TrimSpace(assistantError.Code) == "" {
			return fmt.Errorf("errors[%d].code is required", i)
		}
		if strings.TrimSpace(assistantError.Message) == "" {
			return fmt.Errorf("errors[%d].message is required", i)
		}
	}
	return nil
}
