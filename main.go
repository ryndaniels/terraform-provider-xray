package main

import (
	"github.com/hashicorp/terraform/plugin"

	"github.com/ryndaniels/terraform-provider-xray/pkg/jfrogxray"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: jfrogxray.Provider,
	})
}
