package nifi

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
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
						"parent_group_id": SchemaParentGroupId(),
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
	client := meta.(*Client)

	processor := Processor{}
	processor.Revision.Version = 0

	err := ProcessorFromSchema(d, &processor)
	if err != nil {
		return fmt.Errorf("Failed to parse Processor schema")
	}

	err = client.CreateProcessor(&processor)
	if err != nil {
		return fmt.Errorf("Failed to create Processor")
	}

	d.SetId(processor.Component.Id)
	d.Set("parent_group_id", processor.Component.ParentGroupId)

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
	processorId := d.Id()

	client := meta.(*Client)
	processor, err := client.GetProcessor(processorId)
	if err != nil {
		return fmt.Errorf("Error retrieving Processor: %s", processorId)
	}

	err = ProcessorFromSchema(d, processor)
	if err != nil {
		return fmt.Errorf("Failed to parse Processor schema: %s", processorId)
	}

	err = client.UpdateProcessor(processor)
	if err != nil {
		return fmt.Errorf("Failed to update Processor: %s", processorId)
	}

	return ResourceProcessorRead(d, meta)
}

func ResourceProcessorDelete(d *schema.ResourceData, meta interface{}) error {
	processorId := d.Id()
	log.Printf("[INFO] Deleting Processor: %s", processorId)

	client := meta.(*Client)
	err := client.DeleteProcessor(processorId)
	if err != nil {
		return fmt.Errorf("Error deleting Processor: %s", processorId)
	}

	d.SetId("")
	return nil
}

func ResourceProcessorExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	processorId := d.Id()

	client := meta.(*Client)
	processor, err := client.GetProcessor(processorId)
	if err != nil {
		return false, fmt.Errorf("Error testing existence of Processor: %s", processorId)
	}

	exists := nil != processor
	if !exists {
		log.Printf("[INFO] Processor %s no longer exists, removing from state...", processorId)
		d.SetId("")
	}

	return exists, nil
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

	processor.Component.Config.Properties = map[string]string{}
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
	// TODO:
	return nil
}
