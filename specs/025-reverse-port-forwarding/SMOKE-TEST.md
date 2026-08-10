# Guided Demo Report

**Feature**: Reverse Port Forwarding (ssh -R)
**Date**: 2026-08-08
**Spec**: specs/025-reverse-port-forwarding/spec.md
**Result**: Auto-skipped (no user-observable flows)

---

## Summary

All functional requirements describe internal Go SDK behavior verified by unit tests.
No user-observable demo flows could be synthesized. This is a library-only feature
(no CLI, HTTP server, or UI) with the real gRPC implementation deferred pending
upstream proto support.

## FR Classification

| FR | Classification | Keyword Signal |
|----|---------------|----------------|
| FR-001 | internal-only | interface (code interface) |
| FR-002 | internal-only | constraint |
| FR-003 | internal-only | function |
| FR-004 | internal-only | return value, constraint |
| FR-005 | internal-only | type |
| FR-006 | internal-only | type |
| FR-007 | internal-only | constraint |
| FR-008 | internal-only | return value |
| FR-009 | internal-only | return value |
| FR-010 | internal-only | return value |
| FR-011 | internal-only | constraint |
| FR-012 | internal-only | return value |
