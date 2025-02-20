/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package patch

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/contexts"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/utils"
	"github.com/stretchr/testify/assert"
)

var testSolution = model.SolutionSpec{
	Scope: "default",
}

func TestPatchSolution(t *testing.T) {
	testPatchSolution := os.Getenv("TEST_PATCH_SOLUTION")
	if testPatchSolution != "yes" {
		t.Skip("Skipping becasue TEST_PATCH_SOLUTION is missing or not set to 'yes'")
	}
	provider := PatchStageProvider{}
	err := provider.Init(PatchStageProviderConfig{
		BaseUrl:  "http://localhost:8082/v1alpha2/",
		User:     "admin",
		Password: "",
	})

	provider.SetContext(&contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType":   "solution",
		"objectName":   "test-app",
		"patchSource":  "catalog",
		"patchContent": "ai-config",
		"component":    "frontend",
		"property":     "deployment.replicas",
		"subKey":       "",
		"dedupKey":     "flavor",
		"patchAction":  "add",
	})
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, "OK", outputs["status"])
}
func TestPatchSolutionWholeComponent(t *testing.T) {
	testPatchSolution := os.Getenv("TEST_PATCH_SOLUTION")
	if testPatchSolution != "yes" {
		t.Skip("Skipping becasue TEST_PATCH_SOLUTION is missing or not set to 'yes'")
	}
	provider := PatchStageProvider{}
	err := provider.Init(PatchStageProviderConfig{
		BaseUrl:  "http://localhost:8082/v1alpha2/",
		User:     "admin",
		Password: "",
	})

	provider.SetContext(&contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	})
	assert.Nil(t, err)
	outputs, _, err := provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType":  "solution",
		"objectName":  "test-app",
		"patchSource": "inline",
		"patchContent": model.ComponentSpec{
			Name: "test-ingress2",
			Type: "ingress",
			Properties: map[string]interface{}{
				"ingressClassName": "nginx",
				"rules": []map[string]interface{}{
					{
						"http": map[string]interface{}{
							"paths": []interface{}{
								map[string]interface{}{
									"path":     "/testpath",
									"backend":  map[string]interface{}{"serviceName": "test-app", "servicePort": 100 + 200},
									"pathType": "Prefix",
								},
							},
						},
					},
				},
			},
		},
		"patchAction": "add",
	})
	assert.Nil(t, err)
	assert.NotNil(t, outputs)
	assert.Equal(t, "OK", outputs["status"])
}

func TestPatchInitFromMap(t *testing.T) {
	provider := PatchStageProvider{}
	input := map[string]string{
		"baseUrl":  "http://symphony-service:8080/v1alpha2/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
	assert.Equal(t, "http://symphony-service:8080/v1alpha2/", provider.Config.BaseUrl)
	assert.Equal(t, "admin", provider.Config.User)
	assert.Equal(t, "", provider.Config.Password)

	input = map[string]string{}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"baseUrl": "",
	}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"baseUrl": "http://symphony-service:8080/v1alpha2/",
	}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"baseUrl": "http://symphony-service:8080/v1alpha2/",
		"user":    "",
	}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)

	input = map[string]string{
		"baseUrl": "http://symphony-service:8080/v1alpha2/",
		"user":    "admin",
	}
	err = provider.InitWithMap(input)
	assert.NotNil(t, err)
}

func TestPatchProcessInline(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	provider := PatchStageProvider{}
	input := map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	assert.Nil(t, err)
	testSolution = model.SolutionSpec{
		Scope: "default",
	}
	_, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType":  "solution",
		"objectName":  "solution1",
		"patchSource": "inline",
		"patchContent": model.ComponentSpec{
			Name: "ebpf-module",
			Type: "ebpf",
			Properties: map[string]interface{}{
				"ebpf.url":   "https://github.com/Haishi2016/Vault818/releases/download/vtest/hello.bpf.o",
				"ebpf.name":  "hello",
				"ebpf.event": "xdp",
			},
		},
		"patchAction": "add",
	})
	assert.Nil(t, err)
	assert.Equal(t, "ebpf-module", testSolution.Components[0].Name)
	assert.Equal(t, "ebpf", testSolution.Components[0].Type)
	assert.Equal(t, map[string]interface{}{
		"ebpf.url":   "https://github.com/Haishi2016/Vault818/releases/download/vtest/hello.bpf.o",
		"ebpf.name":  "hello",
		"ebpf.event": "xdp",
	}, testSolution.Components[0].Properties)

	_, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType":  "solution",
		"objectName":  "solution1",
		"patchSource": "inline",
		"patchContent": model.ComponentSpec{
			Name: "ebpf-module",
			Type: "ebpf",
			Properties: map[string]interface{}{
				"ebpf.url":   "https://github.com/Haishi2016/Vault818/releases/download/vtest/hello.bpf.o",
				"ebpf.name":  "hello",
				"ebpf.event": "xdp",
			},
		},
		"patchAction": "remove",
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(testSolution.Components))
}

