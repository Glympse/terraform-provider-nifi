variable "nifi_host" {
  description = "NiFi instance where the flow should be created"
}

variable "nifi_root_process_group_id" {
  description = "ID of process group where flow resources should be created"
}

variable "kafka_brokers" {
  description = "List of comma-separated host:port values"
}

variable "kafka_topic" {
  description = "The name of Kafka topic to read messages from"
}

variable "kafka_partitions" {
  description = "A numer of Kafka partitions"
}

variable "kafka_consumer_group_id" {
  description = "ID of consumer group"
}

variable "kafka_offset_reset" {
  description = "Initial offset setting e.g. 'latest'"
}

variable "kafka_max_poll_records" {
  description = "Maximum number of records to read at a time"
}