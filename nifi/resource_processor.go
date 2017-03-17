package nifi

import (
	"fmt"
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
						"position": {
							Type:     schema.TypeList,
							Required: true,
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
						"config": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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

	_, err := client.CreateProcessor(&processor)
	if err != nil {
		return err
	}

	d.SetId(processor.Component.Id)

	return ResourceProcessorRead(d, meta)
}

func ResourceProcessorRead(d *schema.ResourceData, meta interface{}) error {
	// TODO:
	return nil
}

func ResourceProcessorUpdate(d *schema.ResourceData, meta interface{}) error {
	// TODO:
	return nil
}

func ResourceProcessorDelete(d *schema.ResourceData, meta interface{}) error {
	// TODO:
	return nil
}

func ResourceProcessorExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	// TODO:
	return false, nil
}
