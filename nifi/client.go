package nifi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	Config Config
	Client *http.Client
}

func NewClient(config Config) *Client {
	return &Client{
		Config: config,
		Client: &http.Client{},
	}
}

// Common section

type Revision struct {
	Version int `json:"version"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (c *Client) JsonCall(method string, url string, bodyIn interface{}, bodyOut interface{}) (int, error) {
	var requestBody io.Reader = nil
	if bodyIn != nil {
		var buffer = new(bytes.Buffer)
		json.NewEncoder(buffer).Encode(bodyIn)
		requestBody = buffer
	}
	request, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return 0, err
	}

	if bodyIn != nil {
		request.Header.Add("Content-Type", "application/json; charset=utf-8")
	}

	response, err := c.Client.Do(request)
	if err != nil {
		return 0, err
	}
	if response.StatusCode >= 300 {
		return response.StatusCode, fmt.Errorf("The call has failed with the code of %d", response.StatusCode)
	}
	defer response.Body.Close()

	if bodyOut != nil {
		err = json.NewDecoder(response.Body).Decode(bodyOut)
		if err != nil {
			return response.StatusCode, err
		}
	}

	return response.StatusCode, nil
}

// Process Group section

type ProcessGroupComponent struct {
	Id            string   `json:"id,omitempty"`
	ParentGroupId string   `json:"parentGroupId"`
	Name          string   `json:"name"`
	Position      Position `json:"position"`
}

type ProcessGroup struct {
	Revision  Revision              `json:"revision"`
	Component ProcessGroupComponent `json:"component"`
}

func (c *Client) CreateProcessGroup(processGroup *ProcessGroup) error {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s/process-groups",
		c.Config.Host, c.Config.ApiPath, processGroup.Component.ParentGroupId)
	_, err := c.JsonCall("POST", url, processGroup, processGroup)
	return err
}

func (c *Client) GetProcessGroup(processGroupId string) (*ProcessGroup, error) {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s",
		c.Config.Host, c.Config.ApiPath, processGroupId)
	processGroup := ProcessGroup{}
	code, err := c.JsonCall("GET", url, nil, &processGroup)
	if 404 == code {
		return nil, fmt.Errorf("not_found")
	}
	if nil != err {
		return nil, err
	}
	return &processGroup, nil
}

func (c *Client) UpdateProcessGroup(processGroup *ProcessGroup) error {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s",
		c.Config.Host, c.Config.ApiPath, processGroup.Component.Id)
	_, err := c.JsonCall("PUT", url, processGroup, processGroup)
	return err
}

func (c *Client) DeleteProcessGroup(processGroupId string) error {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s",
		c.Config.Host, c.Config.ApiPath, processGroupId)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

// Processor section

type ProcessorRelationship struct {
	Name          string `json:"name"`
	AutoTerminate bool   `json:"autoTerminate"`
}

type ProcessorConfig struct {
	SchedulingStrategy               string `json:"schedulingStrategy"`
	SchedulingPeriod                 string `json:"schedulingPeriod"`
	ConcurrentlySchedulableTaskCount int    `json:"concurrentlySchedulableTaskCount"`

	Properties                  map[string]interface{} `json:"properties"`
	AutoTerminatedRelationships []string               `json:"autoTerminatedRelationships"`
}

type ProcessorComponent struct {
	Id            string                  `json:"id,omitempty"`
	ParentGroupId string                  `json:"parentGroupId"`
	Name          string                  `json:"name"`
	Type          string                  `json:"type"`
	Position      Position                `json:"position"`
	State         string                  `json:"state,omitempty"`
	Config        ProcessorConfig         `json:"config"`
	Relationships []ProcessorRelationship `json:"relationships"`
}

type Processor struct {
	Revision  Revision           `json:"revision"`
	Component ProcessorComponent `json:"component"`
}

func (c *Client) ProcessorCleanupNilProperties(processor *Processor) error {
	for k, v := range processor.Component.Config.Properties {
		if v == nil {
			delete(processor.Component.Config.Properties, k)
		}
	}
	return nil
}

func (c *Client) CreateProcessor(processor *Processor) error {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s/processors",
		c.Config.Host, c.Config.ApiPath, processor.Component.ParentGroupId)
	_, err := c.JsonCall("POST", url, processor, processor)
	c.ProcessorCleanupNilProperties(processor)
	return err
}

func (c *Client) GetProcessor(processorId string) (*Processor, error) {
	url := fmt.Sprintf("http://%s/%s/processors/%s",
		c.Config.Host, c.Config.ApiPath, processorId)
	processor := Processor{}
	code, err := c.JsonCall("GET", url, nil, &processor)
	if 404 == code {
		return nil, fmt.Errorf("not_found")
	}
	if nil != err {
		return nil, err
	}

	c.ProcessorCleanupNilProperties(&processor)

	relationships := []string{}
	for _, v := range processor.Component.Relationships {
		if v.AutoTerminate {
			relationships = append(relationships, v.Name)
		}
	}
	processor.Component.Config.AutoTerminatedRelationships = relationships

	return &processor, nil
}

func (c *Client) UpdateProcessor(processor *Processor) error {
	url := fmt.Sprintf("http://%s/%s/processors/%s",
		c.Config.Host, c.Config.ApiPath, processor.Component.Id)
	_, err := c.JsonCall("PUT", url, processor, processor)
	c.ProcessorCleanupNilProperties(processor)
	return err
}

func (c *Client) DeleteProcessor(processorId string) error {
	url := fmt.Sprintf("http://%s/%s/processors/%s",
		c.Config.Host, c.Config.ApiPath, processorId)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

func (c *Client) SetProcessorState(processor *Processor, state string) error {
	stateUpdate := Processor{
		Revision: Revision{
			Version: processor.Revision.Version,
		},
		Component: ProcessorComponent{
			Id: processor.Component.Id,
			State: state,
		},
	}
	url := fmt.Sprintf("http://%s/%s/processors/%s",
		c.Config.Host, c.Config.ApiPath, processor.Component.Id)
	err, _ := c.JsonCall("PUT", url, stateUpdate, nil)
	if nil == err {
		processor.Component.State = state
	}
	return err
}

func (c *Client) StartProcessor(processor *Processor) error {
	return c.SetProcessorState(processor, "RUNNING")
}

func (c *Client) StopProcessor(processor *Processor) error {
	return c.SetProcessorState(processor, "STOPPED")
}

// Connection section

type ConnectionHand struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

type ConnectionComponent struct {
	Id                    string         `json:"id,omitempty"`
	ParentGroupId         string         `json:"parentGroupId"`
	Source                ConnectionHand `json:"source"`
	Destination           ConnectionHand `json:"destination"`
	SelectedRelationships []string       `json:"selectedRelationships"`
	Bends                 []Position     `json:"bends"`
}

type Connection struct {
	Revision  Revision            `json:"revision"`
	Component ConnectionComponent `json:"component"`
}

func (c *Client) CreateConnection(connection *Connection) error {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s/connections",
		c.Config.Host, c.Config.ApiPath, connection.Component.ParentGroupId)
	_, err := c.JsonCall("POST", url, connection, connection)
	return err
}

func (c *Client) GetConnection(connectionId string) (*Connection, error) {
	url := fmt.Sprintf("http://%s/%s/connections/%s",
		c.Config.Host, c.Config.ApiPath, connectionId)
	connection := Connection{}
	code, err := c.JsonCall("GET", url, nil, &connection)
	if 404 == code {
		return nil, fmt.Errorf("not_found")
	}
	if nil != err {
		return nil, err
	}
	return &connection, nil
}

func (c *Client) UpdateConnection(connection *Connection) error {
	url := fmt.Sprintf("http://%s/%s/connections/%s",
		c.Config.Host, c.Config.ApiPath, connection.Component.Id)
	_, err := c.JsonCall("PUT", url, connection, connection)
	return err
}

func (c *Client) DeleteConnection(connectionId string) error {
	url := fmt.Sprintf("http://%s/%s/connections/%s",
		c.Config.Host, c.Config.ApiPath, connectionId)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}
