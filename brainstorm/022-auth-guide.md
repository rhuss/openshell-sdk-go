# Brainstorm: Authentication Guide

**Date:** 2026-07-03
**Status:** active
**Issue:** https://github.com/rhuss/openshell-sdk-go/issues/28

## Problem Framing

The SDK documentation covers individual API endpoints (16 pages under
`docs/src/api/`) but has no unified entry point for authentication.
A developer asking "how do I authenticate?" has to piece together
information from `api/client.md`, `api/gateway.md`, `api/oidc.md`,
and `api/refresh.md`. The new OIDC login package (PR #26) makes this
gap more visible: there are now four auth modes, three OIDC grant
types, and a full token lifecycle (login, persist, read, refresh) that
no single page explains.

Both new developers (choosing an auth method) and existing users
(adding OIDC to a working setup) need a guide that covers the full
picture.

## Approaches Considered

### A: Single Guide Page (chosen)

One `docs/src/authentication.md` page under the "Guides" section with:
- Decision tree for auth method selection at the top
- Each auth mode documented with when-to-use and code snippets
- OIDC grant types (browser, keyboard, device code, client credentials)
- Full token lifecycle section (login writes to disk, gateway.NewClient
  reads, RefreshableToken handles refresh, expiry triggers re-auth)
- Cross-links to API reference pages for detailed options

- Pros: single entry point, complete story in one place, easy to find
- Cons: could get long if many auth methods are added later

### B: Authentication Section (Multiple Pages)

A new SUMMARY.md section "# Authentication" with sub-pages per method
and a lifecycle page.

- Pros: clean separation, focused pages
- Cons: more navigation, harder to see the full picture

### C: Guide + Cheat Sheet

Full guide plus a one-pager with just the decision table and code
snippets for quick reference.

- Pros: serves both learning and copy-paste use cases
- Cons: two files to maintain, risk of drift

## Decision

**Option A: Single guide page.** At four auth modes and three OIDC
grant types, a single page stays manageable. Can be split into B later
if the auth surface grows significantly (e.g., mTLS, custom providers).

## Key Requirements

- Decision tree at the top mapping scenarios to auth methods
- All four auth modes covered: none, plaintext, cloudflare_jwt, oidc
- All three OIDC grant types: browser/keyboard, device code, client credentials
- Token lifecycle: login to disk, gateway.NewClient reads, RefreshableToken
  handles refresh, expiry and re-auth
- Code snippets for each auth path (complete, runnable examples)
- Cross-links to api/oidc.md, api/gateway.md, api/refresh.md, api/client.md
- SUMMARY.md entry under "Guides" section
- Works for both new developers and existing SDK users

## Open Questions

- Should the guide mention mTLS as "coming soon" or omit it entirely
  until implemented?
- Should the decision tree be a markdown table or a prose flow
  (if/then style)?
