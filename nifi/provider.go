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
			"admin_cert": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NIFI_ADMIN_CERT", nil),
			},
			"admin_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NIFI_ADMIN_KEY", nil),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"nifi_process_group":      ResourceProcessGroup(),
			"nifi_processor":          ResourceProcessor(),
			"nifi_connection":         ResourceConnection(),
			"nifi_controller_service": ResourceControllerService(),
			"nifi_user":               ResourceUser(),
			"nifi_group":              ResourceGroup(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Host:          d.Get("host").(string),
		ApiPath:       d.Get("api_path").(string),
		AdminCertPath: d.Get("admin_cert").(string),
		AdminKeyPath:  d.Get("admin_key").(string),
	}
	client := NewClient(config)
	return client, nil
}
