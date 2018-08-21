package nifi

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClientProcessGroupCreate(t *testing.T) {
	config := Config{
		Host:    "127.0.0.1:8090",
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
		Host:    "127.0.0.1:8090",
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
			Position: &Position{
				X: 0,
				Y: 0,
			},
			Config: &ProcessorConfig{
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

	err = client.DeleteProcessor(&processor)
	assert.Nil(t, err)
}

func TestClientConnectionCreate(t *testing.T) {
	config := Config{
		Host:    "127.0.0.1:8090",
		ApiPath: "nifi-api",
	}
	client := NewClient(config)

	processor1 := Processor{
		Revision: Revision{
			Version: 0,
		},
		Component: ProcessorComponent{
			ParentGroupId: "root",
			Name:          "generate_flowfile",
			Type:          "org.apache.nifi.processors.standard.GenerateFlowFile",
			Position: &Position{
				X: 0,
				Y: 0,
			},
			Config: &ProcessorConfig{
				SchedulingStrategy:               "TIMER_DRIVEN",
				SchedulingPeriod:                 "0 sec",
				ConcurrentlySchedulableTaskCount: 1,
				Properties: map[string]interface{}{
					"File Size":        "0B",
					"Batch Size":       "1",
					"Data Format":      "Text",
					"Unique FlowFiles": "false",
				},
				AutoTerminatedRelationships: []string{},
			},
		},
	}
	err := client.CreateProcessor(&processor1)
	assert.Nil(t, err)
	assert.NotEmpty(t, processor1.Component.Id)

	processor2 := Processor{
		Revision: Revision{
			Version: 0,
		},
		Component: ProcessorComponent{
			ParentGroupId: "root",
			Name:          "wait",
			Type:          "org.apache.nifi.processors.standard.Wait",
			Position: &Position{
				X: 0,
				Y: 0,
			},
			Config: &ProcessorConfig{
				SchedulingStrategy:               "TIMER_DRIVEN",
				SchedulingPeriod:                 "0 sec",
				ConcurrentlySchedulableTaskCount: 1,
				Properties:                       map[string]interface{}{},
				AutoTerminatedRelationships: []string{
					"success",
				},
			},
		},
	}
	err = client.CreateProcessor(&processor2)
	assert.Nil(t, err)
	assert.NotEmpty(t, processor2.Component.Id)

	connection := Connection{
		Revision: Revision{
			Version: 0,
		},
		Component: ConnectionComponent{
			ParentGroupId: "root",
			Source: ConnectionHand{
				Id:      processor1.Component.Id,
				Type:    "PROCESSOR",
				GroupId: "root",
			},
			Destination: ConnectionHand{
				Id:      processor2.Component.Id,
				Type:    "PROCESSOR",
				GroupId: "root",
			},
			SelectedRelationships: []string{
				"success",
			},
		},
	}
	err = client.CreateConnection(&connection)
	assert.Nil(t, err)
	assert.NotEmpty(t, connection.Component.Id)
}

func TestClientControllerServiceCreate(t *testing.T) {
	config := Config{
		Host:    "127.0.0.1:8090",
		ApiPath: "nifi-api",
	}
	client := NewClient(config)

	processGroup := ProcessGroup{
		Revision: Revision{
			Version: 0,
		},
		Component: ProcessGroupComponent{
			ParentGroupId: "root",
			Name:          "aws_test",
			Position: Position{
				X: 0,
				Y: 0,
			},
		},
	}
	err := client.CreateProcessGroup(&processGroup)
	assert.Nil(t, err)
	assert.NotEmpty(t, processGroup.Component.Id)

	controllerService := ControllerService{
		Revision: Revision{
			Version: 0,
		},
		Component: ControllerServiceComponent{
			ParentGroupId: processGroup.Component.Id,
			Name:          "aws_controller",
			Type:          "org.apache.nifi.processors.aws.credentials.provider.service.AWSCredentialsProviderControllerService",
			State:         "ENABLED",
		},
	}
	err = client.CreateControllerService(&controllerService)
	assert.Nil(t, err)
	assert.NotEmpty(t, controllerService.Component.Id)

	err = client.DisableControllerService(&controllerService)
	assert.Nil(t, err)

	err = client.EnableControllerService(&controllerService)
	assert.Nil(t, err)

	err = client.DisableControllerService(&controllerService)
	assert.Nil(t, err)

	err = client.DeleteControllerService(&controllerService)
	assert.Nil(t, err)

	client.DeleteProcessGroup(&processGroup)
	assert.Nil(t, err)
}

func TestClientReportingTaskCreate(t *testing.T) {
	config := Config{
		Host:    "127.0.0.1:8090",
		ApiPath: "nifi-api",
	}
	client := NewClient(config)

	processGroup := ProcessGroup{
		Revision: Revision{
			Version: 0,
		},
		Component: ProcessGroupComponent{
			ParentGroupId: "root",
			Name:          "aws_test_2",
			Position: Position{
				X: 0,
				Y: 0,
			},
		},
	}
	err := client.CreateProcessGroup(&processGroup)
	time.Sleep(5000 * time.Millisecond)
	assert.Nil(t, err)
	assert.NotEmpty(t, processGroup.Component.Id)

	reportingTask := ReportingTask{
		Revision: Revision{
			Version: 0,
		},
		Component: ReportingTaskComponent{
			ParentGroupId:      processGroup.Component.Id,
			Name:               "aws_reportingtask",
			Type:               "org.apache.nifi.controller.MonitorDiskUsage",
			Comments:           "For testing",
			SchedulingStrategy: "TIMER_DRIVEN",
			SchedulingPeriod:   "5 min",
			Properties: map[string]interface{}{
				"Threshold":          "80%",
				"Directory Location": "/",
			},
		},
	}

	err = client.CreateReportingTask(&reportingTask)
	assert.Nil(t, err)
	assert.NotEmpty(t, reportingTask.Component.Id)

	reportingTask.Component.Name = "aws_reporting_task_mod"
	err = client.UpdateReportingTask(&reportingTask)
	assert.Nil(t, err)
	assert.NotEmpty(t, reportingTask.Component.Id)

	err = client.DeleteReportingTask(&reportingTask)
	assert.Nil(t, err)

	client.DeleteProcessGroup(&processGroup)
	assert.Nil(t, err)
}
