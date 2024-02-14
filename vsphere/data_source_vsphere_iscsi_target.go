// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/iscsi"
)

func dataSourceVSphereIscsiTarget() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereIscsiTargetRead,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:         schema.TypeString,
				Description:  "The host id to gather iscsi information",
				Optional:     true,
				ExactlyOneOf: []string{"hostname"},
			},
			"hostname": {
				Type:        schema.TypeString,
				Description: "The hostname to gather iscsi information",
				Optional:    true,
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
	client := meta.(*Client).vimClient
	host, hr, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host on iscsi data source read: %s", err)
	}

	adapterID := d.Get("adapter_id").(string)

	if err = iscsiTargetRead(client, d, host, adapterID, true); err != nil {
		return fmt.Errorf("error reading iscsi target properties on data source read for host '%s': %s", host.Name(), err)
	}
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s:%s", hr.IDName, adapterID))
	return nil
}
