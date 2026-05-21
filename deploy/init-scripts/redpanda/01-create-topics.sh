#!/bin/sh
set -eu

BROKERS="${KAFKA_BROKERS:-redpanda:9092}"

create_topic_if_missing() {
  topic="$1"
  partitions="$2"
  replicas="$3"

  if rpk -X brokers="$BROKERS" topic list | awk 'NR > 1 {print $1}' | grep -Fxq "$topic"; then
    echo "Kafka topic already exists: $topic"
    return 0
  fi

  echo "Creating Kafka topic: $topic"
  rpk -X brokers="$BROKERS" topic create "$topic" \
    --partitions "$partitions" \
    --replicas "$replicas"
}

create_topic_if_missing "lf.v1.device.events" "1" "1"
create_topic_if_missing "lf.v1.device.state" "1" "1"
