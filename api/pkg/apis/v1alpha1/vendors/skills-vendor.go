/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package vendors

import (
	"encoding/json"

	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/managers/skills"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/model"
	"github.com/eclipse-symphony/symphony/api/pkg/apis/v1alpha1/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/managers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability"
	observ_utils "github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/observability/utils"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/providers/pubsub"
	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2/vendors"
	"github.com/eclipse-symphony/symphony/coa/pkg/logger"
	"github.com/valyala/fasthttp"
)

var kLog = logger.NewLogger("coa.runtime")

type SkillsVendor struct {
	vendors.Vendor
	SkillsManager *skills.SkillsManager
}

func (o *SkillsVendor) GetInfo() vendors.VendorInfo {
	return vendors.VendorInfo{
		Version:  o.Vendor.Version,
		Name:     "Skills",
		Producer: "Microsoft",
	}
}

func (e *SkillsVendor) Init(config vendors.VendorConfig, factories []managers.IManagerFactroy, providers map[string]map[string]providers.IProvider, pubsubProvider pubsub.IPubSubProvider) error {
	err := e.Vendor.Init(config, factories, providers, pubsubProvider)
	if err != nil {
		return err
	}
	for _, m := range e.Managers {
		if c, ok := m.(*skills.SkillsManager); ok {
			e.SkillsManager = c
		}
	}
	if e.SkillsManager == nil {
		return v1alpha2.NewCOAError(nil, "skills manager is not supplied", v1alpha2.MissingConfig)
	}
	return nil
}

func (o *SkillsVendor) GetEndpoints() []v1alpha2.Endpoint {
	route := "skills"
	if o.Route != "" {
		route = o.Route
	}
	return []v1alpha2.Endpoint{
		{
			Methods:    []string{fasthttp.MethodGet, fasthttp.MethodPost, fasthttp.MethodDelete},
			Route:      route,
			Version:    o.Version,
			Handler:    o.onSkills,
			Parameters: []string{"name?"},
		},
	}
}

func (c *SkillsVendor) onSkills(request v1alpha2.COARequest) v1alpha2.COAResponse {
	pCtx, span := observability.StartSpan("Skills Vendor", request.Context, &map[string]string{
		"method": "onSkills",
	})
	defer span.End()
	kLog.Debugf("V (Skills): onSkills, traceId: %s", request.Method, span.SpanContext().TraceID().String())

	switch request.Method {
	case fasthttp.MethodGet:
		ctx, span := observability.StartSpan("onSkills-GET", pCtx, nil)
		id := request.Parameters["__name"]
		var err error
		var state interface{}
		isArray := false
		if id == "" {
			state, err = c.SkillsManager.ListSpec(ctx)
			isArray = true
		} else {
			state, err = c.SkillsManager.GetSpec(ctx, id)
		}
		if err != nil {
			if isArray {
				kLog.Errorf(" V (Skills): onSkills failed to ListSpec, err: %v, traceId: %s", err, span.SpanContext().TraceID().String())
			} else {
				kLog.Errorf(" V (Skills): onSkills failed to GetSpec, id: %s, err: %v, traceId: %s", id, err, span.SpanContext().TraceID().String())
			}
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		jData, _ := utils.FormatObject(state, isArray, request.Parameters["path"], request.Parameters["doc-type"])
		resp := observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State:       v1alpha2.OK,
			Body:        jData,
			ContentType: "application/json",
		})
		if request.Parameters["doc-type"] == "yaml" {
			resp.ContentType = "application/text"
		}
		return resp
	case fasthttp.MethodPost:
		ctx, span := observability.StartSpan("onSkills-POST", pCtx, nil)
		id := request.Parameters["__name"]

		var skill model.SkillSpec

		err := json.Unmarshal(request.Body, &skill)
		if err != nil {
			kLog.Errorf("V (Skills): onSkills failed to pause skill from request body, error: %v traceId: %s", err, span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}

		err = c.SkillsManager.UpsertSpec(ctx, id, skill)
		if err != nil {
			kLog.Errorf("V (Skills): onSkills failed to UpsertSpec, id: %s, error: %v traceId: %s", id, err, span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	case fasthttp.MethodDelete:
		ctx, span := observability.StartSpan("onSkills-DELETE", pCtx, nil)
		id := request.Parameters["__name"]
		err := c.SkillsManager.DeleteSpec(ctx, id)
		if err != nil {
			kLog.Errorf("V (Skills): onSkills failed to DeleteSpec, id: %s, error: %v traceId: %s", id, err, span.SpanContext().TraceID().String())
			return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
				State: v1alpha2.InternalError,
				Body:  []byte(err.Error()),
			})
		}
		return observ_utils.CloseSpanWithCOAResponse(span, v1alpha2.COAResponse{
			State: v1alpha2.OK,
		})
	}
	kLog.Errorf("V (Skills): onSkills returned MethodNotAllowed, traceId: %s", span.SpanContext().TraceID().String())
	resp := v1alpha2.COAResponse{
		State:       v1alpha2.MethodNotAllowed,
		Body:        []byte("{\"result\":\"405 - method not allowed\"}"),
		ContentType: "application/json",
	}
	observ_utils.UpdateSpanStatusFromCOAResponse(span, resp)
	return resp
}
