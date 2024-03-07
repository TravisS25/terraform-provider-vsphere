// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
)

const (
	vsphereVcenterDnsID = "tf-vcenter-dns"

	dnsServersPath = "/appliance/networking/dns/servers"
)

func resourceVSphereVcenterDNS() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereVcenterDNSCreate,
		Read:   resourceVSphereVcenterDNSRead,
		Update: resourceVSphereVcenterDNSUpdate,
		Delete: resourceVSphereVcenterDNSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVSphereVcenterDNSImport,
		},

		Schema: map[string]*schema.Schema{
			"soft_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If set, will skip actually deleting resource and will simply be removed from state",
			},
			"servers": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of the DNS servers to use",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceVSphereVcenterDNSCreate(d *schema.ResourceData, meta interface{}) error {
	err := vsphereVcenterDNSUpdate(d, meta)
	if err != nil {
		return err
	}

	d.SetId(vsphereVcenterDnsID)
	return nil
}

func resourceVSphereVcenterDNSRead(d *schema.ResourceData, meta interface{}) error {
	return vsphereVcenterDNSRead(d, meta)
}

func resourceVSphereVcenterDNSUpdate(d *schema.ResourceData, meta interface{}) error {
	return vsphereVcenterDNSUpdate(d, meta)
}

func resourceVSphereVcenterDNSDelete(d *schema.ResourceData, meta interface{}) error {
	if !d.Get("soft_delete").(bool) {
		var err error

		client := meta.(*Client).restClient
		if _, err = viapi.RestRequest[[]interface{}](
			client,
			http.MethodPut,
			dnsServersPath,
			map[string]interface{}{
				"config": map[string]interface{}{
					"mode":    "is_static",
					"servers": []interface{}{},
				},
			},
		); err != nil {
			return fmt.Errorf("error deleting dns server config: %s", err)
		}
	}

	return nil
}

func resourceVSphereVcenterDNSImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := vsphereVcenterDNSRead(d, meta)
	if err != nil {
		return nil, err
	}

	d.SetId(vsphereVcenterDnsID)
	d.Set("soft_delete", true)
	return []*schema.ResourceData{d}, nil
}

func vsphereVcenterDNSRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).restClient

	bodyRes, err := viapi.RestRequest[map[string]interface{}](client, http.MethodGet, dnsServersPath, nil)
	if err != nil {
		return fmt.Errorf("error retrieving dns servers response: %s", err)
	}

	d.Set("servers", bodyRes["servers"])
	return nil
}

func vsphereVcenterDNSUpdate(d *schema.ResourceData, meta interface{}) error {
	var err error

	client := meta.(*Client).restClient

	// Making request twice here as the first payload is the way to do on older vmware versions
	// and the second payload is how to do on new versions so if first way errors out, try
	// second way before erroring out.  This is a quick fix and if there is a better way
	// this should be updated in the future
	if _, err = viapi.RestRequest[[]interface{}](
		client,
		http.MethodPut,
		dnsServersPath,
		map[string]interface{}{
			"config": map[string]interface{}{
				"mode":    "is_static",
				"servers": d.Get("servers").(*schema.Set).List(),
			},
		},
	); err != nil {
		if _, err = viapi.RestRequest[[]interface{}](
			client,
			http.MethodPut,
			dnsServersPath,
			map[string]interface{}{
				"mode":    "is_static",
				"servers": d.Get("servers").(*schema.Set).List(),
			},
		); err != nil {
			return fmt.Errorf("error making update request for dns server config: %s", err)
		}
	}

	return nil
}
