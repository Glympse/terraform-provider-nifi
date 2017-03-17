package nifi

import (
	"bytes"
	"encoding/json"
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

func (c *Client) PostCall(url string, bodyIn interface{}, bodyOut interface{}) (error) {
	requestBody := new(bytes.Buffer)
	json.NewEncoder(requestBody).Encode(bodyIn)

	request, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json; charset=utf-8")

	response, err := c.Client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(bodyOut)
	if err != nil {
		return err
	}

	return nil
}

// Processor section

type ProcessorConfig struct {
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
	requestBody := new(bytes.Buffer)
	json.NewEncoder(requestBody).Encode(processor)

	request, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return "", err
	}
	request.Header.Add("Content-Type", "application/json; charset=utf-8")

	response, err := c.Client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	result := Processor{}
	err = json.NewDecoder(response.Body).Decode(&result)
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
	err := c.PostCall(url, connection, &result)
	if err != nil {
		return "", err
	}

	connection.Component.Id = result.Component.Id

	return connection.Component.Id, nil
}
