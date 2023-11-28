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
			"ntp_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Config settings for ntp",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "List of ntp servers to use",
							//ValidateFunc: validation.IsIPAddress,
							Elem: &schema.Schema{Type: schema.TypeString},
						},
						"config_file": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Content of ntp.conf host configuration file, split by lines for ntpd version 4.2.8",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
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
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(types.HostDateTimeInfoProtocolNtp),
						string(types.HostDateTimeInfoProtocolPtp),
					},
					true,
				),
			},
			"disable_fallback": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disables falling back to ntp if ptp fails when set",
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(types.HostDateTimeInfoProtocolNtp),
						string(types.HostDateTimeInfoProtocolPtp),
					},
					true,
				),
			},
		},
	}
}

func resourceVSphereHostConfigDateTimeRead(d *schema.ResourceData, meta interface{}) error {
	hostID := d.Get("host_system_id").(string)
	log.Printf("[INFO] reading datetime configuration for host '%s'", hostID)
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
		return fmt.Errorf("error trying to get datetime system object from host '%s': %s", hostID, err)
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

	ntpCfgList := d.Get("ntp_config").([]interface{})

	if len(ntpCfgList) > 0 {
		ntpCfg := ntpCfgList[0].(map[string]interface{})
		serverList := ntpCfg["server"].(*schema.Set).List()
		servers := make([]string, 0, len(serverList))

		for _, v := range serverList {
			servers = append(servers, v.(string))
		}

		cfgFileList := ntpCfg["config_file"].(*schema.Set).List()
		cfgFiles := make([]string, 0, len(cfgFileList))

		for _, v := range cfgFileList {
			cfgFiles = append(cfgFiles, v.(string))
		}

		cfg.NtpConfig = &types.HostNtpConfig{
			Server:     servers,
			ConfigFile: cfgFiles,
		}
	}

	log.Printf("[INFO] creating datetime configuration for host '%s'", hostID)

	if err = hostDt.UpdateConfig(context.Background(), cfg); err != nil {
		return fmt.Errorf("error trying to create datetime configuration for host '%s': %s", hostID, err)
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
		return fmt.Errorf("error trying to get datetime system object from host '%s': %s", hostID, err)
	}

	var hostDtProps mo.HostDateTimeSystem
	if err = hostDt.Properties(context.Background(), hostDt.Reference(), nil, &hostDtProps); err != nil {
		return fmt.Errorf("error trying to gather datetime properties from host '%s': %s", hostID, err)
	}

	disableEvents := d.Get("disable_events").(bool)
	disableFallback := d.Get("disable_fallback").(bool)
	cfg := types.HostDateTimeConfig{
		Protocol:        d.Get("protocol").(string),
		DisableEvents:   &disableEvents,
		DisableFallback: &disableFallback,
	}

	// TODO: Come back to this later
	if d.HasChange("ntp_config") {
		_, newValue := d.GetChange("ntp_config")
		newList := newValue.([]interface{})

		if len(newList) > 0 {
			ntpCfg := newList[0].(map[string]interface{})
			serverList := ntpCfg["server"].(*schema.Set).List()
			servers := make([]string, 0, len(serverList))

			for _, v := range serverList {
				servers = append(servers, v.(string))
			}

			cfgFileList := ntpCfg["config_file"].(*schema.Set).List()
			cfgFiles := make([]string, 0, len(cfgFileList))

			for _, v := range cfgFileList {
				cfgFiles = append(cfgFiles, v.(string))
			}

			cfg.NtpConfig = &types.HostNtpConfig{
				Server:     servers,
				ConfigFile: cfgFiles,
			}
		}
	}

	log.Printf("[INFO] updating datetime configuration for host '%s'", hostID)

	if err = hostDt.UpdateConfig(context.Background(), cfg); err != nil {
		return fmt.Errorf("error trying to update datetime configuration for host '%s': %s", hostID, err)
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
		return fmt.Errorf("error trying to get datetime system object from host '%s': %s", hostID, err)
	}

	log.Printf("[INFO] deleting datetime configuration for host '%s'", hostID)

	factoryDefaults := true

	if err = hostDt.UpdateConfig(context.Background(), types.HostDateTimeConfig{
		Protocol:               string(types.HostDateTimeInfoProtocolNtp),
		ResetToFactoryDefaults: &factoryDefaults,
	}); err != nil {
		return fmt.Errorf("error trying to delete datetime configuration for host '%s': %s", hostID, err)
	}

	return nil
}

func resourceVSphereHostConfigDateTimeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	hostID := d.Id()
	log.Printf("[INFO] importing datetime configuration for host '%s'", hostID)
	err := hostConfigDateTimeRead(d, meta, hostID)
	if err != nil {
		return nil, err
	}

	d.SetId(hostID)
	return []*schema.ResourceData{d}, nil
}

func resourceVSphereHostConfigDateTimeCustomDiff(ctx context.Context, rd *schema.ResourceDiff, meta interface{}) error {
	ntpCfg := rd.Get("ntp_config").([]interface{})

	if len(ntpCfg) == 0 {
		return fmt.Errorf("'ntp_config' is required")
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
		return fmt.Errorf("error trying to get datetime system object from host '%s': %s", hostID, err)
	}

	var hostDtProps mo.HostDateTimeSystem
	if err = hostDt.Properties(context.Background(), hostDt.Reference(), nil, &hostDtProps); err != nil {
		return fmt.Errorf("error trying to gather datetime properties from host '%s': %s", hostID, err)
	}

	ntpCfg := []interface{}{
		map[string]interface{}{
			"server":      hostDtProps.DateTimeInfo.NtpConfig.Server,
			"config_file": hostDtProps.DateTimeInfo.NtpConfig.ConfigFile,
		},
	}

	d.Set("ntp_config", ntpCfg)
	d.Set("protocol", hostDtProps.DateTimeInfo.SystemClockProtocol)
	d.Set("disable_events", hostDtProps.DateTimeInfo.DisableEvents)
	d.Set("disable_fallback", hostDtProps.DateTimeInfo.DisableFallback)

	return nil
}
