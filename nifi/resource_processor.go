package nifi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceProcessor() *schema.Resource {
	return &schema.Resource{
		Create: ResourceProcessorCreate,
		Read:   ResourceProcessorRead,
		Update: ResourceProcessorUpdate,
		Delete: ResourceProcessorDelete,
		Exists: ResourceProcessorExists,

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
						"config": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"concurrently_schedulable_task_count": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  1,
									},
									"scheduling_strategy": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "TIMER_DRIVEN",
									},
									"scheduling_period": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "0 sec",
									},
									"properties": {
										Type:     schema.TypeMap,
										Required: true,
									},
									"auto_terminated_relationships": {
										Type:     schema.TypeList,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
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

func ResourceProcessorCreate(d *schema.ResourceData, meta interface{}) error {
	processor := ProcessorStub()
	processor.Revision.Version = 0

	err := ProcessorFromSchema(d, processor)
	if err != nil {
		return fmt.Errorf("Failed to parse Processor schema")
	}
	parentGroupId := processor.Component.ParentGroupId

	// Create processor
	client := meta.(*Client)
	err = client.CreateProcessor(processor)
	if err != nil {
		return fmt.Errorf("Failed to create Processor")
	}

	// Start processor upon creation
	err = client.StartProcessor(processor)
	if nil != err {
		log.Printf("[INFO] Failed to start Processor: %s ", processor.Component.Id)
	}

	// Indicate successful creation
	d.SetId(processor.Component.Id)
	d.Set("parent_group_id", parentGroupId)

	return ResourceProcessorRead(d, meta)
}

func ResourceProcessorRead(d *schema.ResourceData, meta interface{}) error {
	processorId := d.Id()

	client := meta.(*Client)
	processor, err := client.GetProcessor(processorId)
	if err != nil {
		return fmt.Errorf("Error retrieving Processor: %s", processorId)
	}

	err = ProcessorToSchema(d, processor)
	if err != nil {
		return fmt.Errorf("Failed to serialize Processor: %s", processorId)
	}

	return nil
}

func ResourceProcessorUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	client.Lock.Lock()
	log.Printf("[INFO] Updating Processor: %s...", d.Id())
	err := ResourceProcessorUpdateInternal(d, meta)
	log.Printf("[INFO] Processor updated: %s", d.Id())
	defer client.Lock.Unlock()
	return err
}

func ResourceProcessorUpdateInternal(d *schema.ResourceData, meta interface{}) error {
	processorId := d.Id()

	// Refresh processor details
	client := meta.(*Client)
	processor, err := client.GetProcessor(processorId)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Processor: %s", processorId)
		}
	}

	// Stop processor if it is currently running
	if "RUNNING" == processor.Component.State {
		err = client.StopProcessor(processor)
		if err != nil {
			return fmt.Errorf("Failed to stop Processor: %s", processorId)
		}
	}

	// Load processor's desired state
	err = ProcessorFromSchema(d, processor)
	if err != nil {
		return fmt.Errorf("Failed to parse Processor schema: %s", processorId)
	}

	// Compare new list of auto-terminated connections against the list of processor's existing connections.
	// It is not possible to auto-terminate a relationship if an existing connection declares this relationship type.
	err = ProcessorRemoveOverlappingConnections(client, processor)
	if nil != err {
		return fmt.Errorf("Failed to cleanup connections for Processor: %s", processorId)
	}

	// Update processor
	err = client.UpdateProcessor(processor)
	if err != nil {
		return fmt.Errorf("Failed to update Processor: %s", processorId)
	}

	// Start processor again
	err = client.StartProcessor(processor)
	if err != nil {
		log.Printf("[INFO] Failed to start Processor: %s", processorId)
	}

	return ResourceProcessorRead(d, meta)
}

func ResourceProcessorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	client.Lock.Lock()
	log.Printf("[INFO] Deleting Processor: %s...", d.Id())
	err := ResourceProcessorDeleteInternal(d, meta)
	log.Printf("[INFO] Processor deleted: %s", d.Id())
	defer client.Lock.Unlock()
	return err
}

func ResourceProcessorDeleteInternal(d *schema.ResourceData, meta interface{}) error {
	processorId := d.Id()

	// Refresh processor details
	client := meta.(*Client)
	processor, err := client.GetProcessor(processorId)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Processor: %s", processorId)
		}
	}

	// Stop processor if it is currently running
	if "RUNNING" == processor.Component.State {
		err = client.StopProcessor(processor)
		if err != nil {
			return fmt.Errorf("Failed to stop Processor: %s", processorId)
		}
	}

	// Delete processor
	err = client.DeleteProcessor(processor)
	if err != nil {
		return fmt.Errorf("Error deleting Processor: %s", processorId)
	}

	d.SetId("")
	return nil
}

func ResourceProcessorExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	processorId := d.Id()

	client := meta.(*Client)
	_, err := client.GetProcessor(processorId)
	if nil != err {
		if "not_found" == err.Error() {
			log.Printf("[INFO] Processor %s no longer exists, removing from state...", processorId)
			d.SetId("")
			return false, nil
		} else {
			return false, fmt.Errorf("Error testing existence of Processor: %s", processorId)
		}
	}

	return true, nil
}

// Connection Helpers

