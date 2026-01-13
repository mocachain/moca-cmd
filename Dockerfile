FROM golang:1.23.6-bookworm AS builder

ARG GITHUB_TOKEN
RUN if [ -n "${GITHUB_TOKEN}" ]; then \
        git config --global url."https://${GITHUB_TOKEN}:@github.com/".insteadOf "https://github.com/"; \
    fi

ENV CGO_CFLAGS="-O -D__BLST_PORTABLE__"
ENV CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"
ENV GOPRIVATE="github.com/mocachain/*,github.com/evmos/*"
ENV GONOPROXY="github.com/mocachain/*,github.com/evmos/*"
ENV GONOSUMDB="github.com/mocachain/*,github.com/evmos/*"
ENV GOSUMDB=off

WORKDIR /workspace
COPY go.mod go.sum ./
RUN rm -f go.sum && go mod download && go mod tidy
COPY . .
RUN rm -f go.sum && go mod download && go mod tidy && make build


FROM golang:1.23.6-bookworm

WORKDIR /root

RUN apt-get update -y && apt-get install -y ca-certificates jq tree diffutils vim colordiff dnsutils

COPY --from=builder /workspace/build/moca-cmd /usr/bin/moca-cmd

CMD ["moca-cmd"]