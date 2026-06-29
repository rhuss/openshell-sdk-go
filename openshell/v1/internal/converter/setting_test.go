// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"testing"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1/types"
	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	sbv1 "github.com/rhuss/openshell-sdk-go/proto/sandboxv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// --- SettingValue oneof mapping ---

func TestSettingValueFromProto_StringValue(t *testing.T) {
	pv := &sbv1.SettingValue{
		Value: &sbv1.SettingValue_StringValue{StringValue: "hello"},
	}

	sv := SettingValueFromProto(pv)

	require.NotNil(t, sv)
	assert.Equal(t, v1.SettingValueString, sv.Type)
	assert.Equal(t, "hello", sv.StringVal)
	assert.False(t, sv.BoolVal)
	assert.Zero(t, sv.IntVal)
	assert.Nil(t, sv.BytesVal)
}

func TestSettingValueFromProto_BoolValue(t *testing.T) {
	pv := &sbv1.SettingValue{
		Value: &sbv1.SettingValue_BoolValue{BoolValue: true},
	}

	sv := SettingValueFromProto(pv)

	require.NotNil(t, sv)
	assert.Equal(t, v1.SettingValueBool, sv.Type)
	assert.True(t, sv.BoolVal)
	assert.Empty(t, sv.StringVal)
}

func TestSettingValueFromProto_IntValue(t *testing.T) {
	pv := &sbv1.SettingValue{
		Value: &sbv1.SettingValue_IntValue{IntValue: 42},
	}

	sv := SettingValueFromProto(pv)

	require.NotNil(t, sv)
	assert.Equal(t, v1.SettingValueInt, sv.Type)
	assert.Equal(t, int64(42), sv.IntVal)
}

func TestSettingValueFromProto_BytesValue(t *testing.T) {
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	pv := &sbv1.SettingValue{
		Value: &sbv1.SettingValue_BytesValue{BytesValue: data},
	}

	sv := SettingValueFromProto(pv)

	require.NotNil(t, sv)
	assert.Equal(t, v1.SettingValueBytes, sv.Type)
	assert.Equal(t, data, sv.BytesVal)
}

func TestSettingValueFromProto_BytesDeepCopy(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	pv := &sbv1.SettingValue{
		Value: &sbv1.SettingValue_BytesValue{BytesValue: data},
	}

	sv := SettingValueFromProto(pv)

	require.NotNil(t, sv)
	// Mutate original data — SDK copy must not be affected.
	data[0] = 0xFF
	assert.Equal(t, byte(0x01), sv.BytesVal[0], "deep copy must isolate SDK from proto")
}

func TestSettingValueFromProto_NilOneof(t *testing.T) {
	pv := &sbv1.SettingValue{}

	sv := SettingValueFromProto(pv)

	require.NotNil(t, sv)
	assert.Equal(t, v1.SettingValueType(""), sv.Type)
}

func TestSettingValueFromProto_Nil(t *testing.T) {
	sv := SettingValueFromProto(nil)
	assert.Nil(t, sv)
}

func TestSettingValueToProto_StringValue(t *testing.T) {
	sv := &v1.SettingValue{
		Type:      v1.SettingValueString,
		StringVal: "world",
	}

	pv := SettingValueToProto(sv)

	require.NotNil(t, pv)
	assert.Equal(t, "world", pv.GetStringValue())
}

func TestSettingValueToProto_BoolValue(t *testing.T) {
	sv := &v1.SettingValue{
		Type:    v1.SettingValueBool,
		BoolVal: true,
	}

	pv := SettingValueToProto(sv)

	require.NotNil(t, pv)
	assert.True(t, pv.GetBoolValue())
}

func TestSettingValueToProto_IntValue(t *testing.T) {
	sv := &v1.SettingValue{
		Type:   v1.SettingValueInt,
		IntVal: 99,
	}

	pv := SettingValueToProto(sv)

	require.NotNil(t, pv)
	assert.Equal(t, int64(99), pv.GetIntValue())
}

