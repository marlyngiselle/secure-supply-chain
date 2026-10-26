# Verify an image yourself

You do not need to trust this README. Every claim about a released image can be checked from a laptop with public tools and no credentials.

## Install the tools

```bash
brew install cosign slsa-verifier jq gh
```

## One command

```bash
scripts/verify.sh ghcr.io/marlyngiselle/secure-supply-chain:1.0.0
```

The script resolves the tag to a digest once and verifies only that digest.

## Step by step

Resolve the digest first. Every later command uses it, never the tag.

```bash
IMAGE=ghcr.io/marlyngiselle/secure-supply-chain
REF="$IMAGE@$(cosign triangulate --type digest "$IMAGE:1.0.0" | sed 's/.*@//')"
echo "$REF"
```

Check the signature and who made it.

```bash
cosign verify "$REF" \
  --certificate-identity-regexp '^https://github\.com/marlyngiselle/secure-supply-chain/\.github/workflows/release\.yml@refs/tags/v' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com | jq .
```

Read the SBOM straight from the attestation.

```bash
cosign verify-attestation "$REF" --type cyclonedx \
  --certificate-identity-regexp '^https://github\.com/marlyngiselle/secure-supply-chain/\.github/workflows/release\.yml@refs/tags/v' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  | jq -r '.payload' | base64 -d | jq '.predicate.components[] | {name, version}'
```

Check the SLSA provenance.

```bash
slsa-verifier verify-image "$REF" \
  --source-uri github.com/marlyngiselle/secure-supply-chain \
  --print-provenance | jq '.predicate.invocation.configSource'
```

Check the GitHub build attestation.

```bash
gh attestation verify "oci://$REF" --repo marlyngiselle/secure-supply-chain
```

Re-scan today's vulnerability data against the attested SBOM, without pulling the image.

```bash
cosign verify-attestation "$REF" --type cyclonedx \
  --certificate-identity-regexp '^https://github\.com/marlyngiselle/secure-supply-chain/' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  | jq -r '.payload' | base64 -d | jq '.predicate' > sbom.cdx.json
grype sbom:sbom.cdx.json --fail-on critical
```

## What a failure means

| Failure | Likely cause | What to do |
| --- | --- | --- |
| `no matching signatures` | Image not built by the release workflow, or built from a branch instead of a tag | Treat as untrusted. Do not deploy |
| `none of the expected identities matched` | Signed by another repository or workflow file | Check the regexp. If it is right, the image is not ours |
| `no matching attestations` for `cyclonedx` | Release stopped between signing and attesting | Re-run the release for that tag |
| `slsa-verifier` source mismatch | Provenance names another repository | Treat as untrusted |
