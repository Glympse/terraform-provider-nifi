package main

import (
	"github.com/af6140/terraform-provider-nifi/nifi"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: nifi.Provider,
	})
}