func TestPatchProcessCatalog(t *testing.T) {
	ts := InitializeMockSymphonyAPI()
	provider := PatchStageProvider{}
	input := map[string]string{
		"baseUrl":  ts.URL + "/",
		"user":     "admin",
		"password": "",
	}
	err := provider.InitWithMap(input)
	provider.SetContext(&contexts.ManagerContext{
		VencorContext: &contexts.VendorContext{
			EvaluationContext: &utils.EvaluationContext{},
		},
	})
	assert.Nil(t, err)
	testSolution = model.SolutionSpec{
		Scope: "default",
	}
	// Step 1: first add component to solution spec
	provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType":  "solution",
		"objectName":  "solution1",
		"patchSource": "inline",
		"patchContent": model.ComponentSpec{
			Name: "ebpf-module",
			Type: "ebpf",
			Properties: map[string]interface{}{
				"ebpf.url":   "https://github.com/Haishi2016/Vault818/releases/download/vtest/hello.bpf.o",
				"ebpf.name":  "hello",
				"ebpf.event": "xdp",
				"input": map[string]interface{}{
					"adapter": []string{},
					"scope":   []string{},
				},
			},
		},
		"patchAction": "add",
	})

	// Step 2: update solution with config in catalog
	_, _, err = provider.Process(context.Background(), contexts.ManagerContext{}, map[string]interface{}{
		"objectType":   "solution",
		"objectName":   "solution1",
		"patchSource":  "catalog",
		"patchContent": "catalog1",
		"patchAction":  "add",
		"component":    "ebpf-module",
		"property":     "input",
		"subKey":       "adapter",
	})
	assert.Nil(t, err)
	assert.Equal(t, "ebpf-module", testSolution.Components[0].Name)
	assert.Equal(t, "ebpf", testSolution.Components[0].Type)
	assert.Equal(t, map[string]interface{}{
		"ebpf.url":   "https://github.com/Haishi2016/Vault818/releases/download/vtest/hello.bpf.o",
		"ebpf.name":  "hello",
		"ebpf.event": "xdp",
		"input": map[string]interface{}{
			"adapter": []interface{}{map[string]interface{}{"testkey": "0", "testdict": []interface{}{"1"}, "testmap": map[string]interface{}{}}},
			"scope":   []interface{}{},
		},
	}, testSolution.Components[0].Properties)

}

type AuthResponse struct {
	AccessToken string   `json:"accessToken"`
	TokenType   string   `json:"tokenType"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
}

func InitializeMockSymphonyAPI() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response interface{}
		switch r.URL.Path {
		case "/solutions/solution1":
			if r.Method == "GET" {
				response = model.SolutionState{
					Id:   "solution1",
					Spec: &testSolution,
				}
			} else {
				body, _ := io.ReadAll(r.Body)
				newSpec := model.SolutionSpec{}
				json.Unmarshal(body, &newSpec)
				testSolution = newSpec
				response = model.SolutionState{
					Id:   "solution1",
					Spec: &testSolution,
				}
			}
		case "/catalogs/registry/catalog1":
			response = model.CatalogState{
				Id: "catalog1",
				Spec: &model.CatalogSpec{
					Type: "config",
					Name: "catalog1",
					Properties: map[string]interface{}{
						"testkey":  "0",
						"testdict": []string{"1"},
						"testmap":  map[string]interface{}{},
					},
				},
			}

		default:
			response = AuthResponse{
				AccessToken: "test-token",
				TokenType:   "Bearer",
				Username:    "test-user",
				Roles:       []string{"role1", "role2"},
			}
		}

		json.NewEncoder(w).Encode(response)
	}))
	return ts
}
