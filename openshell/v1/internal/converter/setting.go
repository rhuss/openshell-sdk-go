// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"fmt"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1/types"
	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	sbv1 "github.com/rhuss/openshell-sdk-go/proto/sandboxv1"
	"google.golang.org/protobuf/proto"
)

// --- SettingValue oneof conversion ---

// SettingValueFromProto converts a proto SettingValue (oneof) to an SDK SettingValue.
func SettingValueFromProto(pv *sbv1.SettingValue) *v1.SettingValue {
	if pv == nil {
		return nil
	}
	sv := &v1.SettingValue{}
	switch v := pv.GetValue().(type) {
	case *sbv1.SettingValue_StringValue:
		sv.Type = v1.SettingValueString
		sv.StringVal = v.StringValue
	case *sbv1.SettingValue_BoolValue:
		sv.Type = v1.SettingValueBool
		sv.BoolVal = v.BoolValue
	case *sbv1.SettingValue_IntValue:
		sv.Type = v1.SettingValueInt
		sv.IntVal = v.IntValue
	case *sbv1.SettingValue_BytesValue:
		sv.Type = v1.SettingValueBytes
		sv.BytesVal = CopyByteSlice(v.BytesValue)
	}
	return sv
}

// SettingValueToProto converts an SDK SettingValue to a proto SettingValue (oneof).
func SettingValueToProto(sv *v1.SettingValue) *sbv1.SettingValue {
	if sv == nil {
		return nil
	}
	pv := &sbv1.SettingValue{}
	switch sv.Type {
	case v1.SettingValueString:
		pv.Value = &sbv1.SettingValue_StringValue{StringValue: sv.StringVal}
	case v1.SettingValueBool:
		pv.Value = &sbv1.SettingValue_BoolValue{BoolValue: sv.BoolVal}
	case v1.SettingValueInt:
		pv.Value = &sbv1.SettingValue_IntValue{IntValue: sv.IntVal}
	case v1.SettingValueBytes:
		pv.Value = &sbv1.SettingValue_BytesValue{BytesValue: CopyByteSlice(sv.BytesVal)}
	}
	return pv
}

// --- Enum conversions ---

// SettingScopeFromProto converts a proto SettingScope enum to an SDK SettingScope.
func SettingScopeFromProto(ps sbv1.SettingScope) v1.SettingScope {
	switch ps {
	case sbv1.SettingScope_SETTING_SCOPE_UNSPECIFIED:
		return v1.SettingScopeUnspecified
	case sbv1.SettingScope_SETTING_SCOPE_SANDBOX:
		return v1.SettingScopeSandbox
	case sbv1.SettingScope_SETTING_SCOPE_GLOBAL:
		return v1.SettingScopeGlobal
	default:
		return v1.SettingScope("")
	}
}

// SettingScopeToProto converts an SDK SettingScope to a proto SettingScope enum.
func SettingScopeToProto(s v1.SettingScope) sbv1.SettingScope {
	switch s {
	case v1.SettingScopeUnspecified:
		return sbv1.SettingScope_SETTING_SCOPE_UNSPECIFIED
	case v1.SettingScopeSandbox:
		return sbv1.SettingScope_SETTING_SCOPE_SANDBOX
	case v1.SettingScopeGlobal:
		return sbv1.SettingScope_SETTING_SCOPE_GLOBAL
	default:
		return sbv1.SettingScope_SETTING_SCOPE_UNSPECIFIED
	}
}

// PolicySourceFromProto converts a proto PolicySource enum to an SDK PolicySource.
func PolicySourceFromProto(ps sbv1.PolicySource) v1.PolicySource {
	switch ps {
	case sbv1.PolicySource_POLICY_SOURCE_UNSPECIFIED:
		return v1.PolicySourceUnspecified
	case sbv1.PolicySource_POLICY_SOURCE_SANDBOX:
		return v1.PolicySourceSandbox
	case sbv1.PolicySource_POLICY_SOURCE_GLOBAL:
		return v1.PolicySourceGlobal
	default:
		return v1.PolicySource("")
	}
}

// --- EffectiveSetting ---

