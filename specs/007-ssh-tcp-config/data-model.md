# Data Model: SSH, TCP Forwarding, and Config (Phase 2b-1)

**Date**: 2026-06-29
**Spec**: [spec.md](spec.md)

## New Domain Types (v1/types/)

### SSH Types (types/ssh.go)

```
SSHSession
├── SandboxID          string     // Sandbox this session connects to
├── Token              string     // Session token for gateway tunnel (sensitive)
├── GatewayHost        string     // Gateway host for SSH proxy connection
├── GatewayPort        uint32     // Gateway port (1-65535)
├── GatewayScheme      string     // "http" or "https"
├── HostKeyFingerprint string     // Optional host key fingerprint
└── ExpiresAtMs        int64      // Expiry in ms since epoch (0 = no expiry)
```

### Config Types (types/setting.go)

```
SettingValueType (string enum)
├── SettingValueString   = "string"
├── SettingValueBool     = "bool"
├── SettingValueInt      = "int"
└── SettingValueBytes    = "bytes"

SettingValue
├── Type       SettingValueType
├── StringVal  string
├── BoolVal    bool
├── IntVal     int64
└── BytesVal   []byte

SettingScope (string enum)
├── SettingScopeUnspecified = ""
├── SettingScopeSandbox     = "sandbox"
└── SettingScopeGlobal      = "global"

PolicySource (string enum)
├── PolicySourceUnspecified = ""
├── PolicySourceSandbox     = "sandbox"
└── PolicySourceGlobal      = "global"

EffectiveSetting
├── Value  SettingValue
└── Scope  SettingScope

SandboxConfig
├── Policy              []byte                        // Opaque SandboxPolicy proto bytes
├── PolicyVersion       uint32                        // Monotonically increasing per sandbox
├── PolicyHash          string                        // SHA-256 of serialized policy
├── Settings            map[string]EffectiveSetting   // Effective settings resolved for sandbox
├── ConfigRevision      uint64                        // Fingerprint for effective config
├── PolicySource        PolicySource                  // Where policy came from (sandbox/global)
├── GlobalPolicyVersion uint32                        // Global policy version (0 if N/A)
└── ProviderEnvRevision uint64                        // Provider credential fingerprint

GatewayConfig
├── Settings         map[string]SettingValue  // Global settings map
└── SettingsRevision uint64                   // Monotonically increasing revision

ConfigUpdate
├── Name                    string       // Sandbox name (required for sandbox-scoped)
├── Policy                  []byte       // Opaque SandboxPolicy proto bytes
├── SettingKey               string       // Single setting key to mutate
├── SettingValue             *SettingValue // Setting value for upsert
├── DeleteSetting            bool         // Delete the setting key
├── Global                   bool         // Apply at gateway-global scope
├── MergeOperations          []byte       // Raw proto-encoded PolicyMergeOperation list
└── ExpectedResourceVersion  uint64       // Optimistic concurrency (0 = skip check)

UpdateResult
├── Version          uint32  // Assigned policy version
├── PolicyHash       string  // SHA-256 of serialized policy
├── SettingsRevision uint64  // Settings revision for modified scope
└── Deleted          bool    // True when a setting delete removed an existing key
```

## Proto-to-SDK Mapping

| Proto Type | SDK Type | Notes |
|------------|----------|-------|
| `CreateSshSessionResponse` | `SSHSession` | Direct field mapping |
| `RevokeSshSessionResponse` | `bool` (revoked) | Unwrap single field |
| `GetSandboxConfigResponse` | `SandboxConfig` | Policy as opaque bytes, settings as map |
| `GetGatewayConfigResponse` | `GatewayConfig` | Settings as map |
| `UpdateConfigRequest` | `ConfigUpdate` | Policy and merge ops as opaque bytes |
| `UpdateConfigResponse` | `UpdateResult` | Direct field mapping |
| `SettingValue` (proto oneof) | `SettingValue` (struct) | Oneof → typed struct with Type field |
| `EffectiveSetting` | `EffectiveSetting` | Wraps SettingValue + scope |
| `SettingScope` (proto enum) | `SettingScope` (string) | Enum → string constant |
| `PolicySource` (proto enum) | `PolicySource` (string) | Enum → string constant |
| `TcpForwardFrame` | `io.ReadWriteCloser` | Stream wrapper, not a domain type |
| `TcpForwardInit` | Internal only | Constructed by tcpClient.Forward |

## Relationships

```
ClientInterface
  ├── SSH()    → SSHInterface    ──creates──→ SSHSession
  ├── TCP()    → TCPInterface    ──returns──→ io.ReadWriteCloser
  └── Config() → ConfigInterface
                  ├── GetSandbox ──returns──→ SandboxConfig
                  │                            ├── Settings map ──values──→ EffectiveSetting
                  │                            │                             ├── Value → SettingValue
                  │                            │                             └── Scope → SettingScope
                  │                            └── PolicySource
                  ├── GetGateway ──returns──→ GatewayConfig
                  │                            └── Settings map ──values──→ SettingValue
                  └── Update     ──accepts──→ ConfigUpdate
                                 ──returns──→ UpdateResult
```
