package nifi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	Config Config
	Client *http.Client

	// The mutex is used by the plugin to prevent parallel execution of some update/delete operations.
	// There are scenarios when updating a connection involves modifying related processors and vice versa.
	// This breaks Terraform model to some extent but at the same time is unavoidable in NiFi world.
	// Currently only flows that involve cross-resource interactions are wrapped into lock/unlock sections.
	// Most of operations can still be performed in parallel.
	Lock sync.Mutex
}

func NewClient(config Config) *Client {
	http_client := &http.Client{}
	if config.AdminCertPath != "" && config.AdminKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(config.AdminCertPath, config.AdminKeyPath)
		if err != nil {
			log.Fatal(err)
		} else {
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			tlsConfig.BuildNameToCertificate()
			transport := &http.Transport{TLSClientConfig: tlsConfig}
			http_client = &http.Client{Transport: transport}
		}
	}
	client := &Client{
		Config: config,
		Client: http_client,
	}
	return client
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

func (c *Client) DeleteProcessGroup(processGroup *ProcessGroup) error {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s?version=%d",
		c.Config.Host, c.Config.ApiPath, processGroup.Component.Id, processGroup.Revision.Version)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

func (c *Client) GetProcessGroupConnections(processGroupId string) (*Connections, error) {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s/connections",
		c.Config.Host, c.Config.ApiPath, processGroupId)
	connections := Connections{}
	_, err := c.JsonCall("GET", url, nil, &connections)
	if nil != err {
		return nil, err
	}
	return &connections, nil
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
	ParentGroupId string                  `json:"parentGroupId,omitempty"`
	Name          string                  `json:"name,omitempty"`
	Type          string                  `json:"type,omitempty"`
	Position      *Position               `json:"position,omitempty"`
	State         string                  `json:"state,omitempty"`
	Config        *ProcessorConfig        `json:"config,omitempty"`
	Relationships []ProcessorRelationship `json:"relationships,omitempty"`
}

type Processor struct {
	Revision  Revision           `json:"revision"`
	Component ProcessorComponent `json:"component"`
}

func ProcessorStub() *Processor {
	return &Processor{
		Component: ProcessorComponent{
			Position: &Position{},
			Config:   &ProcessorConfig{},
		},
	}
}

func (c *Client) CleanupNilProperties(properties map[string]interface{}) error {
	for k, v := range properties {
		if v == nil {
			delete(properties, k)
		}
	}
	return nil
}

func (c *Client) CreateProcessor(processor *Processor) error {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s/processors",
		c.Config.Host, c.Config.ApiPath, processor.Component.ParentGroupId)
	_, err := c.JsonCall("POST", url, processor, processor)
	if nil != err {
		return err
	}
	c.CleanupNilProperties(processor.Component.Config.Properties)
	return nil
}

func (c *Client) GetProcessor(processorId string) (*Processor, error) {
	url := fmt.Sprintf("http://%s/%s/processors/%s",
		c.Config.Host, c.Config.ApiPath, processorId)
	processor := ProcessorStub()
	code, err := c.JsonCall("GET", url, nil, &processor)
	if 404 == code {
		return nil, fmt.Errorf("not_found")
	}
	if nil != err {
		return nil, err
	}

	c.CleanupNilProperties(processor.Component.Config.Properties)

	relationships := []string{}
	for _, v := range processor.Component.Relationships {
		if v.AutoTerminate {
			relationships = append(relationships, v.Name)
		}
	}
	processor.Component.Config.AutoTerminatedRelationships = relationships

	return processor, nil
}

func (c *Client) UpdateProcessor(processor *Processor) error {
	url := fmt.Sprintf("http://%s/%s/processors/%s",
		c.Config.Host, c.Config.ApiPath, processor.Component.Id)
	_, err := c.JsonCall("PUT", url, processor, processor)
	if nil != err {
		return err
	}
	c.CleanupNilProperties(processor.Component.Config.Properties)
	return nil
}

