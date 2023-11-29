// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/vmware/govmomi/vim25/mo"
)

func dataSourceVSphereHostConfigDateTime() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereHostConfigDateTimeRead,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Description: "Host id of machine to gather ntp info",
				Required:    true,
			},
			"ntp_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Set ntp config settings",
				//MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "List of ntp servers",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"protocol": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gathers network time configuration for clock",
			},
			"events_disabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Gathers whether events are disabled or not",
			},
			"fallback_disabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Gathers whether fallback to ntp is enabled",
			},
		},
	}
}

func dataSourceVSphereHostConfigDateTimeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return err
	}

	hostDt, err := host.ConfigManager().DateTimeSystem(context.Background())
	if err != nil {
		return fmt.Errorf("error trying to get datetime system object from host '%s': %s", hostID, err)
	}

	var hostDtProps mo.HostDateTimeSystem
	if err = hostDt.Properties(context.Background(), hostDt.Reference(), nil, &hostDtProps); err != nil {
		return fmt.Errorf("error trying to gather datetime properties from host '%s': %s", hostID, err)
	}

	ntpCfg := []interface{}{
		map[string]interface{}{
			"server": hostDtProps.DateTimeInfo.NtpConfig.Server,
		},
	}

	d.SetId(hostID)
	d.Set("host_system_id", hostID)
	d.Set("protocol", hostDtProps.DateTimeInfo.SystemClockProtocol)
	d.Set("events_disabled", hostDtProps.DateTimeInfo.DisableEvents)
	d.Set("fallback_disabled", hostDtProps.DateTimeInfo.DisableFallback)
	d.Set("ntp_config", ntpCfg)
	return nil
}
