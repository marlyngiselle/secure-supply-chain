# syntax=docker/dockerfile:1.10

# Hardened multi-stage build.
#
# Stage 1 compiles a static, stripped, reproducible binary.
# Stage 2 ships only that binary on distroless static, as a non-root user.
# No shell, no package manager, no libc: nothing for an attacker to live off.
#
# Pin every base image by digest before the first release. Dependabot keeps
# the digests current through the "docker" ecosystem in .github/dependabot.yml.
#   docker buildx imagetools inspect golang:1.25-alpine --format '{{json .Manifest.Digest}}'
#   docker buildx imagetools inspect gcr.io/distroless/static-debian12:nonroot --format '{{json .Manifest.Digest}}'

# Replace with golang:1.25-alpine@sha256:<digest>
FROM golang:1.27-alpine AS build

WORKDIR /src

# Dependencies first, so source edits do not invalidate this layer.
COPY go.mod ./
RUN go mod download && go mod verify

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# -trimpath and a fixed buildid keep the binary reproducible across builders.
RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -buildvcs=false \
      -ldflags="-s -w -buildid= -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}" \
      -o /out/server ./cmd/server

# Replace with gcr.io/distroless/static-debian12:nonroot@sha256:<digest>
# A Chainguard alternative with the same contract is cgr.dev/chainguard/static.
FROM gcr.io/distroless/static-debian12:nonroot

ARG VERSION=dev
ARG COMMIT=unknown

LABEL org.opencontainers.image.title="secure-supply-chain" \
      org.opencontainers.image.description="Demo service built, signed and attested end to end" \
      org.opencontainers.image.source="https://github.com/marlyngiselle/secure-supply-chain" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}"

COPY --from=build --chown=65532:65532 /out/server /server

# Numeric UID so Kubernetes can enforce runAsNonRoot without resolving names.
USER 65532:65532
EXPOSE 8080

ENTRYPOINT ["/server"]
