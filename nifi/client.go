package nifi

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func (c *Client) EncodeObject(obj interface{}) *bytes.Buffer {
	if obj != nil {
		buffer := new(bytes.Buffer)
		json.NewEncoder(buffer).Encode(obj)
		return buffer
	} else {
		return nil
	}
}

func (c *Client) JsonCall(method string, url string, bodyIn interface{}, bodyOut interface{}) error {
	var request *http.Request
	var err error
	if bodyIn != nil {
		requestBody := new(bytes.Buffer)
		json.NewEncoder(requestBody).Encode(bodyIn)
		request, err = http.NewRequest(method, url, requestBody)
	} else {
		request, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return err
	}

	if bodyIn != nil {
		request.Header.Add("Content-Type", "application/json; charset=utf-8")
	}

	response, err := c.Client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode >= 300 {
		return fmt.Errorf("The call has failed with the code of %d", response.StatusCode)
	}
	defer response.Body.Close()

	if bodyOut != nil {
		err = json.NewDecoder(response.Body).Decode(bodyOut)
		if err != nil {
			return err
		}
	}

	return nil
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
	err := c.JsonCall("POST", url, processGroup, processGroup)
	return err
}

func (c *Client) GetProcessGroup(processGroupId string) (*ProcessGroup, error) {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s",
		c.Config.Host, c.Config.ApiPath, processGroupId)
	processGroup := ProcessGroup{}
	err := c.JsonCall("GET", url, nil, &processGroup)
	if err != nil {
		return nil, err
	}
	return &processGroup, nil
}

func (c *Client) UpdateProcessGroup(processGroup *ProcessGroup) error {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s",
		c.Config.Host, c.Config.ApiPath, processGroup.Component.Id)
	err := c.JsonCall("PUT", url, processGroup, processGroup)
	return err
}

func (c *Client) DeleteProcessGroup(processGroupId string) error {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s",
		c.Config.Host, c.Config.ApiPath, processGroupId)
	err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

// Processor section

type ProcessorConfig struct {
	SchedulingStrategy               string `json:"schedulingStrategy"`
	SchedulingPeriod                 string `json:"schedulingPeriod"`
	ConcurrentlySchedulableTaskCount int    `json:"concurrentlySchedulableTaskCount"`

	Properties                  map[string]string `json:"properties"`
	AutoTerminatedRelationships []string          `json:"autoTerminatedRelationships"`
}

type ProcessorComponent struct {
	Id            string          `json:"id,omitempty"`
	ParentGroupId string          `json:"parentGroupId"`
	Name          string          `json:"name"`
	Type          string          `json:"type"`
	Position      Position        `json:"position"`
	Config        ProcessorConfig `json:"config"`
}

type Processor struct {
	Revision  Revision           `json:"revision"`
	Component ProcessorComponent `json:"component"`
}

func (c *Client) CreateProcessor(processor *Processor) (string, error) {
	url := "http://" + c.Config.Host + "/" + c.Config.ApiPath + "/process-groups/" + processor.Component.ParentGroupId + "/processors"

	result := Processor{}
	err := c.JsonCall("POST", url, processor, &result)
	if err != nil {
		return "", err
	}

	processor.Component.Id = result.Component.Id

	return processor.Component.Id, nil
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

func (c *Client) CreateConnection(connection *Connection) (string, error) {
	url := "http://" + c.Config.Host + "/" + c.Config.ApiPath + "/process-groups/" + connection.Component.ParentGroupId + "/connections"

	result := Connection{}
	err := c.JsonCall("POST", url, connection, &result)
	if err != nil {
		return "", err
	}

	connection.Component.Id = result.Component.Id

	return connection.Component.Id, nil
}
