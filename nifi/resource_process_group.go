package nifi

import (
	"github.com/hashicorp/terraform/helper/schema"
	"fmt"
)

func ResourceProcessGroup() *schema.Resource {
	return &schema.Resource{
		Create: ResourceProcessGroupCreate,
		Read:   ResourceProcessGroupRead,
		Update: ResourceProcessGroupUpdate,
		Delete: ResourceProcessGroupDelete,
		Exists: ResourceProcessGroupExists,

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
					},
				},
			},
		},
	}
}

func ResourceProcessGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	processGroup := ProcessGroup{}
	processGroup.Revision.Version = 0

	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})
	processGroup.Component.ParentGroupId = component["parent_group_id"].(string)
	processGroup.Component.Name = component["name"].(string)

	v = component["position"].([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component.position is required")
	}
	position := v[0].(map[string]interface{})
	processGroup.Component.Position.X = position["x"].(float64)
	processGroup.Component.Position.Y = position["y"].(float64)

	_, err := client.CreateProcessGroup(&processGroup)
	if err != nil {
		return err
	}

	d.SetId(processGroup.Component.Id)

	return ResourceProcessGroupRead(d, meta)
}

func ResourceProcessGroupRead(d *schema.ResourceData, meta interface{}) error {
	// TODO:
	return nil
}

func ResourceProcessGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	// TODO:
	return nil
}

func ResourceProcessGroupDelete(d *schema.ResourceData, meta interface{}) error {
	// TODO:
	return nil
}

func ResourceProcessGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	// TODO:
	return false, nil
}
