provider "nifi" {
  host = "${var.nifi_host}"
}

resource "nifi_processor" "consume_kafka" {
  component {
    parent_group_id = "${var.nifi_root_process_group_id}"
    name = "consume-kafka"
    type = "org.apache.nifi.processors.kafka.pubsub.ConsumeKafka_0_10"

    position {
      x = 0
      y = 0
    }

    config {
      properties {
        "bootstrap.servers" = "${var.kafka_brokers}"
        "security.protocol" = "PLAINTEXT"
        "topic" = "${var.kafka_topic}"
        "group.id" = "${var.kafka_consumer_group_id}"
        "auto.offset.reset" = "${var.kafka_offset_reset}"
        "key-attribute-encoding" = "utf-8"
        "max.poll.records" = "${var.kafka_max_poll_records}"
      }

      auto_terminated_relationships = [
        "success"
      ]
    }
  }
}
