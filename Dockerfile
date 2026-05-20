# golang:1.25-alpine
FROM golang@sha256:8d22e29d960bc50cd025d93d5b7c7d220b1ee9aa7a239b3c8f55a57e987e8d45 AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/ssh-cv ./cmd/ssh-cv

# Stage to prepare a writable /data with nonroot ownership for the final image.
# alpine:3.20
FROM alpine@sha256:d9e853e87e55526f6b2917df91a2115c36dd7c696a35be12163d44e6e2a4b6bc AS data
RUN mkdir -p /data && chown 65532:65532 /data

# gcr.io/distroless/static-debian12:nonroot
FROM gcr.io/distroless/static-debian12@sha256:d093aa3e30dbadd3efe1310db061a14da60299baff8450a17fe0ccc514a16639
WORKDIR /app
COPY --from=build /out/ssh-cv /app/ssh-cv
# No content is baked in: operators MUST mount their own content.toml at
# /app/content.toml (see compose.yaml / README). Running without it fails
# fast rather than silently serving the example profile.
COPY --from=data --chown=65532:65532 /data /data
ENV SSHCV_HOST=0.0.0.0 \
    SSHCV_PORT=2222 \
    SSHCV_HOST_KEY=/data/host_key \
    SSHCV_CONTENT=/app/content.toml
EXPOSE 2222
USER nonroot:nonroot
ENTRYPOINT ["/app/ssh-cv"]
