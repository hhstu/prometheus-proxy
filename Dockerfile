FROM basefly/golang:1.21.4-alpine AS builder
ENV CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct
WORKDIR /build
COPY . .
RUN COMMIT_SHA1=$(git rev-parse --short HEAD || echo "0.0.0") \
    BUILD_TIME=$(date "+%F %T") \
    go build -ldflags="-s -w -X 'github.com/hhstu/prometheus-proxy/utils.Version=$1' -X 'github.com/hhstu/prometheus-proxy/utils.Commit=${COMMIT_SHA1}' -X 'github.com/hhstu/prometheus-proxy/utils.Date=${BUILD_TIME}'"   -o  prometheus-proxy cmd/app.go

FROM ubuntu:22.04
COPY --from=builder  /build/prometheus-proxy /prometheus-proxy
CMD /prometheus-proxy