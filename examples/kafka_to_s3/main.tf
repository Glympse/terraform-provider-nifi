provider "nifi" {
  host = "${var.nifi_host}"
}

resource "nifi_processor" "consume_kafka" {
  component {
    parent_group_id = "${var.nifi_root_process_group_id}"
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
        "auto.offset.reset" = "${var.kafka_offset_reset}"
        "key-attribute-encoding" = "utf-8"
        "max.poll.records" = "${var.kafka_max_poll_records}"
      }

      auto_terminated_relationships = []
    }
  }
}

resource "nifi_processor" "merge_content" {
  component {
    parent_group_id = "${var.nifi_root_process_group_id}"
    name = "merge_content"
    type = "org.apache.nifi.processors.standard.MergeContent"

    position {
      x = 0
      y = 250
    }

    config {
      properties {
        "Merge Strategy" = "Bin-Packing Algorithm"
        "Merge Format" = "Binary Concatenation"
        "Attribute Strategy" = "Keep Only Common Attributes"
        "Minimum Number of Entries" = "1"
        "Minimum Group Size" = "10 MB"
        "Maximum Group Size" = "12 MB"
        "Max Bin Age" = "1 hour"
        "Maximum number of Bins" = "1000000"
        "Delimiter Strategy" = "Text"
        "Demarcator File" = "\n"
        "Compression Level" = "1"
        "Keep Path" = "false"
      }

      auto_terminated_relationships = [
        "failure", "original"
      ]
    }
  }
}

resource "nifi_connection" "kafka_to_merge" {
  "component" {
    parent_group_id = "${var.nifi_root_process_group_id}"

    source {
      type = "PROCESSOR"
      id = "${nifi_processor.consume_kafka.id}"
    }

    destination {
      type = "PROCESSOR"
      id = "${nifi_processor.merge_content.id}"
    }

    selected_relationships = [
      "success"
    ]
  }
}

resource "nifi_processor" "put_s3_object" {
  component {
    parent_group_id = "${var.nifi_root_process_group_id}"
    name = "put_s3_object"
    type = "org.apache.nifi.processors.aws.s3.PutS3Object"

    position {
      x = 0
      y = 500
    }

    config {
      properties {
        "Object Key" = "exp/nifi/$${now():format('yyyy')}/$${now():format('MM')}/$${now():format('dd')}/$${now():format('HH-mm-ss-SSS')}-$${UUID()}.njson"
        "Bucket" = "s-reporting-tmp"
        "Storage Class" = "Standard"
        "Region" = "us-east-1"
        "Communications Timeout" = "30 secs"
        "Multipart Threshold" = "5 GB"
        "Multipart Part Size" = "5 GB"
        "Multipart Upload AgeOff Interval" = "60 min"
        "Multipart Upload Max Age Threshold" = "7 days"
        "server-side-encryption" = "None"
      }

      auto_terminated_relationships = [
        "success"
      ]
    }
  }
}

resource "nifi_connection" "merge_to_s3" {
  component {
    parent_group_id = "${var.nifi_root_process_group_id}"

    source {
      type = "PROCESSOR"
      id = "${nifi_processor.merge_content.id}"
    }

    destination {
      type = "PROCESSOR"
      id = "${nifi_processor.put_s3_object.id}"
    }

    selected_relationships = [
      "merged"
    ]
  }
}

resource "nifi_connection" "s3_to_s3" {
  component {
    parent_group_id = "${var.nifi_root_process_group_id}"

    source {
      type = "PROCESSOR"
      id = "${nifi_processor.put_s3_object.id}"
    }

    destination {
      type = "PROCESSOR"
      id = "${nifi_processor.put_s3_object.id}"
    }

    selected_relationships = [
      "failure"
    ]

    bends = [{
      x = 460
      y = 540
    }, {
      x = 460
      y = 600
    }]
  }
}
