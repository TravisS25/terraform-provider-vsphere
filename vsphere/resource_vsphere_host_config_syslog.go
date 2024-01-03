// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"log"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostconfig"
)

func resourceVSphereHostConfigSyslog() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereHostConfigSyslogCreate,
		Read:   resourceVSphereHostConfigSyslogRead,
		Update: resourceVSphereHostConfigSyslogUpdate,
		Delete: resourceVSphereHostConfigSyslogDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVSphereHostConfigSyslogImport,
		},

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Host id of machine that will update syslog",
			},
			"log_host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The remote host to output logs to",
			},
			"log_level": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "info",
				Description: "The log level to output logs",
				ValidateFunc: validation.StringInSlice(
					[]string{"info", "debug", "warning", "error"},
					false,
				),
			},
		},
	}
}

func resourceVSphereHostConfigSyslogRead(d *schema.ResourceData, meta interface{}) error {
	hostID := d.Get("host_system_id").(string)

	log.Printf("[INFO] reading syslog settings for host '%s'", hostID)

	err := hostconfig.HostConfigSyslogRead(context.Background(), d, meta.(*Client).vimClient, hostID)
	if err != nil {
		return err
	}

	d.Set("host_system_id", hostID)

	return nil
}

func resourceVSphereHostConfigSyslogCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	log.Printf("[INFO] creating syslog settings for host '%s'", hostID)

	err := hostconfig.UpdateHostConfigSyslog(context.Background(), d, client, hostID, false)
	if err != nil {
		return err
	}

	d.SetId(hostID)

	return nil
}

func resourceVSphereHostConfigSyslogUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	log.Printf("[INFO] updating syslog settings for host '%s'", hostID)

	err := hostconfig.UpdateHostConfigSyslog(context.Background(), d, client, hostID, false)
	if err != nil {
		return err
	}

	return nil
}

func resourceVSphereHostConfigSyslogDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	log.Printf("[INFO] deleting syslog settings for host '%s'", hostID)

	err := hostconfig.UpdateHostConfigSyslog(context.Background(), d, client, hostID, true)
	if err != nil {
		return err
	}

	return nil
}

func resourceVSphereHostConfigSyslogImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	hostID := d.Id()
	err := hostconfig.HostConfigSyslogRead(ctx, d, meta.(*Client).vimClient, hostID)
	if err != nil {
		return nil, err
	}

	d.SetId(hostID)
	d.Set("host_system_id", hostID)

	return []*schema.ResourceData{d}, nil
}
