# SLSA level mapping

[SLSA](https://slsa.dev/spec/v1.0/) v1.0 defines the Build track in levels. This table maps each requirement to the artefact in this repository that meets it.

| Requirement | L1 | L2 | L3 | How this repository meets it |
| --- | :---: | :---: | :---: | --- |
| Provenance exists | yes | yes | yes | `slsa-github-generator` container generator and `actions/attest-build-provenance` both emit in-toto provenance for the image digest |
| Hosted build platform | | yes | yes | GitHub-hosted runners (`ubuntu-24.04`). No self-hosted runners |
| Provenance is authentic (signed) | | yes | yes | Signed keyless through Fulcio, recorded in Rekor |
| Provenance is unforgeable | | | yes | Generated in an isolated reusable workflow. The calling job cannot reach the signing material |
| Builds are isolated | | | yes | Ephemeral runner per job, no shared caches between release runs |

The image therefore meets SLSA Build L3.

## Beyond the Build track

| Practice | Status | Evidence |
| --- | --- | --- |
| Pinned dependencies | Met | Actions by commit SHA, base images by digest (see the Dockerfile note), Go toolchain from `go.mod` |
| SBOM per release | Met | SPDX 2.3 and CycloneDX JSON, attested with Cosign |
| Vulnerability gate | Met | Trivy and Grype, fail on fixable CRITICAL |
| Verification at deploy | Met | Kyverno `verifyImages` checks signature, SBOM and provenance |
| Reproducible build | Partial | `-trimpath`, empty build ID and fixed flags. Not yet verified by an independent rebuild |
| Source track (two-party review) | Repository setting | Enforced through branch protection, which is configuration rather than code |

## Two provenance formats, on purpose

| Format | Producer | Verified by | Why keep it |
| --- | --- | --- | --- |
| SLSA v0.2 in a Cosign attestation | `generator_container_slsa3.yml` | `slsa-verifier`, Kyverno | The L3 claim, and what the admission policy checks |
| SLSA v1 in a GitHub attestation | `actions/attest-build-provenance` | `gh attestation verify` | Native to GitHub, visible in the repository's attestations tab |

See [ADR 0003](adr/0003-slsa-provenance-generator.md) for the reasoning.
