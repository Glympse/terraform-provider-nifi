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

variable "kafka_consumer_group_id" {
  description = "ID of consumer group"
}

variable "mongo_uri" {
  description = "Mongo URI to connect to"
}

variable "mongo_database_name" {
  description = "Mongo database name"
}

variable "mongo_collection_name" {
  description = "Mongo collection name"
}
