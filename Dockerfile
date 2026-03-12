FROM golang:1.23.6-bookworm AS builder

ARG GITHUB_TOKEN
RUN git config --global url."https://${GITHUB_TOKEN}:@github.com/".insteadOf "https://github.com/"

ENV CGO_CFLAGS="-O -D__BLST_PORTABLE__"
ENV CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"
ENV GOPRIVATE=github.com/mocachain
ENV GOPROXY=https://proxy.golang.org,direct
ENV GONOSUMDB=github.com/mocachain/*
ENV GONOSUMCHECK=github.com/mocachain/*

WORKDIR /workspace

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make build


FROM golang:1.23.6-bookworm

WORKDIR /root

RUN apt-get update -y && apt-get install -y ca-certificates jq tree diffutils vim colordiff dnsutils

COPY --from=builder /workspace/build/moca-cmd /usr/bin/moca-cmd

CMD ["moca-cmd"]