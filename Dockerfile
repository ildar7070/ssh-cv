FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/ssh-cv ./cmd/ssh-cv

# Stage to prepare a writable /data with nonroot ownership for the final image.
FROM alpine:3.20 AS data
RUN mkdir -p /data && chown 65532:65532 /data

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/ssh-cv /app/ssh-cv
# Ship the example content as the default. Operators override by mounting
# their own file at /app/content.toml (see compose.yaml / README).
COPY content.example.toml /app/content.toml
COPY --from=data --chown=65532:65532 /data /data
ENV SSHCV_HOST=0.0.0.0 \
    SSHCV_PORT=2222 \
    SSHCV_HOST_KEY=/data/host_key \
    SSHCV_CONTENT=/app/content.toml
EXPOSE 2222
USER nonroot:nonroot
ENTRYPOINT ["/app/ssh-cv"]
