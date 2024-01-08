// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostconfig"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
)

func dataSourceVSphereHostConfigSyslog() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereHostConfigSyslogRead,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Host id of machine to get syslog info from",
				ExactlyOneOf: []string{"hostname"},
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Hostname of machine to get syslog info from",
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
	client := meta.(*Client).vimClient
	host, tfID, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] reading syslog settings from data source for host '%s'", host.Name())

	if err = hostconfig.HostConfigSyslogRead(d, client, host); err != nil {
		return err
	}

	d.SetId(tfID)
	return nil
}
