// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"log"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostconfig"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "Host id of machine that will update syslog",
				ExactlyOneOf: []string{"hostname"},
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Hostname of machine that will update syslog",
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
	client := meta.(*Client).vimClient
	host, _, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] reading syslog settings for host '%s'", host.Name())

	if err = hostconfig.HostConfigSyslogRead(d, client, host); err != nil {
		return err
	}

	return nil
}

func resourceVSphereHostConfigSyslogCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, _, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] creating syslog settings for host '%s'", host.Name())

	if err = hostconfig.UpdateHostConfigSyslog(d, client, host, false); err != nil {
		return err
	}

	d.SetId(host.Name())

	return nil
}

func resourceVSphereHostConfigSyslogUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, _, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] updating syslog settings for host '%s'", host.Name())

	if err = hostconfig.UpdateHostConfigSyslog(d, client, host, false); err != nil {
		return err
	}

	return nil
}

func resourceVSphereHostConfigSyslogDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, _, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return err
	}

	log.Printf("[INFO] deleting syslog settings for host '%s'", host.Name())

	if err = hostconfig.UpdateHostConfigSyslog(d, client, host, true); err != nil {
		return err
	}

	return nil
}

func resourceVSphereHostConfigSyslogImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Client).vimClient
	host, hostReturn, err := hostsystem.CheckIfHostnameOrID(client, d.Id())
	if err != nil {
		return nil, err
	}

	if err = hostconfig.HostConfigSyslogRead(d, meta.(*Client).vimClient, host); err != nil {
		return nil, err
	}

	d.SetId(d.Id())
	d.Set(hostReturn.IDName, hostReturn.Value)
	return []*schema.ResourceData{d}, nil
}
