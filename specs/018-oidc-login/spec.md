# Feature Specification: OIDC Login Package

**Feature Branch**: `018-oidc-login`
**Created**: 2026-07-03
**Status**: Draft
**Input**: Brainstorm 020-oidc-login.md
**Issue**: https://github.com/rhuss/openshell-sdk-go/issues/24

## Clarifications

### Session 2026-07-03

- Q: Where does the OIDC client_id come from for gateway-aware login? → A: Add `oidc_issuer` and `oidc_client_id` fields to the existing gateway `metadata.json`. Single config file keeps gateway setup simple.
- Q: Should the auth code flow include a `state` parameter for CSRF protection? → A: Yes, always generate and validate a cryptographic `state` parameter per OIDC security best practices.
- Q: Should login functions accept `context.Context`? → A: Yes, all public login functions accept `context.Context` as the first parameter for cancellation and timeout propagation.
- Q: Should gateway-aware Login check for existing valid tokens before starting a new interactive flow? → A: Yes, check disk for valid tokens first; only initiate interactive login if tokens are missing or expired.
- Q: What is the lifecycle of the localhost callback server? → A: Each Login call creates its own callback server, scoped to that single login attempt. The server is shut down after receiving the callback or on timeout.

## User Scenarios & Testing

### User Story 1 - Gateway-Aware Interactive Login (Priority: P1)

A developer building a Go CLI tool needs to authenticate against an OpenShell gateway's OIDC provider. They call a single function with the gateway name, the system opens their browser to the provider's login page, and after successful authentication the tokens are persisted to disk for use by subsequent SDK calls.

**Why this priority**: This is the primary use case that eliminates the dependency on the Rust CLI. Without this, Go programs cannot self-authenticate interactively.

**Independent Test**: Can be tested by invoking the login function with a gateway name, completing the browser flow, and verifying that `oidc_token.json` is written to the correct gateway directory with valid tokens.

**Acceptance Scenarios**:

1. **Given** a configured gateway with OIDC metadata, **When** the developer calls `Login(gatewayName)`, **Then** the system discovers the OIDC provider from gateway metadata, opens a browser to the authorization URL, starts a localhost callback server, and exchanges the authorization code for tokens using PKCE.
2. **Given** a successful authorization code exchange, **When** tokens are received, **Then** the token bundle (access_token, refresh_token, expiry, expires_in) is written to `<gateway-dir>/oidc_token.json` in the format compatible with the existing `oidcBundle` schema read by `diskTokenSource` and `gateway.NewClient`.
3. **Given** an empty gateway name, **When** `Login("")` is called, **Then** the active gateway is resolved automatically (same resolution as `gateway.NewClient("")`).
4. **Given** the browser cannot be opened (headless environment), **When** the login flow starts, **Then** the system automatically falls back to the keyboard flow: it displays the authorization URL and prompts the user to paste the authorization code manually.

---

### User Story 2 - Client Credentials for Service Accounts (Priority: P2)

An operator running a Go-based automation pipeline needs to authenticate a service account against an OIDC provider without any user interaction. They provide client ID and client secret, and receive tokens directly.

**Why this priority**: Service account authentication is essential for CI/CD pipelines and operators. It's the second most common authentication need after interactive login.

**Independent Test**: Can be tested by calling the client credentials function with valid credentials against a test OIDC provider and verifying that a valid access token is returned.

**Acceptance Scenarios**:

1. **Given** valid client credentials (client ID and secret), **When** the client credentials function is called, **Then** the system exchanges the credentials for an access token at the provider's token endpoint without any user interaction.
2. **Given** invalid credentials, **When** the client credentials function is called, **Then** the system returns a typed error describing the authentication failure without leaking the client secret in the error message.
3. **Given** a successful token exchange, **When** tokens are returned, **Then** the result is compatible with the standard OAuth2 token type used by the rest of the SDK.

---

### User Story 3 - Device Code Flow for Headless Environments (Priority: P3)

