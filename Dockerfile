FROM golang:1.22.4-bookworm AS builder

ENV CGO_CFLAGS="-O -D__BLST_PORTABLE__"
ENV CGO_CFLAGS_ALLOW="-O -D__BLST_PORTABLE__"

WORKDIR /workspace
COPY . .

RUN make build


FROM golang:1.22.4-bookworm

WORKDIR /root

RUN apt-get update -y && apt-get install -y ca-certificates jq tree diffutils vim colordiff dnsutils

COPY --from=builder /workspace/build/moca-cmd /usr/bin/moca-cmd
COPY /workspace/deployment/ ./deployment/
COPY /workspace/deployment/testnet/ .moca-cmd/

CMD ["moca-cmd"]