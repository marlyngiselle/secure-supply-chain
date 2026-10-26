#!/usr/bin/env bash
# Verify a released image end to end, the same way the cluster does.
#
#   scripts/verify.sh ghcr.io/marlyngiselle/secure-supply-chain:1.0.0
#   scripts/verify.sh ghcr.io/marlyngiselle/secure-supply-chain@sha256:<digest>
#
# Checks, in order:
#   1. keyless Cosign signature from this repository's release workflow
#   2. signed CycloneDX and SPDX SBOM attestations from the same identity
#   3. SLSA Build L3 provenance with slsa-verifier
#   4. GitHub build provenance attestation with the gh CLI (if installed)
#
# Requires cosign >= 2.4 and slsa-verifier >= 2.6. Exits non-zero on the first failure.
set -euo pipefail

REPO="${REPO:-marlyngiselle/secure-supply-chain}"
ISSUER="https://token.actions.githubusercontent.com"
IDENTITY_REGEXP="^https://github\.com/${REPO//./\\.}/\.github/workflows/release\.yml@refs/tags/v[0-9]+\.[0-9]+\.[0-9]+$"

usage() {
  sed -n '2,13p' "$0" | sed 's/^# \{0,1\}//'
  exit 2
}

need() {
  command -v "$1" >/dev/null 2>&1 || { echo "error: $1 is not installed" >&2; exit 3; }
}

step() { printf '\n==> %s\n' "$*"; }
ok() { printf '    ok: %s\n' "$*"; }

[[ $# -eq 1 ]] || usage
IMAGE="$1"
[[ "$IMAGE" == *:latest ]] && { echo "error: refusing to verify a :latest reference" >&2; exit 2; }

need cosign
need slsa-verifier
need jq

# Resolve the tag to a digest once, then verify only the digest. A tag can
# move between two commands. A digest cannot.
step "Resolving ${IMAGE} to a digest"
if [[ "$IMAGE" == *@sha256:* ]]; then
  REF="$IMAGE"
else
  DIGEST="$(cosign triangulate --type digest "$IMAGE" | sed 's/.*@//')"
  REF="${IMAGE%:*}@${DIGEST}"
fi
ok "$REF"

step "Signature (keyless, release workflow identity)"
cosign verify "$REF" \
  --certificate-identity-regexp "$IDENTITY_REGEXP" \
  --certificate-oidc-issuer "$ISSUER" \
  | jq -r '.[0].optional.Subject // "signed"' | sed 's/^/    signer: /'
ok "signature verified against Rekor"

for type in cyclonedx spdxjson; do
  step "SBOM attestation (${type})"
  cosign verify-attestation "$REF" \
    --type "$type" \
    --certificate-identity-regexp "$IDENTITY_REGEXP" \
    --certificate-oidc-issuer "$ISSUER" >/dev/null
  ok "${type} SBOM attested by the release workflow"
done

step "SLSA provenance (slsa-github-generator)"
slsa-verifier verify-image "$REF" --source-uri "github.com/${REPO}"
ok "built from github.com/${REPO} by a SLSA Build L3 builder"

if command -v gh >/dev/null 2>&1; then
  step "GitHub build provenance attestation"
  gh attestation verify "oci://${REF}" --repo "$REPO" >/dev/null
  ok "attestation verified with gh"
else
  step "GitHub build provenance attestation"
  echo "    skipped: gh CLI not installed"
fi

printf '\nAll checks passed for %s\n' "$REF"
