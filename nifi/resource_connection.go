package nifi

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func ResourceConnection() *schema.Resource {
	return &schema.Resource{
		Create: ResourceConnectionCreate,
		Read:   ResourceConnectionRead,
		Update: ResourceConnectionUpdate,
		Delete: ResourceConnectionDelete,
		Exists: ResourceConnectionExists,

		Schema: map[string]*schema.Schema{
			"parent_group_id": SchemaParentGroupId(),
			"revision":        SchemaRevision(),
			"component": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parent_group_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"back_pressure_data_size_threshold": {
							Type:     schema.TypeString,
							Optional: true,
							Default: "1 GB",
						},
						"back_pressure_object_threshold": {
							Type:     schema.TypeInt,
							Optional: true,
							Default: 10000,
						},
						"source": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"group_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"destination": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"group_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"selected_relationships": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"bends": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"x": {
										Type:     schema.TypeFloat,
										Required: true,
									},
									"y": {
										Type:     schema.TypeFloat,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func ResourceConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	connection := Connection{}
	connection.Revision.Version = 0

	err := ConnectionFromSchema(d, &connection)
	if err != nil {
		return fmt.Errorf("Failed to parse Connection schema")
	}
	parentGroupId := connection.Component.ParentGroupId

	// Create connection
	client := meta.(*Client)
	err = client.CreateConnection(&connection)
	if err != nil {
		return fmt.Errorf("Failed to create Connection")
	}

	// Start related processors
	ConnectionStartProcessor(client, connection.Component.Source.Id)
	ConnectionStartProcessor(client, connection.Component.Destination.Id)

	// Indicate successful creation
	d.SetId(connection.Component.Id)
	d.Set("parent_group_id", parentGroupId)

	return ResourceConnectionRead(d, meta)
}

func ResourceConnectionRead(d *schema.ResourceData, meta interface{}) error {
	connectionId := d.Id()

	client := meta.(*Client)
	connection, err := client.GetConnection(connectionId)
	if err != nil {
		return fmt.Errorf("Error retrieving Connection: %s", connectionId)
	}

	err = ConnectionToSchema(d, connection)
	if err != nil {
		return fmt.Errorf("Failed to serialize Connection: %s", connectionId)
	}

	return nil
}

func ResourceConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	client.Lock.Lock()
	log.Printf("[INFO] Updating Connection: %s...", d.Id())
	err := ResourceConnectionUpdateInternal(d, meta)
	log.Printf("[INFO] Connection updated: %s", d.Id())
	defer client.Lock.Unlock()
	return err
}

func ResourceConnectionUpdateInternal(d *schema.ResourceData, meta interface{}) error {
	connectionId := d.Id()

	// Refresh connection details
	client := meta.(*Client)
	connection, err := client.GetConnection(connectionId)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Connection: %s", connectionId)
		}
	}

	// Stop related processors
	err = ConnectionStopProcessor(client, connection.Component.Source.Id)
	if err != nil {
		return fmt.Errorf("Failed to stop source Processor: %s", connection.Component.Source.Id)
	}
	err = ConnectionStopProcessor(client, connection.Component.Destination.Id)
	if err != nil {
		return fmt.Errorf("Failed to stop destination Processor: %s", connection.Component.Destination.Id)
	}

	// Update connection
	err = ConnectionFromSchema(d, connection)
	if err != nil {
		return fmt.Errorf("Failed to parse Connection schema: %s", connectionId)
	}
	err = client.UpdateConnection(connection)
	if err != nil {
		return fmt.Errorf("Failed to update Connection: %s", connectionId)
	}

	// Start related processors
	ConnectionStartProcessor(client, connection.Component.Source.Id)
	ConnectionStartProcessor(client, connection.Component.Destination.Id)

	return ResourceConnectionRead(d, meta)
}

func ResourceConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	client.Lock.Lock()
	log.Printf("[INFO] Deleting Connection: %s...", d.Id())
	err := ResourceConnectionDeleteInternal(d, meta)
	log.Printf("[INFO] Connection deleted: %s", d.Id())
	defer client.Lock.Unlock()
	return err
}

func ResourceConnectionDeleteInternal(d *schema.ResourceData, meta interface{}) error {
	connectionId := d.Id()

	// Refresh connection details
	client := meta.(*Client)
	connection, err := client.GetConnection(connectionId)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Connection: %s", connectionId)
		}
	}

	// Stop related processors if it is started
	err = ConnectionStopProcessor(client, connection.Component.Source.Id)
	if err != nil {
		return fmt.Errorf("Failed to stop source Processor: %s", connection.Component.Source.Id)
	}
	err = ConnectionStopProcessor(client, connection.Component.Destination.Id)
	if err != nil {
		return fmt.Errorf("Failed to stop destination Processor: %s", connection.Component.Destination.Id)
	}

	// Purge connection data
	err = client.DropConnectionData(connection)
	if nil != err {
		return fmt.Errorf("Error purging Connection: %s", connectionId)
	}

	// Delete connection
	err = client.DeleteConnection(connection)
	if err != nil {
		return fmt.Errorf("Error deleting Connection: %s", connectionId)
	}

	// Start related processors
	ConnectionStartProcessor(client, connection.Component.Source.Id)
	ConnectionStartProcessor(client, connection.Component.Destination.Id)

	d.SetId("")
	return nil
}

func ResourceConnectionExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	connectionId := d.Id()

	client := meta.(*Client)
	_, err := client.GetConnection(connectionId)
	if nil != err {
		if "not_found" == err.Error() {
			log.Printf("[INFO] Connection %s no longer exists, removing from state...", connectionId)
			d.SetId("")
			return false, nil
		} else {
			return false, fmt.Errorf("Error testing existence of Connection: %s", connectionId)
		}
	}

	return true, nil
}

// Processor Helpers

func ConnectionStartProcessor(client *Client, processorId string) error {
	processor, err := client.GetProcessor(processorId)
	if err != nil {
		return fmt.Errorf("Error retrieving Processor: %s", processorId)
	}
	if "RUNNING" != processor.Component.State {
		err = client.StartProcessor(processor)
		if err != nil {
			return fmt.Errorf("Failed to start Processor: %s", processorId)
		}
	}
	return nil
}

func ConnectionStopProcessor(client *Client, processorId string) error {
	processor, err := client.GetProcessor(processorId)
	if err != nil {
		return fmt.Errorf("Error retrieving Processor: %s", processorId)
	}
	if "RUNNING" == processor.Component.State {
		err = client.StopProcessor(processor)
		if err != nil {
			return fmt.Errorf("Failed to stop Processor: %s", processorId)
		}
	}
	return nil
}

// Schema Helpers

func ConnectionFromSchema(d *schema.ResourceData, connection *Connection) error {
	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})
	connection.Component.ParentGroupId = component["parent_group_id"].(string)

	connection.Component.BackPressureDataSizeThreshold = component["back_pressure_data_size_threshold"].(string)
	connection.Component.BackPressureObjectThreshold = component["back_pressure_object_threshold"].(int)

	v = component["source"].([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component.source is required")
	}
	source := v[0].(map[string]interface{})
	connection.Component.Source.Type = source["type"].(string)
	connection.Component.Source.Id = source["id"].(string)
	connection.Component.Source.GroupId = source["group_id"].(string)

	v = component["destination"].([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component.destination is required")
	}
	destination := v[0].(map[string]interface{})
	connection.Component.Destination.Type = destination["type"].(string)
	connection.Component.Destination.Id = destination["id"].(string)
	connection.Component.Destination.GroupId = destination["group_id"].(string)

	selectedRelationships := []string{}
	relationships := component["selected_relationships"].([]interface{})
	for _, v := range relationships {
		selectedRelationships = append(selectedRelationships, v.(string))
	}
	connection.Component.SelectedRelationships = selectedRelationships

	v = component["bends"].([]interface{})
	if len(v) > 0 {
		bends := []Position{}
		for _, vv := range v {
			bend := vv.(map[string]interface{})
			bends = append(bends, Position{
				X: bend["x"].(float64),
				Y: bend["y"].(float64),
			})
		}
		connection.Component.Bends = bends
	}

	return nil
}

func ConnectionToSchema(d *schema.ResourceData, connection *Connection) error {
	revision := []map[string]interface{}{{
		"version": connection.Revision.Version,
	}}
	d.Set("revision", revision)

	relationships := []interface{}{}
	for _, v := range connection.Component.SelectedRelationships {
		relationships = append(relationships, v)
	}

	bends := []interface{}{}
	for _, v := range connection.Component.Bends {
		bends = append(bends, map[string]interface{}{
			"x": v.X,
			"y": v.Y,
		})
	}

	component := []map[string]interface{}{{
		"parent_group_id": d.Get("parent_group_id").(string),
		"back_pressure_data_size_threshold": connection.Component.BackPressureDataSizeThreshold,
		"back_pressure_object_threshold": connection.Component.BackPressureObjectThreshold,
		"source": []map[string]interface{}{{
			"type": 	connection.Component.Source.Type,
			"id":   	connection.Component.Source.Id,
			"group_id":	connection.Component.Source.GroupId,
		}},
		"destination": []map[string]interface{}{{
			"type": 	connection.Component.Destination.Type,
			"id":   	connection.Component.Destination.Id,
			"group_id":	connection.Component.Destination.GroupId,
		}},
		"selected_relationships": relationships,
		"bends":                  bends,
	}}
	d.Set("component", component)

	return nil
}