func TestSettingValueToProto_BytesValue(t *testing.T) {
	data := []byte{0xCA, 0xFE}
	sv := &v1.SettingValue{
		Type:     v1.SettingValueBytes,
		BytesVal: data,
	}

	pv := SettingValueToProto(sv)

	require.NotNil(t, pv)
	assert.Equal(t, data, pv.GetBytesValue())
}

func TestSettingValueToProto_Nil(t *testing.T) {
	pv := SettingValueToProto(nil)
	assert.Nil(t, pv)
}

// --- SettingScope enum mapping ---

func TestSettingScopeFromProto(t *testing.T) {
	tests := []struct {
		proto sbv1.SettingScope
		want  v1.SettingScope
	}{
		{sbv1.SettingScope_SETTING_SCOPE_UNSPECIFIED, v1.SettingScopeUnspecified},
		{sbv1.SettingScope_SETTING_SCOPE_SANDBOX, v1.SettingScopeSandbox},
		{sbv1.SettingScope_SETTING_SCOPE_GLOBAL, v1.SettingScopeGlobal},
		{sbv1.SettingScope(999), v1.SettingScope("")},
	}
	for _, tt := range tests {
		t.Run(tt.proto.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, SettingScopeFromProto(tt.proto))
		})
	}
}

func TestSettingScopeToProto(t *testing.T) {
	tests := []struct {
		sdk  v1.SettingScope
		want sbv1.SettingScope
	}{
		{v1.SettingScopeUnspecified, sbv1.SettingScope_SETTING_SCOPE_UNSPECIFIED},
		{v1.SettingScopeSandbox, sbv1.SettingScope_SETTING_SCOPE_SANDBOX},
		{v1.SettingScopeGlobal, sbv1.SettingScope_SETTING_SCOPE_GLOBAL},
		{v1.SettingScope("unknown"), sbv1.SettingScope_SETTING_SCOPE_UNSPECIFIED},
	}
	for _, tt := range tests {
		t.Run(string(tt.sdk), func(t *testing.T) {
			assert.Equal(t, tt.want, SettingScopeToProto(tt.sdk))
		})
	}
}

// --- PolicySource enum mapping ---

func TestPolicySourceFromProto(t *testing.T) {
	tests := []struct {
		proto sbv1.PolicySource
		want  v1.PolicySource
	}{
		{sbv1.PolicySource_POLICY_SOURCE_UNSPECIFIED, v1.PolicySourceUnspecified},
		{sbv1.PolicySource_POLICY_SOURCE_SANDBOX, v1.PolicySourceSandbox},
		{sbv1.PolicySource_POLICY_SOURCE_GLOBAL, v1.PolicySourceGlobal},
		{sbv1.PolicySource(999), v1.PolicySource("")},
	}
	for _, tt := range tests {
		t.Run(tt.proto.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, PolicySourceFromProto(tt.proto))
		})
	}
}

// --- EffectiveSetting ---

func TestEffectiveSettingFromProto(t *testing.T) {
	pv := &sbv1.EffectiveSetting{
		Value: &sbv1.SettingValue{
			Value: &sbv1.SettingValue_StringValue{StringValue: "val"},
		},
		Scope: sbv1.SettingScope_SETTING_SCOPE_SANDBOX,
	}

	es := EffectiveSettingFromProto(pv)

	require.NotNil(t, es)
	assert.Equal(t, v1.SettingValueString, es.Value.Type)
	assert.Equal(t, "val", es.Value.StringVal)
	assert.Equal(t, v1.SettingScopeSandbox, es.Scope)
}

func TestEffectiveSettingFromProto_NilValue(t *testing.T) {
	pv := &sbv1.EffectiveSetting{
		Scope: sbv1.SettingScope_SETTING_SCOPE_GLOBAL,
	}

	es := EffectiveSettingFromProto(pv)

	require.NotNil(t, es)
	assert.Equal(t, v1.SettingValueType(""), es.Value.Type)
	assert.Equal(t, v1.SettingScopeGlobal, es.Scope)
}

