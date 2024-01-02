// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostconfig"
)

func dataSourceVSphereHostConfigSyslog() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereHostConfigSyslogRead,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Host id of machine to get syslog info from",
			},
			"log_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The remote host to output logs to",
			},
			"log_level": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The log level to output logs",
			},
		},
	}
}

func dataSourceVSphereHostConfigSyslogRead(d *schema.ResourceData, meta interface{}) error {
	hostID := d.Get("host_system_id").(string)

	log.Printf("[INFO] reading syslog options from data source for host '%s'", hostID)

	err := hostconfig.HostConfigSyslogRead(context.Background(), d, meta.(*Client).vimClient, hostID)
	if err != nil {
		return err
	}

	d.SetId(hostID)
	d.Set("host_system_id", hostID)

	return nil
}
