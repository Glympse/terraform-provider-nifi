package nifi

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceNifiProcessGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNifiProcessGroupCreate,
		Read:   resourceNifiProcessGroupRaad,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNifiProcessGroupCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceNifiProcessGroupRaad(d *schema.ResourceData, meta interface{}) error {
	return nil
}
