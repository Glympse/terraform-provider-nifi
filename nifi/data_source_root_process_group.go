package nifi

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceRootProcessGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRootProcessGroupRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceRootProcessGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	processGroup, err := client.GetProcessGroup("root")
	if err != nil {
		return fmt.Errorf("Error retrieving root Process Group")
	}
	d.SetId(processGroup.Component.Id)
	d.Set("name", processGroup.Component.Name)
	return nil
}
