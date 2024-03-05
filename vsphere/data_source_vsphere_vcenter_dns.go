// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVSphereVcenterDNS() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereVcenterDNSRead,
		Schema: map[string]*schema.Schema{
			"servers": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The dns servers from vcenter",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceVSphereVcenterDNSRead(d *schema.ResourceData, meta interface{}) error {
	err := vsphereVcenterDNSRead(d, meta)
	if err != nil {
		return err
	}

	d.SetId(vsphereVcenterDnsID)
	return nil
}
