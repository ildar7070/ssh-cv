FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/i12k ./cmd/i12k

# Stage to prepare a writable /data with nonroot ownership for the final image.
FROM alpine:3.20 AS data
RUN mkdir -p /data && chown 65532:65532 /data

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/i12k /app/i12k
COPY content.toml /app/content.toml
COPY --from=data --chown=65532:65532 /data /data
ENV I12K_HOST=0.0.0.0 \
    I12K_PORT=2222 \
    I12K_HOST_KEY=/data/host_key \
    I12K_CONTENT=/app/content.toml
EXPOSE 2222
USER nonroot:nonroot
ENTRYPOINT ["/app/i12k"]
