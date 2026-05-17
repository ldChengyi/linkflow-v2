FROM golang:1.25-alpine AS build

ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.org

ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

WORKDIR /src

COPY pkg/go.mod pkg/go.sum ./pkg/
COPY service/backend/go.mod service/backend/go.sum ./service/backend/
WORKDIR /src/service/backend
RUN go mod download

WORKDIR /src
COPY pkg ./pkg
COPY service/backend ./service/backend

WORKDIR /src/service/backend
RUN CGO_ENABLED=0 go build -o /out/backend ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget

COPY --from=build /out/backend /usr/local/bin/linkflow-backend

EXPOSE 18080

ENTRYPOINT ["/usr/local/bin/linkflow-backend"]
