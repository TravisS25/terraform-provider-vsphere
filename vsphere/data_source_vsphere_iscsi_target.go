// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/iscsi"
)

func dataSourceVSphereIscsiTarget() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereIscsiTargetRead,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Description: "The host to gather iscsi information",
				Required:    true,
			},
			"adapter_id": {
				Type:        schema.TypeString,
				Description: "The iscsi adapter id of the host",
				Required:    true,
			},
			"send_target": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						iscsi.IPResourceKey:   iscsi.IPSchema(),
						iscsi.PortResourceKey: iscsi.PortSchema(),
					},
				},
			},
			"static_target": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						iscsi.IPResourceKey:   iscsi.IPSchema(),
						iscsi.PortResourceKey: iscsi.PortSchema(),
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The iqn of the storage device",
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereIscsiTargetRead(d *schema.ResourceData, meta interface{}) error {
	hostID := d.Get("host_system_id").(string)
	adapterID := d.Get("adapter_id").(string)

	err := iscsiTargetRead(
		d,
		meta,
		hostID,
		adapterID,
		true,
	)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s:%s", hostID, adapterID))
	return nil
}
