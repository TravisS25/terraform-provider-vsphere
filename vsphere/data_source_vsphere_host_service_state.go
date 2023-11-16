// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostservicestate"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
)

func dataSourceVSphereHostServiceState() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereHostServiceStateRead,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Host id of machine",
			},
			"service": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The service state object",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Key for service",
						},
						"running": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "State of the service",
						},
						"policy": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Policy of the service",
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereHostServiceStateRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] entering data_source_vsphere_host_service_state read function")

	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	hsList, err := hostservicestate.GetHostServies(client, hostID, provider.DefaultAPITimeout)
	if err != nil {
		return fmt.Errorf("error retrieving host services for host '%s': %s", hostID, err)
	}

	srvList := make([]interface{}, 0, len(hsList))

	for _, hs := range hsList {
		foundExcludeKey := false

		for _, excludeKey := range hostservicestate.ExcludeServiceKeyList {
			if hs.Key == excludeKey {
				foundExcludeKey = true
			}
		}

		if !foundExcludeKey {
			srvList = append(srvList, map[string]interface{}{
				"key":     hs.Key,
				"policy":  hs.Policy,
				"running": hs.Running,
			})
		}
	}

	d.SetId(hostID)
	d.Set("host_system_id", hostID)
	d.Set("service", srvList)

	return nil
}
