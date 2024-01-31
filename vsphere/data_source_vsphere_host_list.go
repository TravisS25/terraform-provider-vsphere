// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

func dataSourceVSphereHostList() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereHostListRead,

		Schema: map[string]*schema.Schema{
			"datacenter_id": {
				Type:        schema.TypeString,
				Description: "The managed object ID of the datacenter to look for all hosts",
				Required:    true,
			},
			"hosts": {
				Type:        schema.TypeSet,
				Description: "The hosts from the given datacenter_id",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_system_id": {
							Type:        schema.TypeString,
							Description: "The host id of host from given datacenter_id",
							Computed:    true,
						},
						"hostname": {
							Type:        schema.TypeString,
							Description: "The hostname of host from given datacenter_id",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereHostListRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient

	dcID := d.Get("datacenter_id").(string)
	dc, err := datacenterFromID(client, dcID)
	if err != nil {
		return fmt.Errorf("error fetching datacenter: %s", err)
	}

	// Create a view manager
	m := view.NewManager(client.Client)

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	// Create a view for hosts
	view, err := m.CreateContainerView(ctx, dc.Reference(), []string{"HostSystem"}, true)
	if err != nil {
		return fmt.Errorf("error trying to create view for dc: %s", err)
	}

	defer func() {
		_ = view.Destroy(ctx)
	}()

	var moHosts []mo.HostSystem
	if err = view.Retrieve(ctx, []string{"HostSystem"}, nil, &moHosts); err != nil {
		return fmt.Errorf("error retrieving hosts for dc '%s': %s", dc.Name(), err)
	}

	hosts := make([]interface{}, 0, len(moHosts))
	for _, host := range moHosts {
		hosts = append(hosts, map[string]interface{}{
			"host_system_id": host.Reference().Value,
			"hostname":       host.Name,
		})
	}

	d.SetId(dc.Reference().Value)
	d.Set("hosts", hosts)
	return nil
}
