package nifi

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestClientProcessorCreate(t *testing.T) {
	config := Config{
		Host:               "10.0.119.99:3330",
		ApiPath:            "nifi-api",
		RootProcessGroupId: "root",
	}
	client := NewClient(config)

	processor := Processor{
		Revision: ProcessorRevision{
			Version: 0,
		},
		Component: ProcessorComponent{
			Name: "consume-kafka",
			Type: "org.apache.nifi.processors.kafka.pubsub.ConsumeKafka_0_10",
			Position: ProcessorPosition{
				X: 0,
				Y: 0,
			},
			Config: ProcessorConfig{
				Properties: map[string]string{
					"security.protocol":      "PLAINTEXT",
					"topic":                  "cards-core-api",
					"group.id":               "nifi-api-streamer",
					"auto.offset.reset":      "latest",
					"key-attribute-encoding": "utf-8",
					"message-demarcator":     "\n",
					"max.poll.records":       "10000",
				},
				AutoTerminatedRelationships: []string{
					"success",
				},
			},
		},
	}
	client.CreateProcessor("root", &processor)

	assert.NotEmpty(t, processor.Component.Id)
}
