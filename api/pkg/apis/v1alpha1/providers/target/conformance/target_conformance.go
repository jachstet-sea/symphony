/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package conformance

import (
	"context"
	"testing"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/providers/target"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func RequiredPropertiesAndMetadata[P target.ITargetProvider](t *testing.T, p P) {
	desired := []model.ComponentSpec{
		{
			Name:       "test-1",
			Properties: map[string]interface{}{},
			Metadata:   map[string]string{},
		},
	}

	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Component: model.ComponentSpec{
					Name:       "test-1",
					Properties: map[string]interface{}{},
					Metadata:   map[string]string{},
				},
			},
		},
	}

	rule := p.GetValidationRule(context.Background())

	for _, property := range rule.RequiredProperties {
		desired[0].Properties[property] = "dummy property"
		step.Components[0].Component.Properties[property] = "dummy property"
	}

	for _, metadata := range rule.RequiredMetadata {
		desired[0].Metadata[metadata] = "dummy metadata"
		step.Components[0].Component.Metadata[metadata] = "dummy metadata"
	}

	deployment := model.DeploymentSpec{
		Solution: model.SolutionSpec{
			Components: desired,
		},
		ComponentStartIndex: 0,
		ComponentEndIndex:   1,
	}
	_, err := p.Apply(context.Background(), deployment, step, true)
	assert.Nil(t, err)
}
func AnyRequiredPropertiesMissing[P target.ITargetProvider](t *testing.T, p P) {

	desired := []model.ComponentSpec{
		{
			Name:       "test-1",
			Properties: map[string]interface{}{},
			Metadata:   map[string]string{},
		},
	}

	step := model.DeploymentStep{
		Components: []model.ComponentStep{
			{
				Component: model.ComponentSpec{
					Name:       "test-1",
					Properties: map[string]interface{}{},
					Metadata:   map[string]string{},
				},
			},
		},
	}

	rule := p.GetValidationRule(context.Background())

	for _, metadata := range rule.RequiredMetadata {
		desired[0].Metadata[metadata] = "dummy metadata"
	}

	for i, _ := range rule.RequiredProperties {
		desired[0].Properties = make(map[string]interface{}, len(rule.RequiredProperties)-1)
		slice := append(append([]string{}, rule.RequiredProperties[:i]...), rule.RequiredProperties[i+1:]...)
		for _, property := range slice {
			desired[0].Properties[property] = "dummy property"
		}
		deployment := model.DeploymentSpec{
			Solution: model.SolutionSpec{
				Components: desired,
			},
			ComponentStartIndex: 0,
			ComponentEndIndex:   1,
		}
		_, err := p.Apply(context.Background(), deployment, step, true)
		assert.NotNil(t, err)
		coaErr := err.(v1alpha2.COAError)
		assert.Equal(t, v1alpha2.BadRequest, coaErr.State)
	}
}
func ConformanceSuite[P target.ITargetProvider](t *testing.T, p P) {
	t.Run("Level=Basic", func(t *testing.T) {
		RequiredPropertiesAndMetadata(t, p)
		AnyRequiredPropertiesMissing(t, p)
	})
}
