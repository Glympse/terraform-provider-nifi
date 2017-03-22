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

	processGroup.Component.Name = "kafka_to_s3_1"
	err = client.UpdateProcessGroup(&processGroup)
	assert.Equal(t, err, nil)
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
				Properties: map[string]interface{}{
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
	err := client.CreateProcessor(&processor)
	assert.Nil(t, err)
	assert.NotEmpty(t, processor.Component.Id)

	processor.Component.Config.AutoTerminatedRelationships = []string{}
	err = client.UpdateProcessor(&processor)
	assert.Nil(t, err)

	processor.Component.Config.AutoTerminatedRelationships = []string{
		"success",
	}
	err = client.UpdateProcessor(&processor)
	assert.Nil(t, err)
}
