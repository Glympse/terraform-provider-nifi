provider "nifi" {
  host = "${var.nifi_host}"
}

resource "nifi_process_group" "kafka_to_mongo" {
  component {
    parent_group_id = "${var.nifi_root_process_group_id}"
    name = "kafka_to_mongo"

    position {
      x = 0
      y = 0
    }
  }
}

resource "nifi_processor" "consume_kafka" {
  component {
    parent_group_id = "${nifi_process_group.kafka_to_mongo.id}"
    name = "consume_kafka"
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
        "auto.offset.reset" = "latest"
        "key-attribute-encoding" = "utf-8"
        "max.poll.records" = "10000"
        "max-uncommit-offset-wait" = "1 secs"
        "message-demarcator" = "\n"
      }

      auto_terminated_relationships = []
    }
  }
}

resource "nifi_processor" "put_mongo" {
  component {
    parent_group_id = "${nifi_process_group.kafka_to_mongo.id}"
    name = "put_mongo"
    type = "org.apache.nifi.processors.mongodb.PutMongo"

    position {
      x = 0
      y = 250
    }

    config {
      properties {
        "Mongo URI" = "${var.mongo_uri}"
        "Mongo Database Name" = "${var.mongo_database_name}"
        "Mongo Collection Name" = "${var.mongo_collection_name}"
        "ssl-client-auth" = "REQUIRED"
        "Mode" = "insert"
        "Upsert" = "false"
        "Update Query Key" = "_id"
        "Write Concern" = "ACKNOWLEDGED"
        "Character Set" = "UTF-8"
      }

      auto_terminated_relationships = [
        "success", "failure"
      ]
    }
  }
}

resource "nifi_connection" "kafka_to_mongo_conn" {
  component {
    parent_group_id = "${nifi_process_group.kafka_to_mongo.id}"

    source {
      type = "PROCESSOR"
      id = "${nifi_processor.consume_kafka.id}"
    }

    destination {
      type = "PROCESSOR"
      id = "${nifi_processor.put_mongo.id}"
    }

    selected_relationships = [
      "success"
    ]
  }
}
