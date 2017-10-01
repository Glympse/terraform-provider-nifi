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
	HttpScheme string
	Lock       sync.Mutex
}

func NewClient(config Config) *Client {
	http_client := &http.Client{}
	scheme := "http"
	if config.AdminCertPath != "" && config.AdminKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(config.AdminCertPath, config.AdminKeyPath)
		if err != nil {
			log.Fatal(err)
		} else {
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			tlsConfig.BuildNameToCertificate()
			tlsConfig.InsecureSkipVerify = true
			transport := &http.Transport{TLSClientConfig: tlsConfig}
			http_client = &http.Client{Transport: transport}
			scheme = "https"
		}
	}
	client := &Client{
		Config:     config,
		Client:     http_client,
		HttpScheme: scheme,
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
		request.Header.Add("Accept", "application/json")
	}

	response, err := c.Client.Do(request)
	log.Printf("[DEBUG]: http call error code: %d", response.StatusCode)
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
	url := fmt.Sprintf("%s://%s/%s/process-groups/%s/process-groups",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processGroup.Component.ParentGroupId)
	_, err := c.JsonCall("POST", url, processGroup, processGroup)
	return err
}

func (c *Client) GetProcessGroup(processGroupId string) (*ProcessGroup, error) {
	url := fmt.Sprintf("%s://%s/%s/process-groups/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processGroupId)
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
	url := fmt.Sprintf("%s://%s/%s/process-groups/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processGroup.Component.Id)
	_, err := c.JsonCall("PUT", url, processGroup, processGroup)
	return err
}

func (c *Client) DeleteProcessGroup(processGroup *ProcessGroup) error {
	url := fmt.Sprintf("%s://%s/%s/process-groups/%s?version=%d",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processGroup.Component.Id, processGroup.Revision.Version)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

func (c *Client) GetProcessGroupConnections(processGroupId string) (*Connections, error) {
	url := fmt.Sprintf("%s://%s/%s/process-groups/%s/connections",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processGroupId)
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
	url := fmt.Sprintf("%s://%s/%s/process-groups/%s/processors",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processor.Component.ParentGroupId)
	_, err := c.JsonCall("POST", url, processor, processor)
	if nil != err {
		return err
	}
	c.CleanupNilProperties(processor.Component.Config.Properties)
	return nil
}

func (c *Client) GetProcessor(processorId string) (*Processor, error) {
	url := fmt.Sprintf("%s://%s/%s/processors/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processorId)
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
	url := fmt.Sprintf("%s://%s/%s/processors/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processor.Component.Id)
	_, err := c.JsonCall("PUT", url, processor, processor)
	if nil != err {
		return err
	}
	c.CleanupNilProperties(processor.Component.Config.Properties)
	return nil
}

func (c *Client) DeleteProcessor(processor *Processor) error {
	url := fmt.Sprintf("%s://%s/%s/processors/%s?version=%d",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processor.Component.Id, processor.Revision.Version)
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
	url := fmt.Sprintf("%s://%s/%s/processors/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processor.Component.Id)
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
	url := fmt.Sprintf("%s://%s/%s/process-groups/%s/connections",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, connection.Component.ParentGroupId)
	_, err := c.JsonCall("POST", url, connection, connection)
	return err
}

func (c *Client) GetConnection(connectionId string) (*Connection, error) {
	url := fmt.Sprintf("%s://%s/%s/connections/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, connectionId)
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
	url := fmt.Sprintf("%s://%s/%s/connections/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, connection.Component.Id)
	_, err := c.JsonCall("PUT", url, connection, connection)
	return err
}

func (c *Client) DeleteConnection(connection *Connection) error {
	url := fmt.Sprintf("%s://%s/%s/connections/%s?version=%d",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, connection.Component.Id, connection.Revision.Version)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

func (c *Client) DropConnectionData(connection *Connection) error {
	// Create a request to drop the contents of the queue in this connection
	url := fmt.Sprintf("%s://%s/%s/flowfile-queues/%s/drop-requests",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, connection.Component.Id)
	dropRequest := ConnectionDropRequest{}
	_, err := c.JsonCall("POST", url, nil, &dropRequest)
	if nil != err {
		return err
	}

	// Give it some time to complete
	maxAttempts := 10
	for iteration := 0; iteration < maxAttempts; iteration++ {
		// Check status of the request
		url = fmt.Sprintf("%s://%s/%s/flowfile-queues/%s/drop-requests/%s",
			c.HttpScheme, c.Config.Host, c.Config.ApiPath, connection.Component.Id, dropRequest.DropRequest.Id)
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
	url = fmt.Sprintf("%s://%s/%s/flowfile-queues/%s/drop-requests/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, connection.Component.Id, dropRequest.DropRequest.Id)
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
	url := fmt.Sprintf("%s://%s/%s/process-groups/%s/controller-services",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, controllerService.Component.ParentGroupId)
	_, err := c.JsonCall("POST", url, controllerService, controllerService)
	if nil != err {
		return err
	}
	c.CleanupNilProperties(controllerService.Component.Properties)
	return nil
}

func (c *Client) GetControllerService(controllerServiceId string) (*ControllerService, error) {
	url := fmt.Sprintf("%s://%s/%s/controller-services/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, controllerServiceId)
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
	url := fmt.Sprintf("%s://%s/%s/controller-services/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, controllerService.Component.Id)
	_, err := c.JsonCall("PUT", url, controllerService, controllerService)
	if nil != err {
		return err
	}
	c.CleanupNilProperties(controllerService.Component.Properties)
	return nil
}

func (c *Client) DeleteControllerService(controllerService *ControllerService) error {
	url := fmt.Sprintf("%s://%s/%s/controller-services/%s?version=%d",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, controllerService.Component.Id, controllerService.Revision.Version)
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
	url := fmt.Sprintf("%s://%s/%s/controller-services/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, controllerService.Component.Id)
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
type Tenant struct {
	Id string `json:"id"`
}

type TenantSearchResult struct {
	Users      []Tenant `json:"users"`
	UserGroups []Tenant `json:"userGroups"`
}

type UserComponent struct {
	Id            string    `json:"id,omitempty"`
	ParentGroupId string    `json:"parentGroupId,omitempty"`
	Identity      string    `json:"identity,omitempty"`
	Position      *Position `json:"position,omitempty"`
}

func (uc UserComponent) String() string {
	return fmt.Sprintf("Id:%v ParentGroupID:%v, Identity:%v", uc.Id, uc.ParentGroupId, uc.Identity)
}

func (u User) ToTenant() *Tenant {
	return &Tenant{
		Id: u.Component.Id,
	}
}

type User struct {
	Revision  Revision      `json:"revision"`
	Component UserComponent `json:"component"`
}

func (u User) String() string {
	return fmt.Sprintf("User: {Component :{%v}}", u.Component)
}
func UserStub() *User {
	return &User{
		Component: UserComponent{
			Position: &Position{},
		},
	}
}
func (c *Client) CreateUser(user *User) error {
	url := fmt.Sprintf("%s://%s/%s/tenants/users",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath)
	_, err := c.JsonCall("POST", url, user, user)
	return err
}
func (c *Client) GetUser(userId string) (*User, error) {
	url := fmt.Sprintf("%s://%s/%s/tenants/users/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, userId)
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
func (c *Client) GetUserIdsWithIdentity(userIden string) ([]string, error) {
	//https://localhost:9443/nifi-api/tenants/search-results?q=test_user

	searchResult := TenantSearchResult{}

	url := fmt.Sprintf("%s://%s/%s/tenants/search-results?q=%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, userIden)

	code, err := c.JsonCall("GET", url, nil, &searchResult)

	userIds := []string{}
	if 404 == code {
		return userIds, fmt.Errorf("not_found")
	}
	if nil != err {
		return userIds, err
	}
	for i := 0; i < len(searchResult.Users); i++ {
		foundId := searchResult.Users[i].Id
		userIds = append(userIds, foundId)
	}
	return userIds, nil
}

func (c *Client) DeleteUser(user *User) error {
	url := fmt.Sprintf("%s://%s/%s/tenants/users/%s?version=%d",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, user.Component.Id, user.Revision.Version)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

//Group Tennants
type GroupComponent struct {
	Id            string    `json:"id,omitempty"`
	ParentGroupId string    `json:"parentGroupId,omitempty"`
	Identity      string    `json:"identity,omitempty"`
	Position      *Position `json:"position,omitempty"`
	Users         []Tenant  `json:"users,omitempty"`
}

func (c GroupComponent) String() string {
	return fmt.Sprintf("Id: %v ParentGroupID: %v, Identity: %v", c.Id, c.ParentGroupId, c.Identity)
}

type Group struct {
	Revision  Revision       `json:"revision"`
	Component GroupComponent `json:"component"`
}

func (c Group) String() string {
	return fmt.Sprintf("Group: { Component:{ %v } }", c.Component)
}

func GroupStub() *Group {
	return &Group{
		Component: GroupComponent{
			Position: &Position{},
		},
	}
}
func (c *Client) CreateGroup(group *Group) error {
	url := fmt.Sprintf("%s://%s/%s/tenants/user-groups",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath)
	_, err := c.JsonCall("POST", url, group, group)
	return err
}
func (c *Client) GetGroup(groupId string) (*Group, error) {
	url := fmt.Sprintf("%s://%s/%s/tenants/user-groups/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, groupId)
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
func (c *Client) GetGroupIdsWithIdentity(groupIden string) ([]string, error) {
	//https://localhost:9443/nifi-api/tenants/search-results?q=test_user

	searchResult := TenantSearchResult{}

	url := fmt.Sprintf("%s://%s/%s/tenants/search-results?q=%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, groupIden)

	code, err := c.JsonCall("GET", url, nil, &searchResult)

	groupIds := []string{}
	if 404 == code {
		return groupIds, fmt.Errorf("not_found")
	}
	if nil != err {
		return groupIds, err
	}
	for i := 0; i < len(searchResult.UserGroups); i++ {
		foundId := searchResult.UserGroups[i].Id
		groupIds = append(groupIds, foundId)
	}
	return groupIds, nil
}
func (c *Client) UpdateGroup(group *Group) error {
	url := fmt.Sprintf("%s://%s/%s/tenants/user-groups/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, group.Component.Id)
	_, err := c.JsonCall("PUT", url, group, group)
	if nil != err {
		return err
	}
	return nil
}
func (c *Client) DeleteGroup(group *Group) error {
	url := fmt.Sprintf("%s://%s/%s/tenants/user-groups/%s?version=%d",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, group.Component.Id, group.Revision.Version)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

//remote process group
type RemoteProcessGroupComponent struct {
	Id                string   `json:"id,omitempty"`
	ParentGroupId     string   `json:"parentGroupId"`
	Name              string   `json:"name"`
	Position          Position `json:"position"`
	TargetUris        string   `json:"targetUris"`
	TransportProtocol string   `json:"transportProtocol"`
}

type RemoteProcessGroup struct {
	Revision  Revision                    `json:"revision"`
	Component RemoteProcessGroupComponent `json:"component"`
}

func (c *Client) CreateRemoteProcessGroup(processGroup *RemoteProcessGroup) error {
	url := fmt.Sprintf("%s://%s/%s/process-groups/%s/remote-process-groups",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processGroup.Component.ParentGroupId)
	_, err := c.JsonCall("POST", url, processGroup, processGroup)
	return err
}

func (c *Client) GetRemoteProcessGroup(processGroupId string) (*RemoteProcessGroup, error) {
	url := fmt.Sprintf("%s://%s/%s/remote-process-groups/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processGroupId)
	processGroup := RemoteProcessGroup{}
	code, err := c.JsonCall("GET", url, nil, &processGroup)
	if 404 == code {
		return nil, fmt.Errorf("not_found")
	}
	if nil != err {
		return nil, err
	}
	return &processGroup, nil
}

func (c *Client) UpdateRemoteProcessGroup(processGroup *RemoteProcessGroup) error {
	url := fmt.Sprintf("%s://%s/%s/remote-process-groups/%s",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processGroup.Component.Id)
	_, err := c.JsonCall("PUT", url, processGroup, processGroup)
	return err
}

func (c *Client) DeleteRemoteProcessGroup(processGroup *RemoteProcessGroup) error {
	url := fmt.Sprintf("%s://%s/%s/process-groups/%s?version=%d",
		c.HttpScheme, c.Config.Host, c.Config.ApiPath, processGroup.Component.Id, processGroup.Revision.Version)
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

//input port
type Port struct {
	Revision  Revision      `json:"revision"`
	Component PortComponent `json:"component"`
}
type PortComponent struct {
	Id            string   `json:"id,omitempty"`
	ParentGroupId string   `json:"parentGroupId"`
	Name          string   `json:"name"`
	PortType      string   `json:"type"`
	Comments      string   `json:"comments"`
	Position      Position `json:"position"`
	State         string   `json:"state,omitempty"`
}

type PortStateComponent struct {
	Id    string `json:"id,omitempty"`
	State string `json:"state,omitempty"`
}
type PortStateUpdate struct {
	Revision  Revision           `json:"revision"`
	Component PortStateComponent `json:"component"`
}

func (c *Client) CreatePort(port *Port) error {
	parent_group_id := port.Component.ParentGroupId
	port_type := port.Component.PortType
	url := ""
	switch port_type {
	case "INPUT_PORT":
		url = fmt.Sprintf("%s://%s/%s/process-groups/%s/input-ports",
			c.HttpScheme, c.Config.Host, c.Config.ApiPath, parent_group_id)
	case "OUTPUT_PORT":
		url = fmt.Sprintf("%s://%s/%s/process-groups/%s/output-ports",
			c.HttpScheme, c.Config.Host, c.Config.ApiPath, parent_group_id)
	default:
		log.Fatal(fmt.Printf("Invalid port type : %s.", port_type))
	}
	_, err := c.JsonCall("POST", url, port, port)
	return err
}
func (c *Client) UpdatePort(port *Port) error {
	port_type := port.Component.PortType
	portId := port.Component.Id
	url := ""
	switch port_type {
	case "INPUT_PORT":
		url = fmt.Sprintf("%s://%s/%s/input-ports/%s",
			c.HttpScheme, c.Config.Host, c.Config.ApiPath, portId)
	case "OUTPUT_PORT":
		url = fmt.Sprintf("%s://%s/%s/output-ports/%s",
			c.HttpScheme, c.Config.Host, c.Config.ApiPath, portId)
	default:
		log.Fatal(fmt.Printf("Invalid port type : %s.", port_type))
	}
	responseCode, err := c.JsonCall("PUT", url, port, port)
	if responseCode == 409 {
		log.Printf("[WARN]: port not updated, since it's not invalid state")
	}
	return err
}
func (c *Client) GetPort(portId string, port_type string) (*Port, error) {
	url := ""
	switch port_type {
	case "INPUT_PORT":
		url = fmt.Sprintf("%s://%s/%s/input-ports/%s",
			c.HttpScheme, c.Config.Host, c.Config.ApiPath, portId)
	case "OUTPUT_PORT":
		url = fmt.Sprintf("%s://%s/%s/output-ports/%s",
			c.HttpScheme, c.Config.Host, c.Config.ApiPath, portId)
	default:
		log.Fatal(fmt.Printf("Invalid port type : %s.", port_type))
	}
	port := Port{}
	code, err := c.JsonCall("GET", url, nil, &port)
	if 404 == code {
		return nil, fmt.Errorf("not_found")
	}
	if nil != err {
		return nil, err
	}
	return &port, nil
}

func (c *Client) DeletePort(port *Port) error {
	port_id := port.Component.Id
	port_type := port.Component.PortType
	url := ""
	switch port_type {
	case "INPUT_PORT":
		url = fmt.Sprintf("%s://%s/%s/input-ports/%s?version=%d",
			c.HttpScheme, c.Config.Host, c.Config.ApiPath, port_id, port.Revision.Version)
	case "OUTPUT_PORT":
		url = fmt.Sprintf("%s://%s/%s/output-ports/%s?version=%d",
			c.HttpScheme, c.Config.Host, c.Config.ApiPath, port_id, port.Revision.Version)
	default:
		log.Fatal(fmt.Printf("Invalid port type : %s.", port_type))
	}
	_, err := c.JsonCall("DELETE", url, nil, nil)
	return err
}

func (c *Client) SetPortState(port *Port, state string) error {
	log.Printf("[Info] Set port to state %s", state)
	//https://community.hortonworks.com/questions/67900/startstop-processor-via-nifi-api.html
	stateUpdate := PortStateUpdate{
		Revision: Revision{
			Version: port.Revision.Version,
		},
		Component: PortStateComponent{
			Id:    port.Component.Id,
			State: state,
		},
	}

	port_type := port.Component.PortType
	portId := port.Component.Id
	url := ""
	switch port_type {
	case "INPUT_PORT":
		url = fmt.Sprintf("%s://%s/%s/input-ports/%s",
			c.HttpScheme, c.Config.Host, c.Config.ApiPath, portId)
	case "OUTPUT_PORT":
		url = fmt.Sprintf("%s://%s/%s/output-ports/%s",
			c.HttpScheme, c.Config.Host, c.Config.ApiPath, portId)
	default:
		log.Fatal(fmt.Printf("Invalid port type : %s.", port_type))
	}
	// var buffer = new(bytes.Buffer)
	// json.NewEncoder(buffer).Encode(stateUpdate)
	// log.Printf(url)
	// log.Printf(buffer.String())
	// buffer = new(bytes.Buffer)
	// json.NewEncoder(buffer).Encode(port)
	// log.Printf(buffer.String())

	responseCode, err := c.JsonCall("PUT", url, stateUpdate, port)
	if err != nil {
		log.Printf("[Fatal]: Failed to set state of  Port, error code %s.", err)
		if responseCode == 409 {
			// if 409, same state
			log.Printf(fmt.Sprintf("[WARN]: 409 %s.", err))
			err = nil
		}
	}

	//verify port state
	maxAttempts := 5
	state_verified := false
	for iteration := 0; iteration < maxAttempts; iteration++ {
		// Check status of the request
		// Wait a bit
		time.Sleep(3 * time.Second)
		_, err = c.JsonCall("GET", url, nil, port)
		if nil != err {
			continue
		} else {
			if port.Component.State == state {
				log.Printf("[DEBUG] port status set")
				state_verified = true
				break
			}
		}
		// Log progress
		log.Printf("[DEBUG] Checking Port status %d %d...", portId, iteration+1)

		if maxAttempts-1 == iteration {
			log.Printf("[DEBUG] Failed to verify Port new status %s", state)
		}
	}
	if !state_verified {
		log.Printf("[DEBUG] Failed to verify Port new status %s", state)
	}
	return err
}

func (c *Client) StartPort(port *Port) error {
	return c.SetPortState(port, "RUNNING")
}

func (c *Client) StopPort(port *Port) error {
	return c.SetPortState(port, "STOPPED")
}

func (c *Client) DisablePort(port *Port) error {
	return c.SetPortState(port, "DISABLED")
}

func (c *Client) StopConnectionHand(connectionHand *ConnectionHand) error {
	handType := connectionHand.Type
	handId := connectionHand.Id
	log.Printf("[DEBUG] Stop connection hand %s , %d", handType, handId)
	switch handType {
	case "PROCESSOR":
		processor, err := c.GetProcessor(handId)
		if err != nil {
			return c.StopProcessor(processor)
		} else {
			return err
		}
	case "INPUT_PORT":
		port, err := c.GetPort(handId, "INPUT_PORT")
		if err == nil {
			return c.StopPort(port)
		} else {
			log.Printf("Fail to get Port %s", handId)
			return err
		}
	case "OUTPUT_PORT":
		port, err := c.GetPort(handId, "OUTPUT_PORT")
		if err == nil {
			return c.StopPort(port)
		} else {
			log.Printf("Fail to get Port %s", handId)
			return err
		}
	case "FUNNEL":
		log.Printf("No need to stop Funnel")
		return nil
	default:
		log.Fatal(fmt.Sprintf("[WARN]: not supported connection source/target type : %s", handType))
	}
	return nil
}

func (c *Client) StartConnectionHand(connectionHand *ConnectionHand) error {
	handType := connectionHand.Type
	handId := connectionHand.Id
	log.Printf("[DEBUG] Start connection hand %s , %d", handType, handId)
	switch handType {
	case "PROCESSOR":
		processor, err := c.GetProcessor(handId)
		if err != nil {
			return c.StartProcessor(processor)
		} else {
			return err
		}
	case "INPUT_PORT":
		port, err := c.GetPort(handId, "INPUT_PORT")
		if err == nil {
			return c.StartPort(port)
		} else {
			return err
		}
	case "OUTPUT_PORT":
		port, err := c.GetPort(handId, "OUTPUT_PORT")
		if err == nil {
			return c.StartPort(port)
		} else {
			return err
		}
	case "FUNNEL":
		log.Printf("No need to start Funnel")
		return nil
	default:
		log.Printf("[WARN]: not supported connection source/target type : %s", handType)
	}
	return nil
}
