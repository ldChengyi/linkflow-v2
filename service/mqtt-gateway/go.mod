module github.com/ldchengyi/linkflow-v2/service/mqtt-gateway

go 1.25.0

require (
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/ldchengyi/linkflow-v2/pkg v0.0.0
	github.com/segmentio/kafka-go v0.4.51
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
)

replace github.com/ldchengyi/linkflow-v2/pkg => ../../pkg
