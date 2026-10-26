# ADR 0003. SLSA provenance from an isolated generator

| Status | Date |
| --- | --- |
| Accepted | 2026-10-26 |

## Context

Build provenance must be unforgeable by the build itself to reach SLSA Build L3. Provenance generated in the same job that runs the build can be influenced by any step in that job, including a compromised action.

## Decision

Generate the authoritative provenance with `slsa-framework/slsa-github-generator` (`generator_container_slsa3.yml`), a reusable workflow that runs in its own job and signs with its own identity. Also emit a GitHub attestation with `actions/attest-build-provenance`, which is convenient to inspect but produced from within the calling workflow.

The generator is referenced by tag (`@v2.1.0`) rather than by SHA. This is the one deliberate exception to SHA pinning. The generator verifies its own ref at runtime to prove which builder ran, and a SHA reference fails that check.

## Consequences

- The admission policy trusts the generator's identity for provenance, not the release workflow's. A compromised release job cannot forge it.
- Scorecard flags the tag reference under `Pinned-Dependencies`. That finding is expected and documented here.
- Two provenance formats coexist. `slsa-verifier` and Kyverno use the SLSA one. `gh attestation verify` uses the GitHub one.