A developer deploying Go-based agents on input-constrained devices or CI runners needs to authenticate via the device code flow (RFC 8628). The system displays a verification URL and user code, the user completes authentication on a separate device, and the system polls until tokens are granted.

**Why this priority**: Device code flow serves a smaller but important audience (CI runners, headless servers, IoT devices). It complements the interactive browser flow for environments where a browser is unavailable.

**Independent Test**: Can be tested by calling the device login function, verifying the verification URL and user code are displayed, simulating user authorization at the provider, and confirming tokens are returned after polling.

**Acceptance Scenarios**:

1. **Given** a provider that supports device authorization, **When** the device login function is called, **Then** the system requests a device code and displays the verification URL and user code to the user.
2. **Given** a pending device authorization, **When** the system polls the token endpoint, **Then** it respects the provider's specified polling interval and handles `authorization_pending`, `slow_down`, and `expired_token` responses.
3. **Given** the user completes authorization on a separate device, **When** the next poll succeeds, **Then** tokens are returned.
4. **Given** a custom display callback is provided, **When** the verification URL and code are ready, **Then** the callback is invoked instead of writing to standard output.

---

### User Story 4 - Standalone OIDC Login Without Gateway (Priority: P3)

A developer building a Go application that authenticates against an OIDC provider not tied to an OpenShell gateway needs to use explicit provider configuration. They specify the issuer URL and client ID directly.

**Why this priority**: Standalone mode makes the OIDC package reusable beyond gateway-specific scenarios, but the gateway-aware flow is the primary use case.

**Independent Test**: Can be tested by calling the login function with explicit issuer URL and client ID, completing the browser flow, and verifying tokens are returned without any gateway metadata lookup.

**Acceptance Scenarios**:

1. **Given** an explicit issuer URL and client ID, **When** the login function is called with these options, **Then** the system fetches the OIDC discovery document from `<issuer>/.well-known/openid-configuration` and performs the authorization code flow.
2. **Given** the in-memory option is enabled, **When** login succeeds, **Then** tokens are returned but not persisted to disk.

---

### Edge Cases

- What happens when the OIDC discovery endpoint is unreachable or returns invalid JSON? The system must return a typed error with the underlying cause.
- What happens when the localhost callback server cannot bind to any port? The system must fall back to the keyboard flow or return an error if keyboard flow is also disabled.
- What happens when the user does not complete the browser flow within the timeout period? The system must clean up the callback server and return a timeout error.
- What happens when the device code expires before the user completes authorization? The system must return an expiration error.
- What happens when the OIDC provider does not support PKCE? The system should proceed without PKCE but log a warning, as PKCE is recommended but not universally required.
- What happens when the gateway metadata does not contain OIDC configuration? The system must return a clear error indicating that OIDC is not configured for this gateway.
- What happens when the token response is missing required fields (e.g., no refresh token)? The system must accept partial token responses (access token only is valid for client credentials).

## Requirements

### Functional Requirements

