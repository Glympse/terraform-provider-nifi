package nifi

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NIFI_HOST", nil),
			},

			"api_path": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NIFI_API_PATH", "nifi-api"),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"nifi_processor":  ResourceProcessor(),
			"nifi_connection": ResourceConnection(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Host:    d.Get("host").(string),
		ApiPath: d.Get("api_path").(string),
	}
	client := NewClient(config)
	return client, nil
}