func TestEffectiveSettingFromProto_Nil(t *testing.T) {
	es := EffectiveSettingFromProto(nil)
	assert.Nil(t, es)
}

// --- SandboxConfig (GetSandboxConfigResponse → SandboxConfig) ---

func TestSandboxConfigFromProto(t *testing.T) {
	policy := &sbv1.SandboxPolicy{}
	policyBytes, err := proto.Marshal(policy)
	require.NoError(t, err)

	resp := &sbv1.GetSandboxConfigResponse{
		Policy:  policy,
		Version: 3,
		PolicyHash:          "sha256:abc",
		Settings: map[string]*sbv1.EffectiveSetting{
			"timeout": {
				Value: &sbv1.SettingValue{
					Value: &sbv1.SettingValue_IntValue{IntValue: 30},
				},
				Scope: sbv1.SettingScope_SETTING_SCOPE_SANDBOX,
			},
			"debug": {
				Value: &sbv1.SettingValue{
					Value: &sbv1.SettingValue_BoolValue{BoolValue: true},
				},
				Scope: sbv1.SettingScope_SETTING_SCOPE_GLOBAL,
			},
		},
		ConfigRevision:      100,
		PolicySource:        sbv1.PolicySource_POLICY_SOURCE_SANDBOX,
		GlobalPolicyVersion: 5,
		ProviderEnvRevision: 200,
	}

	sc := SandboxConfigFromProto(resp)

	require.NotNil(t, sc)
	assert.Equal(t, policyBytes, sc.Policy)
	assert.Equal(t, uint32(3), sc.PolicyVersion)
	assert.Equal(t, "sha256:abc", sc.PolicyHash)
	assert.Equal(t, uint64(100), sc.ConfigRevision)
	assert.Equal(t, v1.PolicySourceSandbox, sc.PolicySource)
	assert.Equal(t, uint32(5), sc.GlobalPolicyVersion)
	assert.Equal(t, uint64(200), sc.ProviderEnvRevision)

	require.Len(t, sc.Settings, 2)

	timeout := sc.Settings["timeout"]
	assert.Equal(t, v1.SettingValueInt, timeout.Value.Type)
	assert.Equal(t, int64(30), timeout.Value.IntVal)
	assert.Equal(t, v1.SettingScopeSandbox, timeout.Scope)

	debug := sc.Settings["debug"]
	assert.Equal(t, v1.SettingValueBool, debug.Value.Type)
	assert.True(t, debug.Value.BoolVal)
	assert.Equal(t, v1.SettingScopeGlobal, debug.Scope)
}

func TestSandboxConfigFromProto_NilPolicy(t *testing.T) {
	resp := &sbv1.GetSandboxConfigResponse{
		Version:    1,
		PolicyHash: "sha256:empty",
	}

	sc := SandboxConfigFromProto(resp)

	require.NotNil(t, sc)
	assert.Nil(t, sc.Policy)
	assert.Equal(t, uint32(1), sc.PolicyVersion)
	assert.Empty(t, sc.Settings)
}

func TestSandboxConfigFromProto_Nil(t *testing.T) {
	sc := SandboxConfigFromProto(nil)
	assert.Nil(t, sc)
}

func TestSandboxConfigFromProto_SettingsDeepCopy(t *testing.T) {
	resp := &sbv1.GetSandboxConfigResponse{
		Settings: map[string]*sbv1.EffectiveSetting{
			"key1": {
				Value: &sbv1.SettingValue{
					Value: &sbv1.SettingValue_StringValue{StringValue: "original"},
				},
				Scope: sbv1.SettingScope_SETTING_SCOPE_SANDBOX,
			},
		},
	}

	sc := SandboxConfigFromProto(resp)

	require.NotNil(t, sc)
	// Mutate the proto map — SDK map must not be affected.
	resp.Settings["key1"].Value.Value = &sbv1.SettingValue_StringValue{StringValue: "mutated"}
	assert.Equal(t, "original", sc.Settings["key1"].Value.StringVal,
		"deep copy must isolate SDK settings from proto")
}

