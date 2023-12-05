// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"log"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func resourceVSphereHostConfigDateTime() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereHostConfigDateTimeCreate,
		Read:   resourceVSphereHostConfigDateTimeRead,
		Update: resourceVSphereHostConfigDateTimeUpdate,
		Delete: resourceVSphereHostConfigDateTimeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVSphereHostConfigDateTimeImport,
		},
		CustomizeDiff: resourceVSphereHostConfigDateTimeCustomDiff,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Host id of machine to configure ntp",
			},
			"ntp_servers": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of ntp servers to use",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"protocol": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     string(types.HostDateTimeInfoProtocolNtp),
				Description: "Specify which network time configuration to discipline vmkernel clock",
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(types.HostDateTimeInfoProtocolNtp),
						string(types.HostDateTimeInfoProtocolPtp),
					},
					true,
				),
			},
			"disable_events": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disables detected failures being sent to VCenter if set",
			},
			"disable_fallback": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disables falling back to ntp if ptp fails when set",
			},
		},
	}
}

func resourceVSphereHostConfigDateTimeRead(d *schema.ResourceData, meta interface{}) error {
	hostID := d.Get("host_system_id").(string)
	log.Printf("[INFO] reading date time configuration for host '%s'", hostID)
	return hostConfigDateTimeRead(d, meta, hostID)
}

func resourceVSphereHostConfigDateTimeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return err
	}

	hostDt, err := host.ConfigManager().DateTimeSystem(context.Background())
	if err != nil {
		return fmt.Errorf("error trying to get date time system object from host '%s': %s", hostID, err)
	}

	disableEvents := d.Get("disable_events").(bool)
	disableFallback := d.Get("disable_fallback").(bool)
	enabled := true
	cfg := types.HostDateTimeConfig{
		Enabled:         &enabled,
		Protocol:        d.Get("protocol").(string),
		DisableEvents:   &disableEvents,
		DisableFallback: &disableFallback,
	}

	ntpServerList := d.Get("ntp_servers").(*schema.Set).List()

	if len(ntpServerList) > 0 {
		servers := make([]string, 0, len(ntpServerList))

		for _, v := range ntpServerList {
			servers = append(servers, v.(string))
		}

		cfg.NtpConfig = &types.HostNtpConfig{
			Server: servers,
		}
	}

	log.Printf("[INFO] creating date time configuration for host '%s'", hostID)

	if err = hostDt.UpdateConfig(context.Background(), cfg); err != nil {
		return fmt.Errorf("error trying to create date time configuration for host '%s': %s", hostID, err)
	}

	d.SetId(hostID)

	return nil
}

func resourceVSphereHostConfigDateTimeUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return err
	}

	hostDt, err := host.ConfigManager().DateTimeSystem(context.Background())
	if err != nil {
		return fmt.Errorf("error trying to get date time system object from host '%s': %s", hostID, err)
	}

	var hostDtProps mo.HostDateTimeSystem
	if err = hostDt.Properties(context.Background(), hostDt.Reference(), nil, &hostDtProps); err != nil {
		return fmt.Errorf("error trying to gather date time properties from host '%s': %s", hostID, err)
	}

	disableEvents := d.Get("disable_events").(bool)
	disableFallback := d.Get("disable_fallback").(bool)
	enabled := true
	cfg := types.HostDateTimeConfig{
		Enabled:         &enabled,
		Protocol:        d.Get("protocol").(string),
		DisableEvents:   &disableEvents,
		DisableFallback: &disableFallback,
	}

	if d.HasChange("ntp_servers") {
		_, newValue := d.GetChange("ntp_servers")
		newList := newValue.(*schema.Set).List()

		if len(newList) > 0 {
			servers := make([]string, 0, len(newList))

			for _, v := range newList {
				servers = append(servers, v.(string))
			}

			cfg.NtpConfig = &types.HostNtpConfig{
				Server: servers,
			}
		}
	}

	log.Printf("[INFO] updating date time configuration for host '%s'", hostID)

	if err = hostDt.UpdateConfig(context.Background(), cfg); err != nil {
		return fmt.Errorf("error trying to update date time configuration for host '%s': %s", hostID, err)
	}

	return nil
}

func resourceVSphereHostConfigDateTimeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return err
	}

	hostDt, err := host.ConfigManager().DateTimeSystem(context.Background())
	if err != nil {
		return fmt.Errorf("error trying to get date time system object from host '%s': %s", hostID, err)
	}

	log.Printf("[INFO] deleting date time configuration for host '%s'", hostID)

	factoryDefaults := true
	if err = hostDt.UpdateConfig(context.Background(), types.HostDateTimeConfig{
		Protocol:               string(types.HostDateTimeInfoProtocolNtp),
		ResetToFactoryDefaults: &factoryDefaults,
	}); err != nil {
		return fmt.Errorf("error trying to delete date time configuration for host '%s': %s", hostID, err)
	}

	return nil
}

func resourceVSphereHostConfigDateTimeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	hostID := d.Id()
	log.Printf("[INFO] importing date time configuration for host '%s'", hostID)
	err := hostConfigDateTimeRead(d, meta, hostID)
	if err != nil {
		return nil, err
	}

	d.SetId(hostID)
	d.Set("host_system_id", hostID)
	return []*schema.ResourceData{d}, nil
}

func resourceVSphereHostConfigDateTimeCustomDiff(ctx context.Context, rd *schema.ResourceDiff, meta interface{}) error {
	ntpServers := rd.Get("ntp_servers").(*schema.Set).List()

	if len(ntpServers) == 0 {
		return fmt.Errorf("'ntp_servers' parameter is required")
	}

	return nil
}

func hostConfigDateTimeRead(d *schema.ResourceData, meta interface{}, hostID string) error {
	client := meta.(*Client).vimClient
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return err
	}

	hostDt, err := host.ConfigManager().DateTimeSystem(context.Background())
	if err != nil {
		return fmt.Errorf("error trying to get date time system object from host '%s': %s", hostID, err)
	}

	var hostDtProps mo.HostDateTimeSystem
	if err = hostDt.Properties(context.Background(), hostDt.Reference(), nil, &hostDtProps); err != nil {
		return fmt.Errorf("error trying to gather date time properties from host '%s': %s", hostID, err)
	}

	d.Set("ntp_servers", hostDtProps.DateTimeInfo.NtpConfig.Server)
	d.Set("protocol", hostDtProps.DateTimeInfo.SystemClockProtocol)
	d.Set("disable_events", hostDtProps.DateTimeInfo.DisableEvents)
	d.Set("disable_fallback", hostDtProps.DateTimeInfo.DisableFallback)
	return nil
}
