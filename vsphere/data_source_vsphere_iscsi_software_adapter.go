// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVSphereIscsiSoftwareAdapter() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereIscsiSoftwareAdapterRead,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Description: "The host to gather iscsi information",
				Required:    true,
			},
			"iscsi_name": {
				Type:        schema.TypeString,
				Description: "The host to gather iscsi information",
				Computed:    true,
			},
		},
	}
}

func dataSourceVSphereIscsiSoftwareAdapterRead(d *schema.ResourceData, meta interface{}) error {
	err := iscsiSoftwareAdapterRead(d, meta, true)
	if err != nil {
		return err
	}

	d.SetId(d.Get("host_system_id").(string))
	return nil
}
