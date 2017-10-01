package nifi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceRemoteProcessGroup() *schema.Resource {
	return &schema.Resource{
		Create: ResourceRemoteProcessGroupCreate,
		Read:   ResourceRemoteProcessGroupRead,
		Update: ResourceRemoteProcessGroupUpdate,
		Delete: ResourceRemoteProcessGroupDelete,
		Exists: ResourceRemoteProcessGroupExists,

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
						"target_uris": {
							Type:     schema.TypeString,
							Required: true,
						},
						"transportProtocol": {
							Type:     schema.TypeString,
							Required: true,
							Default:  "http",
						},
					},
				},
			},
		},
	}
}

func ResourceRemoteProcessGroupCreate(d *schema.ResourceData, meta interface{}) error {
	processGroup := RemoteProcessGroup{}
	processGroup.Revision.Version = 0

	err := RemoteProcessGroupFromSchema(d, &processGroup)
	if err != nil {
		return fmt.Errorf("Failed to parse Remote Process Group schema")
	}
	parentGroupId := processGroup.Component.ParentGroupId

	client := meta.(*Client)
	err = client.CreateRemoteProcessGroup(&processGroup)
	if err != nil {
		return fmt.Errorf("Failed to create Remote Process Group")
	}

	d.SetId(processGroup.Component.Id)
	d.Set("parent_group_id", parentGroupId)

	return ResourceRemoteProcessGroupRead(d, meta)
}

func ResourceRemoteProcessGroupRead(d *schema.ResourceData, meta interface{}) error {
	processGroupId := d.Id()

	client := meta.(*Client)
	processGroup, err := client.GetRemoteProcessGroup(processGroupId)
	if err != nil {
		return fmt.Errorf("Error retrieving Remote Process Group: %s", processGroupId)
	}

	err = RemoteProcessGroupToSchema(d, processGroup)
	if err != nil {
		return fmt.Errorf("Failed to serialize Remote Process Group: %s", processGroupId)
	}

	return nil
}

func ResourceRemoteProcessGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	processGroupId := d.Id()

	client := meta.(*Client)
	processGroup, err := client.GetRemoteProcessGroup(processGroupId)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Remote Process Group: %s", processGroupId)
		}
	}

	err = RemoteProcessGroupFromSchema(d, processGroup)
	if err != nil {
		return fmt.Errorf("Failed to parse Remote Process Group schema: %s", processGroupId)
	}

	err = client.UpdateRemoteProcessGroup(processGroup)
	if err != nil {
		return fmt.Errorf("Failed to update Remote Process Group: %s", processGroupId)
	}

	return ResourceRemoteProcessGroupRead(d, meta)
}

func ResourceRemoteProcessGroupDelete(d *schema.ResourceData, meta interface{}) error {
	processGroupId := d.Id()
	log.Printf("[INFO] Deleting Remote Process Group: %s", processGroupId)

	client := meta.(*Client)
	processGroup, err := client.GetRemoteProcessGroup(processGroupId)
	if nil != err {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Remote Process Group: %s", processGroupId)
		}
	}

	err = client.DeleteRemoteProcessGroup(processGroup)
	if err != nil {
		return fmt.Errorf("Error deleting Remote Process Group: %s", processGroupId)
	}

	d.SetId("")
	return nil
}

func ResourceRemoteProcessGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	processGroupId := d.Id()

	client := meta.(*Client)
	_, err := client.GetRemoteProcessGroup(processGroupId)
	if nil != err {
		if "not_found" == err.Error() {
			log.Printf("[INFO] Remote Process Group %s no longer exists, removing from state...", processGroupId)
			d.SetId("")
			return false, nil
		} else {
			return false, fmt.Errorf("Error testing existence of Remote Process Group: %s", processGroupId)
		}
	}

	return true, nil
}

// Schema Helpers

func RemoteProcessGroupFromSchema(d *schema.ResourceData, processGroup *RemoteProcessGroup) error {
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

	processGroup.Component.TargetUris = component["targetUris"].(string)
	processGroup.Component.TransportProtocol = component["transportProtocol"].(string)

	return nil
}

func RemoteProcessGroupToSchema(d *schema.ResourceData, processGroup *RemoteProcessGroup) error {
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
		"targetUris":        processGroup.Component.TargetUris,
		"transportProtocol": processGroup.Component.TransportProtocol,
	}}
	d.Set("component", component)

	return nil
}
