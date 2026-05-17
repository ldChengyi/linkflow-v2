FROM golang:1.25-alpine AS build

ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.org

ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

WORKDIR /src

COPY pkg/go.mod pkg/go.sum ./pkg/
COPY service/device-event-processor/go.mod service/device-event-processor/go.sum ./service/device-event-processor/
WORKDIR /src/service/device-event-processor
RUN go mod download

WORKDIR /src
COPY pkg ./pkg
COPY service/device-event-processor ./service/device-event-processor

WORKDIR /src/service/device-event-processor
RUN CGO_ENABLED=0 go build -o /out/device-event-processor ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY contracts /app/contracts
COPY --from=build /out/device-event-processor /usr/local/bin/linkflow-device-event-processor

ENTRYPOINT ["/usr/local/bin/linkflow-device-event-processor"]
