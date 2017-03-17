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

// Processor section

type ProcessorRevision struct {
	Version int `json:"version"`
}

type ProcessorPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type ProcessorConfig struct {
	Properties                  map[string]string `json:"properties"`
	AutoTerminatedRelationships []string          `json:"autoTerminatedRelationships"`
}

type ProcessorComponent struct {
	Id            string            `json:"id,omitempty"`
	ParentGroupId string            `json:"parentGroupId"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Position      ProcessorPosition `json:"position"`
	Config        ProcessorConfig   `json:"config"`
}

type Processor struct {
	Revision  ProcessorRevision  `json:"revision"`
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
