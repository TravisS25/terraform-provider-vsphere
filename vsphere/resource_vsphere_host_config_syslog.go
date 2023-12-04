// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"log"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	syslogHostKey = "Syslog.global.logHost"
)

func resourceVsphereHostConfigSyslog() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereHostConfigSyslogCreate,
		Read:   resourceVSphereHostConfigSyslogRead,
		Update: resourceVSphereHostConfigSyslogUpdate,
		Delete: resourceVSphereHostConfigSyslogDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVSphereHostConfigSyslogImport,
		},
		CustomizeDiff: resourceVSphereHostConfigSyslogCustomDiff,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Host id of machine that will update service",
			},
			"log_host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The remote host to output logs to",
			},
		},
	}
}

func resourceVSphereHostConfigSyslogRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	optManager, err := host.ConfigManager().OptionManager(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving option manager for host '%s': %s", hostID, err)
	}

	log.Printf("[INFO] querying options in syslog read for host '%s'", hostID)

	opts, err := optManager.Query(ctx, syslogHostKey)
	if err != nil {
		return fmt.Errorf("error trying to query against option manager for host '%s': %s", hostID, err)
	}

	if len(opts) > 0 {
		d.Set("log_host", opts[0].GetOptionValue().Value)
	}

	d.Set("host_system_id", hostID)

	return nil
}

func resourceVSphereHostConfigSyslogCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	logHost := d.Get("log_host").(string)
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	optManager, err := host.ConfigManager().OptionManager(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving option manager for host '%s': %s", hostID, err)
	}

	log.Printf("[INFO] creating syslog options for host '%s'", hostID)

	if err = optManager.Update(
		ctx,
		[]types.BaseOptionValue{
			&types.OptionValue{
				Key:   syslogHostKey,
				Value: logHost,
			},
		},
	); err != nil {
		return fmt.Errorf("error trying to create syslog options for host '%s'", hostID)
	}

	d.SetId(hostID)

	return nil
}

func resourceVSphereHostConfigSyslogUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	logHost := d.Get("log_host").(string)
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	optManager, err := host.ConfigManager().OptionManager(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving option manager for host '%s': %s", hostID, err)
	}

	log.Printf("[INFO] updating syslog options for host '%s'", hostID)

	if err = optManager.Update(
		ctx,
		[]types.BaseOptionValue{
			&types.OptionValue{
				Key:   syslogHostKey,
				Value: logHost,
			},
		},
	); err != nil {
		return fmt.Errorf("error trying to update syslog options for host '%s'", hostID)
	}

	return nil
}

func resourceVSphereHostConfigSyslogDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	optManager, err := host.ConfigManager().OptionManager(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving option manager for host '%s': %s", hostID, err)
	}

	log.Printf("[INFO] deleting syslog options for host '%s'", hostID)

	if err = optManager.Update(
		ctx,
		[]types.BaseOptionValue{
			&types.OptionValue{
				Key:   syslogHostKey,
				Value: nil,
			},
		},
	); err != nil {
		return fmt.Errorf("error trying to delete syslog options for host '%s'", hostID)
	}

	return nil
}

func resourceVSphereHostConfigSyslogImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Client).vimClient
	hostID := d.Id()
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return nil, err
	}

	optManager, err := host.ConfigManager().OptionManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving option manager for host '%s': %s", hostID, err)
	}

	queryOpts, err := optManager.Query(ctx, syslogHostKey)
	if err != nil {
		return nil, fmt.Errorf("error querying for log host on host '%s': %s", hostID, err)
	}

	if len(queryOpts) > 0 {
		d.Set("log_host", queryOpts[0].GetOptionValue().Value)
	}

	d.SetId(hostID)
	d.Set("host_system_id", hostID)

	return []*schema.ResourceData{d}, nil
}

func resourceVSphereHostConfigSyslogCustomDiff(ctx context.Context, rd *schema.ResourceDiff, meta interface{}) error {
	srvs := rd.Get("service").([]interface{})
	trackerMap := map[string]bool{}

	for _, val := range srvs {
		srv := val.(map[string]interface{})

		if _, ok := trackerMap[srv["key"].(string)]; ok {
			return fmt.Errorf("duplicate values for 'key' attribute in 'service' resource is not allowed")
		}
		trackerMap[srv["key"].(string)] = true
	}

	return nil
}
