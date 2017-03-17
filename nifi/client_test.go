package nifi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClientProcessGroupCreate(t *testing.T) {
	config := Config{
		Host:    "10.0.119.99:3330",
		ApiPath: "nifi-api",
	}
	client := NewClient(config)

	processGroup := ProcessGroup{
		Revision: Revision{
			Version: 0,
		},
		Component: ProcessGroupComponent{
			ParentGroupId: "root",
			Name:          "kafka_to_s3",
			Position: Position{
				X: 0,
				Y: 0,
			},
		},
	}
	client.CreateProcessGroup(&processGroup)

	assert.NotEmpty(t, processGroup.Component.Id)

	processGroup2, err := client.GetProcessGroup(processGroup.Component.Id)

	assert.Equal(t, err, nil)
	assert.NotEmpty(t, processGroup2.Component.Id)
}

func TestClientProcessorCreate(t *testing.T) {
	config := Config{
		Host:    "10.0.119.99:3330",
		ApiPath: "nifi-api",
	}
	client := NewClient(config)

	processor := Processor{
		Revision: Revision{
			Version: 0,
		},
		Component: ProcessorComponent{
			ParentGroupId: "root",
			Name:          "consume_kafka",
			Type:          "org.apache.nifi.processors.kafka.pubsub.ConsumeKafka_0_10",
			Position: Position{
				X: 0,
				Y: 0,
			},
			Config: ProcessorConfig{
				SchedulingStrategy:               "TIMER_DRIVEN",
				SchedulingPeriod:                 "0 sec",
				ConcurrentlySchedulableTaskCount: 1,
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
	client.CreateProcessor(&processor)

	assert.NotEmpty(t, processor.Component.Id)
}
