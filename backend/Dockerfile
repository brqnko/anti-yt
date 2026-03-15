ARG GO_VERSION=1.25
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS base
WORKDIR /src

FROM base AS dev

RUN apk add --no-cache bash libgcc libstdc++ ripgrep curl

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest && \
    go install github.com/air-verse/air@v1.64.0 && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest && \
    go install github.com/pressly/goose/v3/cmd/goose@latest && \
    go install golang.org/x/tools/gopls@latest && \
    go install github.com/k1LoW/tbls@latest

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

FROM base AS build

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x