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
			Name:          "generate_flowfile",
			Type:          "org.apache.nifi.processors.standard.GenerateFlowFile",
			Position: Position{
				X: 0,
				Y: 0,
			},
			Config: ProcessorConfig{
				SchedulingStrategy:               "TIMER_DRIVEN",
				SchedulingPeriod:                 "0 sec",
				ConcurrentlySchedulableTaskCount: 1,
				Properties: map[string]interface{}{
					"File Size":        "0B",
					"Batch Size":       "1",
					"Data Format":      "Text",
					"Unique FlowFiles": "false",
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

	err = client.StartProcessor(&processor)
	assert.Nil(t, err)

	err = client.StopProcessor(&processor)
	assert.Nil(t, err)
}
