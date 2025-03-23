# ビルドステージ
FROM public.ecr.aws/docker/library/golang:1.23.4 AS builder
ENV GO111MODULE=on \
    GOPATH=/go \
    GOBIN=/go/bin \
    PATH=/go/bin:$PATH
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
# Install golangci-lint
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.4
# COPY main module
COPY . /app
# Check and Build
RUN make validate && \
    make build-linux

### If use TLS connection in container, add ca-certificates following command.
### > RUN apt-get update && apt-get install -y ca-certificates
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/bin/batch .
COPY config/config.yaml ./config/
CMD ["./batch"] 