package nifi

import (
	"github.com/hashicorp/terraform/helper/schema"
	"fmt"
)

func ResourceConnection() *schema.Resource {
	return &schema.Resource{
		Create: ResourceConnectionCreate,
		Read:   ResourceConnectionRead,
		Update: ResourceConnectionUpdate,
		Delete: ResourceConnectionDelete,
		Exists: ResourceConnectionExists,

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
							Elem:	  &schema.Resource{
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
	client := meta.(*Client)

	connection := Connection{}
	connection.Revision.Version = 0

	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})
	connection.Component.ParentGroupId = component["parent_group_id"].(string)

	v = component["source"].([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component.source is required")
	}
	source := v[0].(map[string]interface{})
	connection.Component.Source.Type = source["type"].(string)
	connection.Component.Source.Id = source["id"].(string)

	v = component["destination"].([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component.destination is required")
	}
	destination := v[0].(map[string]interface{})
	connection.Component.Destination.Type = destination["type"].(string)
	connection.Component.Destination.Id = destination["id"].(string)

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

	_, err := client.CreateConnection(&connection)
	if err != nil {
		return err
	}

	d.SetId(connection.Component.Id)

	return ResourceConnectionRead(d, meta)
}

func ResourceConnectionRead(d *schema.ResourceData, meta interface{}) error {
	// TODO:
	return nil
}

func ResourceConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	// TODO:
	return nil
}

func ResourceConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	// TODO:
	return nil
}

func ResourceConnectionExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	// TODO:
	return false, nil
}