- **FR-001**: System MUST fetch and parse OIDC discovery documents from `<issuer>/.well-known/openid-configuration` to resolve authorization, token, and device authorization endpoints.
- **FR-002**: System MUST cache OIDC discovery documents in memory for the lifetime of the process to avoid redundant network requests.
- **FR-003**: System MUST implement the Authorization Code flow with PKCE (S256) and a cryptographic `state` parameter (CSRF protection) for interactive browser-based login.
- **FR-004**: System MUST start a localhost HTTP callback server to receive authorization codes, trying ports 8000 and 18000 by default, with an option to specify a custom port.
- **FR-005**: System MUST auto-open the system browser on supported platforms (macOS, Linux, Windows) to the authorization URL.
- **FR-006**: System MUST fall back to a keyboard-based flow (display URL, accept pasted code) when the browser cannot be opened or when explicitly requested.
- **FR-007**: System MUST implement the Device Code flow (RFC 8628), including device code request, user code display, and token endpoint polling at the provider's specified interval.
- **FR-008**: System MUST implement the Client Credentials grant for non-interactive service account authentication.
- **FR-009**: System MUST support gateway-aware login by reading `oidc_issuer` and `oidc_client_id` fields from the gateway's `metadata.json` file.
- **FR-010**: System MUST persist tokens to `<gateway-dir>/oidc_token.json` after gateway-aware login, in the format compatible with the existing `oidcBundle` schema (access_token, refresh_token, expiry as RFC 3339, expires_in as seconds).
- **FR-011**: System MUST support an in-memory-only mode that skips disk persistence for embedded scenarios.
- **FR-012**: System MUST return an `oauth2.TokenSource` (from `golang.org/x/oauth2`) or an `*oauth2.Token` that integrates with the existing `RefreshableToken` mechanism and `diskTokenSource`.
- **FR-013**: System MUST enforce a configurable timeout for interactive flows (default: 2 minutes).
- **FR-014**: System MUST never expose secrets (tokens, client secrets, authorization codes) in error messages or log output.
- **FR-015**: System MUST use functional options for all configuration, following SDK conventions.
- **FR-016**: System MUST support configurable scopes with a sensible default set (openid, profile, email).
- **FR-017**: System MUST accept a display callback function for customizing how device code verification information is presented.
- **FR-018**: All public login functions MUST accept a context parameter as the first argument for cancellation and timeout propagation.
- **FR-019**: Gateway-aware Login MUST check for existing valid tokens on disk before initiating a new interactive flow. If valid tokens exist, they are returned without user interaction.
- **FR-020**: Each interactive login attempt MUST create its own callback server instance, scoped to that single attempt. The server MUST be shut down after receiving the callback or on timeout.

### Key Entities

- **OIDC Discovery Document**: Represents the provider's `.well-known/openid-configuration` response containing endpoint URLs and supported features. Cached per issuer URL.
- **Token Bundle**: The set of tokens returned by the OIDC provider (access_token, optional refresh_token, expiry as RFC 3339, expires_in as seconds). Persisted to disk for gateway-aware flows using the existing `oidcBundle` schema.
- **Login Options**: Configuration for a login attempt, including issuer URL, client ID, callback port, timeout, scopes, flow type preference, and persistence settings.
- **Gateway OIDC Config**: The `oidc_issuer` and `oidc_client_id` fields in the gateway's existing `metadata.json` file, used to discover the OIDC provider for gateway-aware login.

## Success Criteria

### Measurable Outcomes

- **SC-001**: A Go program can complete a full interactive OIDC login (browser flow) against a gateway and use the resulting tokens for SDK operations, without invoking the Rust CLI.
- **SC-002**: Tokens written by the login package are read correctly by the existing `gateway.NewClient` and `diskTokenSource`, confirming full interoperability.
- **SC-003**: The keyboard fallback flow activates automatically in headless environments, allowing login completion without a graphical browser.
- **SC-004**: Service accounts can authenticate via client credentials in under 2 seconds (excluding network latency to the OIDC provider).
- **SC-005**: Device code flow completes successfully on input-constrained environments where no browser or keyboard input is available locally.
- **SC-006**: All login functions return `*oauth2.Token` values compatible with the existing `diskTokenSource` and `RefreshableToken` mechanism, requiring zero changes to existing token refresh and gateway client code.
- **SC-007**: No secrets (tokens, client secrets, authorization codes) appear in any error message returned by the package.

## Assumptions

- The OIDC provider supports the standard `.well-known/openid-configuration` discovery endpoint.
- Gateway `metadata.json` will be extended with `oidc_issuer` and `oidc_client_id` fields to support gateway-aware login discovery.
- The existing `oidc_token.json` schema (access_token, refresh_token, expiry, expires_in) is sufficient for all supported grant types.
- Token refresh using the refresh_token is handled by the existing `RefreshableToken` mechanism, not by this package. This package is responsible for the initial token acquisition only.
- ID token validation (signature verification, audience check) is out of scope for the initial version. The package trusts the token endpoint response.
- The package targets the same platforms as the rest of the SDK (macOS, Linux, Windows).
- The default scopes (openid, profile, email) are appropriate for OpenShell gateway authentication. Custom scopes can be specified via options.
