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

func TestNewAIAssistantOperation(t *testing.T) {
	op := NewAIAssistantOperation()
	if op == nil {
		t.Fatal("expected operation")
	}
	if op.Description == "" {
		t.Fatal("expected operation description")
	}
	if len(op.AdditionalProperties) != 0 {
		t.Fatalf("expected no additional properties, got %v", op.AdditionalProperties)
	}
}

func TestAIAssistantRequestValidate(t *testing.T) {
	if err := (AIAssistantRequest{}).Validate(); err == nil {
		t.Fatal("expected missing user intent error")
	}
	if err := (AIAssistantRequest{UserIntent: "explain this design"}).Validate(); err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
	if err := (AIAssistantRequest{
		UserIntent:           "explain this design",
		CurrentDesignContext: json.RawMessage(`{"components":[]}`),
	}).Validate(); err != nil {
		t.Fatalf("expected valid current design context, got %v", err)
	}
	if err := (AIAssistantRequest{
		UserIntent:           "explain this design",
		CurrentDesignContext: json.RawMessage(`{"components":`),
	}).Validate(); err == nil {
		t.Fatal("expected invalid current design context error")
	}
	if err := (AIAssistantRequest{
		UserIntent:              "explain this design",
		AdditionalContextFields: json.RawMessage(`{"scope":[]}`),
	}).Validate(); err != nil {
		t.Fatalf("expected valid additional context fields, got %v", err)
	}
	if err := (AIAssistantRequest{
		UserIntent:              "explain this design",
		AdditionalContextFields: json.RawMessage(`{"scope":`),
	}).Validate(); err == nil {
		t.Fatal("expected invalid additional context fields error")
	}
}

func TestAIAssistantResponseValidate(t *testing.T) {
	tests := []struct {
		name    string
		resp    AIAssistantResponse
		wantErr bool
	}{
		{
			name:    "empty response",
			resp:    AIAssistantResponse{},
			wantErr: true,
		},
		{
			name:    "explanation",
			resp:    AIAssistantResponse{Explanation: "This design contains one service."},
			wantErr: false,
		},
		{
			name: "recommendation",
			resp: AIAssistantResponse{Recommendations: []AIAssistantRecommendation{{
				Title: "Review service selectors",
			}}},
			wantErr: false,
		},
		{
			name: "redirect",
			resp: AIAssistantResponse{Redirects: []AIAssistantRedirect{{
				Title: "Open Meshery Models",
				URL:   "/extensions/models",
			}}},
			wantErr: false,
		},
		{
			name: "structured error",
			resp: AIAssistantResponse{Errors: []AIAssistantError{{
				Code:    "CONTEXT_TOO_LARGE",
				Message: "Narrow the selected context.",
			}}},
			wantErr: false,
		},
		{
			name: "multiple output classes",
			resp: AIAssistantResponse{
				Explanation: "This design contains one service.",
				Recommendations: []AIAssistantRecommendation{{
					Title: "Review service selectors",
				}},
			},
			wantErr: true,
		},
		{
			name: "recommendation without title",
			resp: AIAssistantResponse{Recommendations: []AIAssistantRecommendation{{
				Description: "Review service selectors.",
			}}},
			wantErr: true,
		},
		{
			name: "redirect without url",
			resp: AIAssistantResponse{Redirects: []AIAssistantRedirect{{
				Title: "Open Meshery Models",
			}}},
			wantErr: true,
		},
		{
			name: "structured error without message",
			resp: AIAssistantResponse{Errors: []AIAssistantError{{
				Code: "CONTEXT_TOO_LARGE",
			}}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resp.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