// --- GatewayConfig (GetGatewayConfigResponse → GatewayConfig) ---

func TestGatewayConfigFromProto(t *testing.T) {
	resp := &sbv1.GetGatewayConfigResponse{
		Settings: map[string]*sbv1.SettingValue{
			"region": {
				Value: &sbv1.SettingValue_StringValue{StringValue: "us-west-2"},
			},
			"max_sandboxes": {
				Value: &sbv1.SettingValue_IntValue{IntValue: 100},
			},
		},
		SettingsRevision: 42,
	}

	gc := GatewayConfigFromProto(resp)

	require.NotNil(t, gc)
	assert.Equal(t, uint64(42), gc.SettingsRevision)
	require.Len(t, gc.Settings, 2)

	region := gc.Settings["region"]
	assert.Equal(t, v1.SettingValueString, region.Type)
	assert.Equal(t, "us-west-2", region.StringVal)

	maxSb := gc.Settings["max_sandboxes"]
	assert.Equal(t, v1.SettingValueInt, maxSb.Type)
	assert.Equal(t, int64(100), maxSb.IntVal)
}

func TestGatewayConfigFromProto_EmptySettings(t *testing.T) {
	resp := &sbv1.GetGatewayConfigResponse{
		SettingsRevision: 1,
	}

	gc := GatewayConfigFromProto(resp)

	require.NotNil(t, gc)
	assert.Equal(t, uint64(1), gc.SettingsRevision)
	assert.Empty(t, gc.Settings)
}

func TestGatewayConfigFromProto_Nil(t *testing.T) {
	gc := GatewayConfigFromProto(nil)
	assert.Nil(t, gc)
}

func TestGatewayConfigFromProto_SettingsDeepCopy(t *testing.T) {
	resp := &sbv1.GetGatewayConfigResponse{
		Settings: map[string]*sbv1.SettingValue{
			"key": {
				Value: &sbv1.SettingValue_StringValue{StringValue: "original"},
			},
		},
	}

	gc := GatewayConfigFromProto(resp)

	require.NotNil(t, gc)
	// Mutate the proto map — SDK map must not be affected.
	resp.Settings["key"].Value = &sbv1.SettingValue_StringValue{StringValue: "mutated"}
	assert.Equal(t, "original", gc.Settings["key"].StringVal,
		"deep copy must isolate SDK settings from proto")
}

// --- ConfigUpdate (ConfigUpdate → UpdateConfigRequest) ---

func TestConfigUpdateToProto(t *testing.T) {
	cu := &v1.ConfigUpdate{
		Name:       "my-sandbox",
		SettingKey: "timeout",
		SettingValue: &v1.SettingValue{
			Type:   v1.SettingValueInt,
			IntVal: 60,
		},
		DeleteSetting:           false,
		Global:                  false,
		ExpectedResourceVersion: 7,
	}

	req, err := ConfigUpdateToProto(cu)
	require.NoError(t, err)

	require.NotNil(t, req)
	assert.Equal(t, "my-sandbox", req.Name)
	assert.Equal(t, "timeout", req.SettingKey)
	require.NotNil(t, req.SettingValue)
	assert.Equal(t, int64(60), req.SettingValue.GetIntValue())
	assert.False(t, req.DeleteSetting)
	assert.False(t, req.Global)
	assert.Equal(t, uint64(7), req.ExpectedResourceVersion)
	assert.Nil(t, req.Policy)
	assert.Empty(t, req.MergeOperations)
}

func TestConfigUpdateToProto_WithPolicy(t *testing.T) {
	// Policy bytes must be valid proto-encoded SandboxPolicy.
	policy := &sbv1.SandboxPolicy{}
	policyBytes, err := proto.Marshal(policy)
	require.NoError(t, err)

	cu := &v1.ConfigUpdate{
		Name:   "sb-policy",
		Policy: policyBytes,
	}

	req, err := ConfigUpdateToProto(cu)
	require.NoError(t, err)

	require.NotNil(t, req)
	require.NotNil(t, req.Policy, "valid proto bytes must deserialize into SandboxPolicy")
}

