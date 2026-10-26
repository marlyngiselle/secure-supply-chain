# ADR 0001. Keyless signing with Sigstore

| Status | Date |
| --- | --- |
| Accepted | 2026-10-26 |

## Context

Images must be signed so that the cluster can reject anything the pipeline did not produce. A classic key pair means a secret to store, rotate, scope and eventually revoke after a leak. Leaked signing keys are a recurring root cause in supply chain incidents.

## Decision

Sign with Cosign in keyless mode. The release job requests a GitHub OIDC token (`id-token: write`), exchanges it with Fulcio for a certificate valid for ten minutes, signs the image digest and records the signature in the Rekor transparency log. The certificate embeds the workflow path and the git ref, which becomes the identity the admission policy checks.

## Consequences

- No signing secret exists anywhere, so none can leak or need rotation.
- Identity is precise. A signature from another workflow, another repository or a branch build does not match the policy regex.
- Every signature is public in Rekor. That is acceptable for an open-source image and would need a private Sigstore instance for confidential workloads.
- Verification depends on the availability of the public Sigstore infrastructure at admission time. Kyverno caches results, and the webhook timeout is set to 30 seconds.
