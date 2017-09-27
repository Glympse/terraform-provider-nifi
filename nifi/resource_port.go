package nifi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func ResourcePort() *schema.Resource {
	return &schema.Resource{
		Create: ResourcePortCreate,
		Read:   ResourcePortRead,
		Update: ResourcePortUpdate,
		Delete: ResourcePortDelete,
		Exists: ResourcePortExists,

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
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"position": SchemaPosition(),
						"comments": {
							Type:     schema.TypeString,
							Required: false,
						},
					},
				},
			},
		},
	}
}

func ResourcePortCreate(d *schema.ResourceData, meta interface{}) error {
	port := &Port{}
	port.Revision.Version = 0

	err := PortFromSchema(d, port)
	if err != nil {
		return fmt.Errorf("Failed to parse Processor schema")
	}
	parentGroupId := port.Component.ParentGroupId

	// Create processor
	client := meta.(*Client)
	err = client.CreatePort(port)
	if err != nil {
		return fmt.Errorf("Failed to create Port")
	}

	// Start processor upon creation
	err = client.StartPort(port)
	if nil != err {
		log.Printf("[INFO] Failed to start Port: %s ", port.Component.Id)
	}
	// Indicate successful creation
	d.SetId(port.Component.Id)
	d.Set("parent_group_id", parentGroupId)

	return ResourcePortRead(d, meta)
}

func ResourcePortRead(d *schema.ResourceData, meta interface{}) error {
	portId := d.Id()
	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})
	port_type := component["type"].(string)

	client := meta.(*Client)
	port, err := client.GetPort(portId, port_type)
	if err != nil {
		return fmt.Errorf("Error retrieving Port: %s", portId)
	}

	err = PortToSchema(d, port)
	if err != nil {
		return fmt.Errorf("Failed to serialize Port: %s", portId)
	}

	return nil
}

func ResourcePortUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	client.Lock.Lock()
	log.Printf("[INFO] Updating Port: %s...", d.Id())
	err := ResourcePortUpdateInternal(d, meta)
	log.Printf("[INFO] Port updated: %s", d.Id())
	defer client.Lock.Unlock()
	return err
}

func ResourcePortUpdateInternal(d *schema.ResourceData, meta interface{}) error {
	portId := d.Id()
	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})
	port_type := component["type"].(string)
	// Refresh processor details
	client := meta.(*Client)
	port, err := client.GetPort(portId, port_type)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Processor: %s", portId)
		}
	}

	// Stop processor if it is currently running
	if "RUNNING" == port.Component.State {
		err = client.StopPort(port)
		if err != nil {
			return fmt.Errorf("Failed to stop Port: %s", portId)
		}
	}

	// Load processor's desired state
	err = PortFromSchema(d, port)
	if err != nil {
		return fmt.Errorf("Failed to parse Port schema: %s", portId)
	}

	// Update processor
	// err = client.UpdateProcessor(processor)
	// if err != nil {
	// 	return fmt.Errorf("Failed to update Processor: %s", processorId)
	// }

	// Start processor again
	err = client.StartPort(port)
	if err != nil {
		log.Printf("[INFO] Failed to start Port: %s", portId)
	}

	return ResourcePortRead(d, meta)
}

func ResourcePortDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	client.Lock.Lock()
	log.Printf("[INFO] Deleting Port: %s...", d.Id())
	err := ResourceProcessorDeleteInternal(d, meta)
	log.Printf("[INFO] Port deleted: %s", d.Id())
	defer client.Lock.Unlock()
	return err
}

func ResourcePortDeleteInternal(d *schema.ResourceData, meta interface{}) error {
	portId := d.Id()
	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})
	port_type := component["type"].(string)

	// Refresh processor details
	client := meta.(*Client)
	port, err := client.GetPort(portId, port_type)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Port: %s", portId)
		}
	}

	// Stop processor if it is currently running
	if "RUNNING" == port.Component.State {
		err = client.StopPort(port)
		if err != nil {
			return fmt.Errorf("Failed to stop Port: %s", portId)
		}
	}

	// Delete processor
	err = client.DeletePort(port)
	if err != nil {
		return fmt.Errorf("Error deleting Port: %s", portId)
	}

	d.SetId("")
	return nil
}

func ResourcePortExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	portId := d.Id()
	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return false, fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})
	port_type := component["type"].(string)

	client := meta.(*Client)
	_, err := client.GetPort(portId, port_type)
	if nil != err {
		if "not_found" == err.Error() {
			log.Printf("[INFO] Port %s no longer exists, removing from state...", portId)
			d.SetId("")
			return false, nil
		} else {
			return false, fmt.Errorf("Error testing existence of Port: %s", portId)
		}
	}

	return true, nil
}

// Connection Helpers

// Schema Helpers

func PortFromSchema(d *schema.ResourceData, port *Port) error {
	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})
	port.Component.ParentGroupId = component["parent_group_id"].(string)
	port.Component.Name = component["name"].(string)
	port.Component.PortType = component["type"].(string)

	v = component["position"].([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component.position is required")
	}
	position := v[0].(map[string]interface{})
	port.Component.Position.X = position["x"].(float64)
	port.Component.Position.Y = position["y"].(float64)
	return nil
}

func PortToSchema(d *schema.ResourceData, port *Port) error {
	revision := []map[string]interface{}{{
		"version": port.Revision.Version,
	}}
	d.Set("revision", revision)

	component := []map[string]interface{}{{
		"parent_group_id": d.Get("parent_group_id").(string),
		"name":            port.Component.Name,
		"type":            port.Component.PortType,
		"position": []map[string]interface{}{{
			"x": port.Component.Position.X,
			"y": port.Component.Position.Y,
		}},
	}}
	d.Set("component", component)
	return nil
}
