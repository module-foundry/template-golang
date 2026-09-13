# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
    -o /out/app ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/app /app

EXPOSE 3001
USER nonroot:nonroot

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD ["/app", "-healthcheck"]

ENTRYPOINT ["/app"]
