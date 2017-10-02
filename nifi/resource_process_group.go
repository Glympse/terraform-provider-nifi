package nifi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceProcessGroup() *schema.Resource {
	return &schema.Resource{
		Create: ResourceProcessGroupCreate,
		Read:   ResourceProcessGroupRead,
		Update: ResourceProcessGroupUpdate,
		Delete: ResourceProcessGroupDelete,
		Exists: ResourceProcessGroupExists,

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
						"position": SchemaPosition(),
					},
				},
			},
		},
	}
}

func ResourceProcessGroupCreate(d *schema.ResourceData, meta interface{}) error {
	processGroup := ProcessGroup{}
	processGroup.Revision.Version = 0

	err := ProcessGroupFromSchema(d, &processGroup)
	if err != nil {
		return fmt.Errorf("Failed to parse Process Group schema")
	}
	parentGroupId := processGroup.Component.ParentGroupId

	client := meta.(*Client)
	err = client.CreateProcessGroup(&processGroup)
	if err != nil {
		return fmt.Errorf("Failed to create Process Group")
	}

	d.SetId(processGroup.Component.Id)
	d.Set("parent_group_id", parentGroupId)

	return ResourceProcessGroupRead(d, meta)
}

func ResourceProcessGroupRead(d *schema.ResourceData, meta interface{}) error {
	processGroupId := d.Id()

	client := meta.(*Client)
	processGroup, err := client.GetProcessGroup(processGroupId)
	if err != nil {
		return fmt.Errorf("Error retrieving Process Group: %s", processGroupId)
	}

	err = ProcessGroupToSchema(d, processGroup)
	if err != nil {
		return fmt.Errorf("Failed to serialize Process Group: %s", processGroupId)
	}

	return nil
}

func ResourceProcessGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	processGroupId := d.Id()

	client := meta.(*Client)
	processGroup, err := client.GetProcessGroup(processGroupId)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Process Group: %s", processGroupId)
		}
	}

	err = ProcessGroupFromSchema(d, processGroup)
	if err != nil {
		return fmt.Errorf("Failed to parse Process Group schema: %s", processGroupId)
	}

	err = client.UpdateProcessGroup(processGroup)
	if err != nil {
		return fmt.Errorf("Failed to update Process Group: %s", processGroupId)
	}

	return ResourceProcessGroupRead(d, meta)
}

func ResourceProcessGroupDelete(d *schema.ResourceData, meta interface{}) error {
	processGroupId := d.Id()
	log.Printf("[INFO] Deleting Process Group: %s", processGroupId)

	client := meta.(*Client)
	processGroup, err := client.GetProcessGroup(processGroupId)
	if nil != err {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Process Group: %s", processGroupId)
		}
	}

	err = client.DeleteProcessGroup(processGroup)
	if err != nil {
		return fmt.Errorf("Error deleting Process Group: %s", processGroupId)
	}

	d.SetId("")
	return nil
}

func ResourceProcessGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	processGroupId := d.Id()

	client := meta.(*Client)
	_, err := client.GetProcessGroup(processGroupId)
	if nil != err {
		if "not_found" == err.Error() {
			log.Printf("[INFO] Process Group %s no longer exists, removing from state...", processGroupId)
			d.SetId("")
			return false, nil
		} else {
			return false, fmt.Errorf("Error testing existence of Process Group: %s", processGroupId)
		}
	}

	return true, nil
}

// Schema Helpers

func ProcessGroupFromSchema(d *schema.ResourceData, processGroup *ProcessGroup) error {
	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})

	parentGroupId := component["parent_group_id"].(string)
	processGroup.Component.ParentGroupId = parentGroupId
	processGroup.Component.Name = component["name"].(string)

	v = component["position"].([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component.position is required")
	}
	position := v[0].(map[string]interface{})
	processGroup.Component.Position.X = position["x"].(float64)
	processGroup.Component.Position.Y = position["y"].(float64)

	return nil
}

func ProcessGroupToSchema(d *schema.ResourceData, processGroup *ProcessGroup) error {
	revision := []map[string]interface{}{{
		"version": processGroup.Revision.Version,
	}}
	d.Set("revision", revision)

	component := []map[string]interface{}{{
		"parent_group_id": d.Get("parent_group_id").(string),
		"name":            processGroup.Component.Name,
		"position": []map[string]interface{}{{
			"x": processGroup.Component.Position.X,
			"y": processGroup.Component.Position.Y,
		}},
	}}
	d.Set("component", component)

	return nil
}