// EffectiveSettingFromProto converts a proto EffectiveSetting to an SDK EffectiveSetting.
func EffectiveSettingFromProto(pv *sbv1.EffectiveSetting) *v1.EffectiveSetting {
	if pv == nil {
		return nil
	}
	es := &v1.EffectiveSetting{
		Scope: SettingScopeFromProto(pv.GetScope()),
	}
	if sv := SettingValueFromProto(pv.GetValue()); sv != nil {
		es.Value = *sv
	}
	return es
}

// --- SandboxConfig ---

// SandboxConfigFromProto converts a GetSandboxConfigResponse to an SDK SandboxConfig.
func SandboxConfigFromProto(resp *sbv1.GetSandboxConfigResponse) *v1.SandboxConfig {
	if resp == nil {
		return nil
	}
	sc := &v1.SandboxConfig{
		PolicyVersion:       resp.GetVersion(),
		PolicyHash:          resp.GetPolicyHash(),
		ConfigRevision:      resp.GetConfigRevision(),
		PolicySource:        PolicySourceFromProto(resp.GetPolicySource()),
		GlobalPolicyVersion: resp.GetGlobalPolicyVersion(),
		ProviderEnvRevision: resp.GetProviderEnvRevision(),
	}

	// Serialize SandboxPolicy to opaque bytes.
	if p := resp.GetPolicy(); p != nil {
		if b, err := proto.Marshal(p); err == nil {
			sc.Policy = b
		}
	}

	// Deep-copy settings map.
	if m := resp.GetSettings(); len(m) > 0 {
		sc.Settings = make(map[string]v1.EffectiveSetting, len(m))
		for k, v := range m {
			if es := EffectiveSettingFromProto(v); es != nil {
				sc.Settings[k] = *es
			}
		}
	}

	return sc
}

// --- GatewayConfig ---

// GatewayConfigFromProto converts a GetGatewayConfigResponse to an SDK GatewayConfig.
func GatewayConfigFromProto(resp *sbv1.GetGatewayConfigResponse) *v1.GatewayConfig {
	if resp == nil {
		return nil
	}
	gc := &v1.GatewayConfig{
		SettingsRevision: resp.GetSettingsRevision(),
	}

	// Deep-copy settings map.
	if m := resp.GetSettings(); len(m) > 0 {
		gc.Settings = make(map[string]v1.SettingValue, len(m))
		for k, v := range m {
			if sv := SettingValueFromProto(v); sv != nil {
				gc.Settings[k] = *sv
			}
		}
	}

	return gc
}

// --- ConfigUpdate ---

// ConfigUpdateToProto converts an SDK ConfigUpdate to an UpdateConfigRequest.
func ConfigUpdateToProto(cu *v1.ConfigUpdate) (*pb.UpdateConfigRequest, error) {
	if cu == nil {
		return nil, nil
	}
	req := &pb.UpdateConfigRequest{
		Name:                    cu.Name,
		SettingKey:              cu.SettingKey,
		SettingValue:            SettingValueToProto(cu.SettingValue),
		DeleteSetting:           cu.DeleteSetting,
		Global:                  cu.Global,
		ExpectedResourceVersion: cu.ExpectedResourceVersion,
	}

	if cu.Policy != nil {
		policy := &sbv1.SandboxPolicy{}
		if err := proto.Unmarshal(cu.Policy, policy); err != nil {
			return nil, fmt.Errorf("invalid policy bytes: %w", err)
		}
		req.Policy = policy
	}

	return req, nil
}

// --- ConfigUpdateResult ---

// ConfigUpdateResultFromProto converts an UpdateConfigResponse to an SDK ConfigUpdateResult.
func ConfigUpdateResultFromProto(resp *pb.UpdateConfigResponse) *v1.ConfigUpdateResult {
	if resp == nil {
		return nil
	}
	return &v1.ConfigUpdateResult{
		Version:          resp.GetVersion(),
		PolicyHash:       resp.GetPolicyHash(),
		SettingsRevision: resp.GetSettingsRevision(),
		Deleted:          resp.GetDeleted(),
	}
}
