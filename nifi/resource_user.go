package nifi

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

func ResourceUser() *schema.Resource {
	log.Println("ResourceUser")
	return &schema.Resource{
		Create: ResourceUserCreate,
		Read:   ResourceUserRead,
		Update: ResourceUserUpdate,
		Delete: ResourceUserDelete,
		Exists: ResourceUserExists,

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
							Required: false,
						},
						"identity": {
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

func ResourceUserCreate(d *schema.ResourceData, meta interface{}) error {
	log.Println("ResourceUserCreate")
	user := User{}
	user.Revision.Version = 0

	err := UserFromSchema(d, &user)
	if err != nil {
		return fmt.Errorf("Failed to parse User schema")
	}
	parentGroupId := user.Component.ParentGroupId

	// Create user
	client := meta.(*Client)
	err = client.CreateUser(&user)
	if err != nil {
		return fmt.Errorf("Failed to create Connection")
	}

	// Indicate successful creation
	d.SetId(user.Component.Id)
	d.Set("parent_group_id", parentGroupId)

	return ResourceUserRead(d, meta)
}

func ResourceUserRead(d *schema.ResourceData, meta interface{}) error {
	log.Println("ResourceUserRead")
	userId := d.Id()

	client := meta.(*Client)
	user, err := client.GetUser(userId)
	if err != nil {
		return fmt.Errorf("Error retrieving User: %s", userId)
	}

	err = UserToSchema(d, user)
	if err != nil {
		return fmt.Errorf("Failed to serialize User: %s", userId)
	}

	return nil
}

func ResourceUserUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Println("ResourceUserUpdate")
	client := meta.(*Client)
	client.Lock.Lock()
	log.Printf("[INFO] Updating User: %s..., not implemented", d.Id())
	log.Printf("[INFO] User updated: %s", d.Id())
	defer client.Lock.Unlock()
	return nil
}

func ResourceUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	client.Lock.Lock()
	log.Printf("[INFO] Deleting User: %s...", d.Id())
	err := ResourceUserDeleteInternal(d, meta)
	log.Printf("[INFO] User deleted: %s", d.Id())
	defer client.Lock.Unlock()
	return err
}

func ResourceUserDeleteInternal(d *schema.ResourceData, meta interface{}) error {
	userId := d.Id()

	// Refresh user details
	client := meta.(*Client)
	user, err := client.GetUser(userId)
	if err != nil {
		if "not_found" == err.Error() {
			d.SetId("")
			return nil
		} else {
			return fmt.Errorf("Error retrieving User: %s", userId)
		}
	}

	// Delete user
	err = client.DeleteUser(user)
	if err != nil {
		return fmt.Errorf("Error deleting User: %s", userId)
	}

	d.SetId("")
	return nil
}

func ResourceUserExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	log.Println("ResourceUserExists")
	userId := d.Id()

	client := meta.(*Client)
	_, err := client.GetUser(userId)
	if nil != err {
		if "not_found" == err.Error() {
			log.Printf("[INFO] User %s no longer exists, removing from state...", userId)
			d.SetId("")
			return false, nil
		} else {
			return false, fmt.Errorf("Error testing existence of User: %s", userId)
		}
	}

	return true, nil
}

// Schema Helpers

func UserFromSchema(d *schema.ResourceData, user *User) error {
	log.Println("ResourceUserFromSchema")
	v := d.Get("component").([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component is required")
	}
	component := v[0].(map[string]interface{})
	user.Component.ParentGroupId = component["parent_group_id"].(string)
	user.Component.Identity = component["identity"].(string)

	v = component["position"].([]interface{})
	if len(v) != 1 {
		return fmt.Errorf("Exactly one component.position is required")
	}
	position := v[0].(map[string]interface{})
	user.Component.Position.X = position["x"].(float64)
	user.Component.Position.Y = position["y"].(float64)

	return nil
}

func UserToSchema(d *schema.ResourceData, user *User) error {
	log.Println("ResourceUserToSchema")
	revision := []map[string]interface{}{{
		"version": user.Revision.Version,
	}}
	d.Set("revision", revision)
	component := []map[string]interface{}{{
		"parent_group_id": interface{}(user.Component.ParentGroupId).(string),
		"position": []map[string]interface{}{{
			"x": user.Component.Position.X,
			"y": user.Component.Position.Y,
		}},
		"identity": interface{}(user.Component.Identity).(string),
	}}
	d.Set("component", component)
	return nil
}
