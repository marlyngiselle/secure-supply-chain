# ADR 0002. Kyverno as the admission controller

| Status | Date |
| --- | --- |
| Accepted | 2026-10-26 |

## Context

The cluster must verify signatures and attestations at admission, and enforce runtime hardening. Candidates were Kyverno, the Sigstore policy-controller and OPA Gatekeeper.

| Criterion | Kyverno | Sigstore policy-controller | Gatekeeper |
| --- | --- | --- | --- |
| Keyless signature verification | Native (`verifyImages`) | Native | Not native, needs external data provider |
| Attestation conditions (SBOM, SLSA) | Native, JMESPath conditions | Native, CUE or Rego | Not native |
| Pod hardening rules | Native | No | Native (Rego) |
| Digest mutation after verification | Yes (`mutateDigest`) | No | No |
| Policy language | YAML | YAML plus CUE or Rego | Rego |

## Decision

Use Kyverno for both supply chain verification and pod hardening. Ship a Gatekeeper constraint for the runtime rules as an alternative for clusters already standardised on OPA, and rely on Pod Security Admission `restricted` as an independent second layer.

## Consequences

- One engine and one language for every rule a reviewer needs to read.
- Digest mutation closes the gap between verifying a tag and pulling it.
- Kyverno becomes critical path for pod creation. It must run with several replicas and a PodDisruptionBudget, and the webhook failure policy must be decided explicitly per environment.