func ProcessorRemoveOverlappingConnections(client *Client, processor *Processor) error {
	// Build a set of processor's auto-terminated relationships
	terminatedRelationships := map[string]bool{}
	for _, v := range processor.Component.Config.AutoTerminatedRelationships {
		terminatedRelationships[v] = true
	}
	if 0 == len(terminatedRelationships) {
		// Nothing to do if processor does not define any auto-terminated relationships
		return nil
	}

	// Fetch the list of process group connections.
	groupConnections, err := client.GetProcessGroupConnections(processor.Component.ParentGroupId)
	if nil != err {
		return fmt.Errorf("Error retrieving Process Group connections: %s", processor.Component.ParentGroupId)
	}

	// Find a subset of these connections that overlap with the processor's auto-terminated relationships.
	overlappingConnections := []Connection{}
	for _, connection := range groupConnections.Connections {
		if connection.Component.Source.Id != processor.Component.Id {
			continue
		}
		found := false
		for _, relationship := range connection.Component.SelectedRelationships {
			if _, contains := terminatedRelationships[relationship]; contains {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		overlappingConnections = append(overlappingConnections, connection)
	}

	// Remove overlaps
	for _, connection := range overlappingConnections {
		// Stop destination processor
		//err = ConnectionStopProcessor(client, connection.Component.Destination.Id)
		err = client.StopConnectionHand(&connection.Component.Destination)
		if nil != err {
			log.Printf("[INFO] Failed to stop Processor: %s", connection.Component.Destination.Id)
			continue
		}

		// Prepare the list of relationships connection is allowed to keep
		filteredRelationships := connection.Component.SelectedRelationships[:0]
		for _, relationship := range connection.Component.SelectedRelationships {
			if _, contains := terminatedRelationships[relationship]; !contains {
				filteredRelationships = append(filteredRelationships, relationship)
			}
		}

		// Update/remove connection
		if len(filteredRelationships) > 0 {
			err = client.UpdateConnection(&connection)
			if nil != err {
				log.Printf("[INFO] Failed to update Connection: %s", connection.Component.Id)
			}
		} else {
			// Purge connection data
			err = client.DropConnectionData(&connection)
			if nil != err {
				log.Printf("[INFO] Error purging Connection: %s", connection.Component.Id)
			}

			// Remove the connection
			err = client.DeleteConnection(&connection)
			if nil != err {
				log.Printf("[INFO] Failed to delete Connection: %s", connection.Component.Id)
			}
		}

		// Start destination processor
		//err = ConnectionStartProcessor(client, connection.Component.Destination.Id)
		err = client.StartConnectionHand(&connection.Component.Destination)
		if nil != err {
			log.Printf("[INFO] Failed to start Processor: %s", connection.Component.Destination.Id)
		}
	}

	return nil
}

// Schema Helpers

func ProcessorFromSchema(d *schema.ResourceData, processor *Processor) error {
	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})
	processor.Component.ParentGroupId = component["parent_group_id"].(string)
	processor.Component.Name = component["name"].(string)
	processor.Component.Type = component["type"].(string)

	v = component["position"].([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component.position is required")
	}
	position := v[0].(map[string]interface{})
	processor.Component.Position.X = position["x"].(float64)
	processor.Component.Position.Y = position["y"].(float64)

	v = component["config"].([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component.config is required")
	}
	config := v[0].(map[string]interface{})

	processor.Component.Config.SchedulingStrategy = config["scheduling_strategy"].(string)
	processor.Component.Config.SchedulingPeriod = config["scheduling_period"].(string)
	processor.Component.Config.ConcurrentlySchedulableTaskCount = config["concurrently_schedulable_task_count"].(int)

	processor.Component.Config.Properties = map[string]interface{}{}
	properties := config["properties"].(map[string]interface{})
	for k, v := range properties {
		processor.Component.Config.Properties[k] = v.(string)
	}

	autoTerminatedRelationships := []string{}
	relationships := config["auto_terminated_relationships"].([]interface{})
	for _, v := range relationships {
		autoTerminatedRelationships = append(autoTerminatedRelationships, v.(string))
	}
	processor.Component.Config.AutoTerminatedRelationships = autoTerminatedRelationships

	return nil
}

func ProcessorToSchema(d *schema.ResourceData, processor *Processor) error {
	revision := []map[string]interface{}{{
		"version": processor.Revision.Version,
	}}
	d.Set("revision", revision)

	relationships := []interface{}{}
	for _, v := range processor.Component.Config.AutoTerminatedRelationships {
		relationships = append(relationships, v)
	}

	component := []map[string]interface{}{{
		"parent_group_id": d.Get("parent_group_id").(string),
		"name":            processor.Component.Name,
		"type":            processor.Component.Type,
		"position": []map[string]interface{}{{
			"x": processor.Component.Position.X,
			"y": processor.Component.Position.Y,
		}},
		"config": []map[string]interface{}{{
			"concurrently_schedulable_task_count": processor.Component.Config.ConcurrentlySchedulableTaskCount,
			"scheduling_strategy":                 processor.Component.Config.SchedulingStrategy,
			"scheduling_period":                   processor.Component.Config.SchedulingPeriod,
			"properties":                          processor.Component.Config.Properties,
			"auto_terminated_relationships":       relationships,
		}},
	}}
	d.Set("component", component)

	return nil
}
