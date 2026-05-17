FROM golang:1.25-alpine AS build

ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.org

ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

WORKDIR /src

COPY pkg/go.mod pkg/go.sum ./pkg/
COPY service/mqtt-gateway/go.mod service/mqtt-gateway/go.sum ./service/mqtt-gateway/
WORKDIR /src/service/mqtt-gateway
RUN go mod download

WORKDIR /src
COPY pkg ./pkg
COPY service/mqtt-gateway ./service/mqtt-gateway

WORKDIR /src/service/mqtt-gateway
RUN CGO_ENABLED=0 go build -o /out/mqtt-gateway ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=build /out/mqtt-gateway /usr/local/bin/linkflow-mqtt-gateway

ENTRYPOINT ["/usr/local/bin/linkflow-mqtt-gateway"]