func (c *Client) DeleteProcessor(processor *Processor) error {
	url := fmt.Sprintf("http://%s/%s/processors/%s?version=%d",
		c.Config.Host, c.Config.ApiPath, processor.Component.Id, processor.Revision.Version)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

func (c *Client) SetProcessorState(processor *Processor, state string) error {
	stateUpdate := Processor{
		Revision: Revision{
			Version: processor.Revision.Version,
		},
		Component: ProcessorComponent{
			Id:    processor.Component.Id,
			State: state,
		},
	}
	url := fmt.Sprintf("http://%s/%s/processors/%s",
		c.Config.Host, c.Config.ApiPath, processor.Component.Id)
	_, err := c.JsonCall("PUT", url, stateUpdate, processor)
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
	Type    string `json:"type"`
	Id      string `json:"id"`
	GroupId string `json:"groupId"`
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

type Connections struct {
	Connections []Connection `json:"connections"`
}

type ConnectionDropRequest struct {
	DropRequest struct {
		Id       string `json:"id"`
		Finished bool   `json:"finished"`
	} `json:"dropRequest"`
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

func (c *Client) DeleteConnection(connection *Connection) error {
	url := fmt.Sprintf("http://%s/%s/connections/%s?version=%d",
		c.Config.Host, c.Config.ApiPath, connection.Component.Id, connection.Revision.Version)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

func (c *Client) DropConnectionData(connection *Connection) error {
	// Create a request to drop the contents of the queue in this connection
	url := fmt.Sprintf("http://%s/%s/flowfile-queues/%s/drop-requests",
		c.Config.Host, c.Config.ApiPath, connection.Component.Id)
	dropRequest := ConnectionDropRequest{}
	_, err := c.JsonCall("POST", url, nil, &dropRequest)
	if nil != err {
		return err
	}

	// Give it some time to complete
	maxAttempts := 10
	for iteration := 0; iteration < maxAttempts; iteration++ {
		// Check status of the request
		url = fmt.Sprintf("http://%s/%s/flowfile-queues/%s/drop-requests/%s",
			c.Config.Host, c.Config.ApiPath, connection.Component.Id, dropRequest.DropRequest.Id)
		_, err = c.JsonCall("GET", url, nil, &dropRequest)
		if nil != err {
			continue
		}
		if dropRequest.DropRequest.Finished {
			break
		}

		// Log progress
		log.Printf("[INFO] Purging Connection data %s %d...", dropRequest.DropRequest.Id, iteration+1)

		// Wait a bit
		time.Sleep(3 * time.Second)

		if maxAttempts-1 == iteration {
			log.Printf("[INFO] Failed to purge the Connection %s", dropRequest.DropRequest.Id)
		}
	}

	// Remove a request to drop the contents of this connection
	url = fmt.Sprintf("http://%s/%s/flowfile-queues/%s/drop-requests/%s",
		c.Config.Host, c.Config.ApiPath, connection.Component.Id, dropRequest.DropRequest.Id)
	_, err = c.JsonCall("DELETE", url, nil, nil)
	if nil != err {
		return err
	}

	return nil
}

// Controller Service section

type ControllerServiceComponent struct {
	Id            string                 `json:"id,omitempty"`
	ParentGroupId string                 `json:"parentGroupId,omitempty"`
	Name          string                 `json:"name,omitempty"`
	Type          string                 `json:"type,omitempty"`
	State         string                 `json:"state,omitempty"`
	Properties    map[string]interface{} `json:"properties"`
}

type ControllerService struct {
	Revision  Revision                   `json:"revision"`
	Component ControllerServiceComponent `json:"component"`
}

func (c *Client) CreateControllerService(controllerService *ControllerService) error {
	url := fmt.Sprintf("http://%s/%s/process-groups/%s/controller-services",
		c.Config.Host, c.Config.ApiPath, controllerService.Component.ParentGroupId)
	_, err := c.JsonCall("POST", url, controllerService, controllerService)
	if nil != err {
		return err
	}
	c.CleanupNilProperties(controllerService.Component.Properties)
	return nil
}

func (c *Client) GetControllerService(controllerServiceId string) (*ControllerService, error) {
	url := fmt.Sprintf("http://%s/%s/controller-services/%s",
		c.Config.Host, c.Config.ApiPath, controllerServiceId)
	controllerService := ControllerService{}
	code, err := c.JsonCall("GET", url, nil, &controllerService)
	if 404 == code {
		return nil, fmt.Errorf("not_found")
	}
	if nil != err {
		return nil, err
	}
	c.CleanupNilProperties(controllerService.Component.Properties)
	return &controllerService, nil
}

func (c *Client) UpdateControllerService(controllerService *ControllerService) error {
	url := fmt.Sprintf("http://%s/%s/controller-services/%s",
		c.Config.Host, c.Config.ApiPath, controllerService.Component.Id)
	_, err := c.JsonCall("PUT", url, controllerService, controllerService)
	if nil != err {
		return err
	}
	c.CleanupNilProperties(controllerService.Component.Properties)
	return nil
}

func (c *Client) DeleteControllerService(controllerService *ControllerService) error {
	url := fmt.Sprintf("http://%s/%s/controller-services/%s?version=%d",
		c.Config.Host, c.Config.ApiPath, controllerService.Component.Id, controllerService.Revision.Version)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

func (c *Client) SetControllerServiceState(controllerService *ControllerService, state string) error {
	stateUpdate := ControllerService{
		Revision: Revision{
			Version: controllerService.Revision.Version,
		},
		Component: ControllerServiceComponent{
			Id:    controllerService.Component.Id,
			State: state,
		},
	}
	url := fmt.Sprintf("http://%s/%s/controller-services/%s",
		c.Config.Host, c.Config.ApiPath, controllerService.Component.Id)
	_, err := c.JsonCall("PUT", url, stateUpdate, controllerService)
	return err
}

func (c *Client) EnableControllerService(controllerService *ControllerService) error {
	return c.SetControllerServiceState(controllerService, "ENABLED")
}

func (c *Client) DisableControllerService(controllerService *ControllerService) error {
	return c.SetControllerServiceState(controllerService, "DISABLED")
}

//User Tennants
type UserComponent struct {
	Id            string    `json:"id,omitempty"`
	ParentGroupId string    `json:"parentGroupId,omitempty"`
	Identity      string    `json:"identity,omitempty"`
	Position      *Position `json:"position,omitempty"`
}
type User struct {
	Revision  Revision      `json:"revision"`
	Component UserComponent `json:"component"`
}

func UserStub() *User {
	return &User{
		Component: UserComponent{
			Position: &Position{},
		},
	}
}
func (c *Client) CreateUser(user *User) error {
	url := fmt.Sprintf("http://%s/%s/tenants/users",
		c.Config.Host, c.Config.ApiPath)
	_, err := c.JsonCall("POST", url, user, user)
	return err
}
func (c *Client) GetUser(userId string) (*User, error) {
	url := fmt.Sprintf("http://%s/%s/tenants/users/%s",
		c.Config.Host, c.Config.ApiPath, userId)
	user := UserStub()
	code, err := c.JsonCall("GET", url, nil, &user)
	if 404 == code {
		return nil, fmt.Errorf("not_found")
	}
	if nil != err {
		return nil, err
	}
	return user, nil
}

func (c *Client) DeleteUser(user *User) error {
	url := fmt.Sprintf("http://%s/%s/tenants/users/%s?version=%d",
		c.Config.Host, c.Config.ApiPath, user.Component.Id, user.Revision.Version)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

//Group Tennants
type GroupComponent struct {
	Id            string    `json:"id,omitempty"`
	ParentGroupId string    `json:"parentGroupId,omitempty"`
	Identity      string    `json:"identity,omitempty"`
	Position      *Position `json:"position,omitempty"`
	Users         []*User   `json:"users"`
}
type Group struct {
	Revision  Revision       `json:"revision"`
	Component GroupComponent `json:"component"`
}

func GroupStub() *Group {
	return &Group{
		Component: GroupComponent{
			Position: &Position{},
			Users:    []*User{},
		},
	}
}
func (c *Client) CreateGroup(group *Group) error {
	url := fmt.Sprintf("http://%s/%s/tenants/user-groups",
		c.Config.Host, c.Config.ApiPath)
	_, err := c.JsonCall("POST", url, group, group)
	return err
}
func (c *Client) GetGroup(groupId string) (*Group, error) {
	url := fmt.Sprintf("http://%s/%s/tenants/user-groups/%s",
		c.Config.Host, c.Config.ApiPath, groupId)
	group := GroupStub()
	code, err := c.JsonCall("GET", url, nil, &group)
	if 404 == code {
		return nil, fmt.Errorf("not_found")
	}
	if nil != err {
		return nil, err
	}
	return group, nil
}
func (c *Client) UpdateGroup(group *Group) error {
	url := fmt.Sprintf("http://%s/%s/tenants/user-groups/%s",
		c.Config.Host, c.Config.ApiPath, group.Component.Id)
	_, err := c.JsonCall("PUT", url, group, group)
	if nil != err {
		return err
	}
	return nil
}
func (c *Client) DeleteGroup(group *Group) error {
	url := fmt.Sprintf("http://%s/%s/tenants/user-groups/%s?version=%d",
		c.Config.Host, c.Config.ApiPath, group.Component.Id, group.Revision.Version)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}
