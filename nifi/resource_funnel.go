package nifi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceFunnel() *schema.Resource {
	return &schema.Resource{
		Create: ResourceFunnelCreate,
		Read:   ResourceFunnelRead,
		Update: ResourceFunnelUpdate,
		Delete: ResourceFunnelDelete,
		Exists: ResourceFunnelExists,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				//d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

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
						"position": SchemaPosition(),
					},
				},
			},
		},
	}
}

func ResourceFunnelCreate(d *schema.ResourceData, meta interface{}) error {
	funnel := FunnelStub()
	funnel.Revision.Version = 0

	err := FunnelFromSchema(meta, d, funnel)
	if err != nil {
		return fmt.Errorf("Failed to parse User schema")
	}
	parentGroupId := funnel.Component.ParentGroupId

	// Create user
	client := meta.(*Client)
	err = client.CreateGroup(funnel)
	if err != nil {
		return fmt.Errorf("Failed to create Connection")
	}

	// Indicate successful creation
	d.SetId(funnel.Component.Id)
	d.Set("parent_group_id", parentGroupId)

	return ResourceFunnelRead(d, meta)
}

func ResourceFunnelRead(d *schema.ResourceData, meta interface{}) error {
	funnelIId := d.Id()

	client := meta.(*Client)
	funnel, err := client.GetFunnel(funnelIId)
	if err != nil {
		return fmt.Errorf("Error retrieving Group: %s", funnelIId)
	}

	err = FunnelToSchema(d, funnel)
	if err != nil {
		return fmt.Errorf("Failed to serialize Group: %s", funnelIId)
	}

	return nil
}

func ResourceFunnelUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	client.Lock.Lock()
	err := ResourceFunnelUpdateInternal(d, meta)
	if err == nil {
		log.Printf("[INFO] Funnel updated: %s", d.Id())
	} else {
		log.Printf("[ERROR] Funnel Update failed: %s", d.Id())
	}
	defer client.Lock.Unlock()
	return err
}
func ResourceFunnelUpdateInternal(d *schema.ResourceData, meta interface{}) error {
	funnelId := d.Id()

	// Refresh funnel details
	client := meta.(*Client)
	funnel, err := client.GetFunnel(funnelId)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Funnel: %s", funnelId)
		}
	}

	// Load funnel's desired state
	err = FunnelFromSchema(meta, d, funnel)
	if err != nil {
		return fmt.Errorf("Failed to parse Funnel schema: %s", funnelId)
	}

	// Update funnel
	err = client.UpdateFunnel(funnel)
	if err != nil {
		return fmt.Errorf("Failed to update Funnel: %s", funnelId)
	}

	return ResourceGroupRead(d, meta)
}

func ResourceFunnelDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	client.Lock.Lock()
	log.Printf("[INFO] Deleting Funnel: %s...", d.Id())
	err := ResourceFunnelDeleteInternal(d, meta)
	if err == nil {
		log.Printf("[INFO] Funnel deleted: %s", d.Id())
	}else {
		log.Printf("[ERROR] Funnel deletion failed: %s", d.Id())
	}
	defer client.Lock.Unlock()
	return err
}

func ResourceFunnelDeleteInternal(d *schema.ResourceData, meta interface{}) error {
	funnelId := d.Id()

	// Refresh funnel details
	client := meta.(*Client)
	funnel, err := client.GetFunnel(funnelId)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving Funnel: %s", funnelId)
		}
	}

	// Delete funnel
	err = client.DeleteFunnel(funnel)
	if err != nil {
		return fmt.Errorf("Error deleting Funnel: %s", funnelId)
	}

	d.SetId("")
	return nil
}

func ResourceFunnelExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	funnelId := d.Id()
	client := meta.(*Client)
	_, err := client.GetFunnel(funnelId)
	if nil != err {
		if "not_found" == err.Error() {
			log.Printf("[INFO] Funnel %s no longer exists, removing from state...", funnelId)
			d.SetId("")
			return false, nil
		} else {
			return false, fmt.Errorf("Error testing existence of Funnel: %s", funnelId)
		}
	}
	return true, nil
}

// Schema Helpers

func FunnelFromSchema(meta interface{}, d *schema.ResourceData, funnel *Funnel) error {
	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})
	funnel.Component.ParentGroupId = component["parent_group_id"].(string)

	v = component["position"].([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component.position is required")
	}
	position := v[0].(map[string]interface{})
	funnel.Component.Position.X = position["x"].(float64)
	funnel.Component.Position.Y = position["y"].(float64)

	return nil
}

func FunnelToSchema(d *schema.ResourceData, funnel *Funnel) error {
	revision := []map[string]interface{}{{
		"version": funnel.Revision.Version,
	}}
	d.Set("revision", revision)

	component := []map[string]interface{}{{
		"parent_group_id": interface{}(funnel.Component.ParentGroupId).(string),
		"position": []map[string]interface{}{{
			"x": funnel.Component.Position.X,
			"y": funnel.Component.Position.Y,
		}},
	}}
	d.Set("component", component)

	return nil
}
