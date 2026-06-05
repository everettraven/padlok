# padlok Documentation

## Introduction

- [Concepts](concepts.md) — Key terminology and mental model.
- [Architecture](architecture.md) — How `padlok` fits into the Kubernetes authentication flow.

## Getting Started

- [Installation](installation.md) — Building, deploying, and connecting `padlok` to a Kubernetes cluster.
- [Quickstart](quickstart.md) — A minimal end-to-end example.

## Configuration

- [AuthenticationConfiguration](authentication-configuration.md) — The configuration file format, hot-reload, and validation.
- [Configuring OIDC Issuers](configuring-oidc-issuers.md) — Issuer URL, discovery, audiences, and TLS.
- [Claim Mappings](claim-mappings.md) — Mapping token claims to Kubernetes identity attributes.
- [Validation Rules](validation-rules.md) — Claim and user validation with CEL expressions.
- [External Claims Sources](external-claims-sources.md) — Fetching additional claims from external HTTP endpoints.
- [CEL Expressions](cel-expressions.md) — The CEL expression environment, variables, and common patterns.

## Operations

- [TLS Configuration](tls-configuration.md) — Server TLS, OIDC issuer TLS, and external source TLS.
- [Troubleshooting](troubleshooting.md) — Diagnosing authentication issues.
- [Running on OpenShift](running-on-openshift.md) — Guide for configuring OpenShift to delegate authentication decisions to `padlok`.

## Reference

- [CLI Reference](cli-reference.md) — `padlok run` flags.
- [Configuration Schema](configuration-schema.md) — Full field-level schema for `AuthenticationConfiguration`.
