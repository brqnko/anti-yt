FROM alpine:3.23

ARG TARGETARCH

RUN apk add --no-cache \
    go \
    nodejs npm \
    nginx \
    sudo curl wget bash unzip jq

ENV GOPATH=/root/go
ENV PATH="/usr/local/go/bin:${PATH}"
ENV PATH="${GOPATH}/bin:${PATH}"

RUN wget -O /tmp/tbls.tar.gz \
        https://github.com/k1LoW/tbls/releases/download/v1.93.0/tbls_v1.93.0_linux_${TARGETARCH}.tar.gz \
    && tar -xzf /tmp/tbls.tar.gz -C /usr/local/bin tbls \
    && chmod +x /usr/local/bin/tbls \
    && rm /tmp/tbls.tar.gz

RUN wget -O /tmp/sqlc.zip \
        https://github.com/sqlc-dev/sqlc/releases/download/v1.30.0/sqlc_1.30.0_linux_${TARGETARCH}.zip \
    && unzip /tmp/sqlc.zip -d /usr/local/bin \
    && chmod +x /usr/local/bin/sqlc \
    && rm /tmp/sqlc.zip

RUN wget -O /usr/local/bin/air \
        https://github.com/air-verse/air/releases/download/v1.64.5/air_1.64.5_linux_${TARGETARCH} \
    && chmod +x /usr/local/bin/air

RUN wget -O /tmp/golangci-lint.tar.gz \
        https://github.com/golangci/golangci-lint/releases/download/v2.11.3/golangci-lint-2.11.3-linux-${TARGETARCH}.tar.gz \
    && tar -xzf /tmp/golangci-lint.tar.gz \
        --strip-components=1 \
        -C /usr/local/bin \
        golangci-lint-2.11.3-linux-${TARGETARCH}/golangci-lint \
    && rm /tmp/golangci-lint.tar.gz

RUN GOOSE_ARCH=$([ "$TARGETARCH" = "amd64" ] && echo "x86_64" || echo "$TARGETARCH") \
    && wget -O /usr/local/bin/goose \
        https://github.com/pressly/goose/releases/download/v3.27.0/goose_linux_${GOOSE_ARCH} \
    && chmod +x /usr/local/bin/goose

RUN MOQ_ARCH=$([ "$TARGETARCH" = "amd64" ] && echo "x86_64" || echo "$TARGETARCH") \
    && wget -O /tmp/moq.tar.gz \
        https://github.com/matryer/moq/releases/download/v0.6.0/moq_Linux_${MOQ_ARCH}.tar.gz \
    && tar -xzf /tmp/moq.tar.gz -C /tmp moq \
    && cp /tmp/moq /usr/local/bin/moq \
    && chmod +x /usr/local/bin/moq \
    && rm /tmp/moq.tar.gz /tmp/moq

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0

CMD nginx && sleep infinity