func TestConfigUpdateToProto_WithDeleteSetting(t *testing.T) {
	cu := &v1.ConfigUpdate{
		Name:          "sb-del",
		SettingKey:    "obsolete-key",
		DeleteSetting: true,
	}

	req, err := ConfigUpdateToProto(cu)
	require.NoError(t, err)

	require.NotNil(t, req)
	assert.Equal(t, "obsolete-key", req.SettingKey)
	assert.True(t, req.DeleteSetting)
}

func TestConfigUpdateToProto_GlobalScope(t *testing.T) {
	cu := &v1.ConfigUpdate{
		SettingKey: "global-setting",
		SettingValue: &v1.SettingValue{
			Type:      v1.SettingValueString,
			StringVal: "global-val",
		},
		Global: true,
	}

	req, err := ConfigUpdateToProto(cu)
	require.NoError(t, err)

	require.NotNil(t, req)
	assert.True(t, req.Global)
	assert.Empty(t, req.Name)
}

func TestConfigUpdateToProto_NilSettingValue(t *testing.T) {
	cu := &v1.ConfigUpdate{
		Name:       "sb-nil",
		SettingKey: "key",
	}

	req, err := ConfigUpdateToProto(cu)
	require.NoError(t, err)

	require.NotNil(t, req)
	assert.Nil(t, req.SettingValue)
}

func TestConfigUpdateToProto_Nil(t *testing.T) {
	req, err := ConfigUpdateToProto(nil)
	require.NoError(t, err)
	assert.Nil(t, req)
}

func TestConfigUpdateToProto_InvalidPolicyBytes(t *testing.T) {
	cu := &v1.ConfigUpdate{
		Name:   "sb-bad-policy",
		Policy: []byte{0xFF, 0xFE, 0xFD},
	}

	_, err := ConfigUpdateToProto(cu)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid policy bytes")
}

// --- ConfigUpdateResult (UpdateConfigResponse → ConfigUpdateResult) ---

func TestConfigUpdateResultFromProto(t *testing.T) {
	resp := &pb.UpdateConfigResponse{
		Version:          10,
		PolicyHash:       "sha256:updated",
		SettingsRevision: 55,
		Deleted:          true,
	}

	result := ConfigUpdateResultFromProto(resp)

	require.NotNil(t, result)
	assert.Equal(t, uint32(10), result.Version)
	assert.Equal(t, "sha256:updated", result.PolicyHash)
	assert.Equal(t, uint64(55), result.SettingsRevision)
	assert.True(t, result.Deleted)
}

func TestConfigUpdateResultFromProto_DefaultValues(t *testing.T) {
	resp := &pb.UpdateConfigResponse{}

	result := ConfigUpdateResultFromProto(resp)

	require.NotNil(t, result)
	assert.Zero(t, result.Version)
	assert.Empty(t, result.PolicyHash)
	assert.Zero(t, result.SettingsRevision)
	assert.False(t, result.Deleted)
}

func TestConfigUpdateResultFromProto_Nil(t *testing.T) {
	result := ConfigUpdateResultFromProto(nil)
	assert.Nil(t, result)
}

// --- CopyByteSlice helper ---

func TestCopyByteSlice(t *testing.T) {
	original := []byte{0x01, 0x02, 0x03}
	copied := CopyByteSlice(original)

	assert.Equal(t, original, copied)

	// Mutate original — copy must not be affected.
	original[0] = 0xFF
	assert.Equal(t, byte(0x01), copied[0], "copy must be independent of original")
}

func TestCopyByteSlice_Nil(t *testing.T) {
	copied := CopyByteSlice(nil)
	assert.Nil(t, copied)
}

func TestCopyByteSlice_Empty(t *testing.T) {
	original := []byte{}
	copied := CopyByteSlice(original)
	assert.NotNil(t, copied)
	assert.Empty(t, copied)
}
