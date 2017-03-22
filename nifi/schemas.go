package nifi

import "github.com/hashicorp/terraform/helper/schema"

func SchemaParentGroupId() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
}

func SchemaRevision() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"version": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			},
		},
	}
}

func SchemaPosition() *schema.Schema {
	return &schema.Schema{
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
	}
}
