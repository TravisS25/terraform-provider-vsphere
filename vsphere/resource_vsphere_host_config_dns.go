// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vim25/mo"
)

func resourceVSphereHostConfigDNS() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereHostConfigDNSCreate,
		Read:   resourceVSphereHostConfigDNSRead,
		Update: resourceVSphereHostConfigDNSUpdate,
		Delete: resourceVSphereHostConfigDNSDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVSphereHostConfigDNSImport,
		},
		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:     schema.TypeString,
				Required: true,
				// We might want to do a force-new on this so if a host leaves vcenter inventory and gets re-added (changes IDs) we want TF to clean up the state from the old host ID and add for the new one
				ForceNew: false,
			},
			"hostname": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"dns_servers": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: false,
				Elem: &schema.Schema{Type: schema.TypeString},
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"search_domains": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: false,
				Elem: &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceVSphereHostConfigDNSCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client).vimClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	hns, err := hostNetworkSystemFromHostSystemID(c, d.Get("host_system_id").(string))
	if err != nil{
		return fmt.Errorf("create func - error getting host network system: %s", err)
	}

	holder_dns_servers := []string{}
	for _, v := range d.Get("dns_servers").(*schema.Set).List() {
		holder_dns_servers = append(holder_dns_servers, v.(string))
	}

	holder_search_domains := []string{}
	for _, v := range d.Get("search_domains").(*schema.Set).List() {
		holder_search_domains = append(holder_search_domains, v.(string))
	}

	host_dns_config := &types.HostDnsConfig{
		Dhcp: false,
		HostName: d.Get("hostname").(string),
		DomainName: d.Get("domain_name").(string),
		Address: holder_dns_servers,
		SearchDomain: holder_search_domains,
	}

	err = hns.UpdateDnsConfig(ctx, host_dns_config)
	if err != nil{
		return fmt.Errorf("create func - error updating dns config: %s", err)
	}

	// add the resource into the terraform state
	d.SetId(d.Get("hostname").(string))

	return resourceVSphereHostConfigDNSRead(d, meta)
}

func resourceVSphereHostConfigDNSRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client).vimClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	hns, err := hostNetworkSystemFromHostSystemID(c, d.Get("host_system_id").(string))
	if err != nil{
		return fmt.Errorf("read func - error getting host network system: %s", err)
	}

	var hostNetworkProps mo.HostNetworkSystem
	err = hns.Properties(ctx, hns.Reference(), nil, &hostNetworkProps)
	if err != nil {
		fmt.Printf("read func - had an error getting the network system properties: %s", err)
	}

	dns_config := hostNetworkProps.DnsConfig.GetHostDnsConfig()
	d.Set("hostname", dns_config.HostName)
	d.Set("dns_servers", dns_config.Address)
	d.Set("search_domains", dns_config.SearchDomain)
	d.Set("domain_name", dns_config.DomainName)

	return nil
}

func resourceVSphereHostConfigDNSUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChanges("hostname", "dns_servers", "search_domains", "domain_name") {
		c := meta.(*Client).vimClient
		ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer cancel()

		hns, err := hostNetworkSystemFromHostSystemID(c, d.Get("host_system_id").(string))
		if err != nil{
			return fmt.Errorf("update func - error getting host network system: %s", err)
		}

		holder_dns_servers := []string{}
		for _, v := range d.Get("dns_servers").(*schema.Set).List() {
			holder_dns_servers = append(holder_dns_servers, v.(string))
		}

		holder_search_domains := []string{}
		for _, v := range d.Get("search_domains").(*schema.Set).List() {
			holder_search_domains = append(holder_search_domains, v.(string))
		}

		host_dns_config := &types.HostDnsConfig{
			Dhcp: false,
			HostName: d.Get("hostname").(string),
			DomainName: d.Get("domain_name").(string),
			Address: holder_dns_servers,
			SearchDomain: holder_search_domains,
		}

		err = hns.UpdateDnsConfig(ctx, host_dns_config)
		if err != nil{
			return fmt.Errorf("update func - error updating dns config: %s", err)
		}
	}

	return nil
}

// We do not want to completely remove the DNS settings from the host so if TF performs a delete we just want to leave the host as-is
func resourceVSphereHostConfigDNSDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceVSphereHostConfigDNSImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	c := meta.(*Client).vimClient

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	hns, err := hostNetworkSystemFromHostSystemID(c, d.Id())
	if err != nil{
		return nil, fmt.Errorf("import func - error getting host network system: %s", err)
	}

	var hostNetworkProps mo.HostNetworkSystem
	err = hns.Properties(ctx, hns.Reference(), nil, &hostNetworkProps)
	if err != nil {
		return nil, fmt.Errorf("import func - had an error getting the network system properties: %s", err)
	}

	dns_config := hostNetworkProps.DnsConfig.GetHostDnsConfig()
	d.SetId(d.Id())
	d.Set("host_system_id", d.Id())
	d.Set("hostname", dns_config.HostName)
	d.Set("dns_servers", dns_config.Address)
	d.Set("search_domains", dns_config.SearchDomain)
	d.Set("domain_name", dns_config.DomainName)

	return []*schema.ResourceData{d}, nil
}
