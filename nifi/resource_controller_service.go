package nifi

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

func ResourceControllerService() *schema.Resource {
	return &schema.Resource{
		Create: ResourceControllerServiceCreate,
		Read:   ResourceControllerServiceRead,
		Update: ResourceControllerServiceUpdate,
		Delete: ResourceControllerServiceDelete,
		Exists: ResourceControllerServiceExists,

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
						"properties": {
							Type:     schema.TypeMap,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func ResourceControllerServiceCreate(d *schema.ResourceData, meta interface{}) error {
	controllerService := ControllerService{}
	controllerService.Revision.Version = 0

	err := ControllerServiceFromSchema(d, &controllerService)
	if err != nil {
		return fmt.Errorf("Failed to parse Controller Service schema")
	}
	parentGroupId := controllerService.Component.ParentGroupId

	client := meta.(*Client)
	err = client.CreateControllerService(&controllerService)
	if err != nil {
		return fmt.Errorf("Failed to create Controller Service")
	}

	err = client.EnableControllerService(&controllerService)
	if nil != err {
		log.Printf("[INFO] Failed to enable Controller Service: %s", controllerService.Component.Id)
	}

	d.SetId(controllerService.Component.Id)
	d.Set("parent_group_id", parentGroupId)

	return ResourceControllerServiceRead(d, meta)
}

func ResourceControllerServiceRead(d *schema.ResourceData, meta interface{}) error {
	controllerServiceId := d.Id()

	client := meta.(*Client)
	controllerService, err := client.GetControllerService(controllerServiceId)
	if err != nil {
		return fmt.Errorf("Error retrieving Controller Service: %s", controllerServiceId)
	}

	err = ControllerServiceToSchema(d, controllerService)
	if err != nil {
		return fmt.Errorf("Failed to serialize Controller Service: %s", controllerServiceId)
	}

	return nil
}

func ResourceControllerServiceUpdate(d *schema.ResourceData, meta interface{}) error {
	controllerServiceId := d.Id()

	client := meta.(*Client)
	controllerService, err := client.GetControllerService(controllerServiceId)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Controller Service: %s", controllerServiceId)
		}
	}

	if "ENABLED" == controllerService.Component.State {
		err = client.DisableControllerService(controllerService)
		if err != nil {
			return fmt.Errorf("Failed to disable Controller Service: %s", controllerService)
		}
	}

	err = ControllerServiceFromSchema(d, controllerService)
	if err != nil {
		return fmt.Errorf("Failed to parse Controller Service schema: %s", controllerServiceId)
	}
	err = client.UpdateControllerService(controllerService)
	if err != nil {
		return fmt.Errorf("Failed to update Controller Service: %s", controllerServiceId)
	}

	err = client.EnableControllerService(controllerService)
	if nil != err {
		log.Printf("[INFO] Failed to enable Controller Service: %s", controllerServiceId)
	}

	return ResourceControllerServiceRead(d, meta)
}

func ResourceControllerServiceDelete(d *schema.ResourceData, meta interface{}) error {
	controllerServiceId := d.Id()
	log.Printf("[INFO] Deleting Controller Service: %s", controllerServiceId)

	client := meta.(*Client)
	controllerService, err := client.GetControllerService(controllerServiceId)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Controller Service: %s", controllerServiceId)
		}
	}

	err = client.DeleteControllerService(controllerService)
	if err != nil {
		return fmt.Errorf("Error deleting Controller Service: %s", controllerServiceId)
	}

	d.SetId("")
	return nil
}

func ResourceControllerServiceExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	controllerServiceId := d.Id()

	client := meta.(*Client)
	_, err := client.GetControllerService(controllerServiceId)
	if nil != err {
		if "not_found" == err.Error() {
			log.Printf("[INFO] Controller Service %s no longer exists, removing from state...", controllerServiceId)
			d.SetId("")
			return false, nil
		} else {
			return false, fmt.Errorf("Error testing existence of Controller Service: %s", controllerServiceId)
		}
	}

	return true, nil
}

// Schema Helpers

func ControllerServiceFromSchema(d *schema.ResourceData, controllerService *ControllerService) error {
	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})
	controllerService.Component.ParentGroupId = component["parent_group_id"].(string)
	controllerService.Component.Name = component["name"].(string)
	controllerService.Component.Type = component["type"].(string)

	controllerService.Component.Properties = map[string]interface{}{}
	properties := component["properties"].(map[string]interface{})
	for k, v := range properties {
		controllerService.Component.Properties[k] = v.(string)
	}
	return nil
}

func ControllerServiceToSchema(d *schema.ResourceData, controllerService *ControllerService) error {
	revision := []map[string]interface{}{{
		"version": controllerService.Revision.Version,
	}}
	d.Set("revision", revision)

	component := []map[string]interface{}{{
		"parent_group_id": d.Get("parent_group_id").(string),
		"name":            controllerService.Component.Name,
		"type":            controllerService.Component.Type,
		"properties":      controllerService.Component.Properties,
	}}
	d.Set("component", component)

	return nil
}
