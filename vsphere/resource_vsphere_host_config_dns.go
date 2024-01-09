// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"hostname"},
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"dns_hostname": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"dns_servers": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceVSphereHostConfigDNSCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client).vimClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	// TODO: maybe we can move this to a ValidateFunc on the argument list above so "tf plan" catches it and you don't have to wait for a "TF apply" to see the error
	if strings.Contains(d.Get("dns_hostname").(string), ".") {
		return fmt.Errorf("create func - Invalid hostname supplied. Should not be FQDN")
	}

	host, hostID, err := hostsystem.FromHostnameOrID(c, d)
	if err != nil {
		return fmt.Errorf("create func - error getting host ID FromHostnameOrID")
	}

	hns, err := hostNetworkSystemFromHostSystemID(c, host.Reference().Value)
	if err != nil {
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
		Dhcp:         false,
		HostName:     d.Get("dns_hostname").(string),
		DomainName:   d.Get("domain_name").(string),
		Address:      holder_dns_servers,
		SearchDomain: holder_search_domains,
	}

	err = hns.UpdateDnsConfig(ctx, host_dns_config)
	if err != nil {
		return fmt.Errorf("create func - error updating dns config: %s", err)
	}

	// add the resource into the terraform state
	d.SetId(hostID)

	return resourceVSphereHostConfigDNSRead(d, meta)
}

func resourceVSphereHostConfigDNSRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client).vimClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	host, _, err := hostsystem.FromHostnameOrID(c, d)
	if err != nil {
		return fmt.Errorf("read func - error getting host ID FromHostnameOrID")
	}

	hns, err := hostNetworkSystemFromHostSystemID(c, host.Reference().Value)
	if err != nil {
		return fmt.Errorf("read func - error getting host network system: %s - %s - %s", err, host.Name(), host.Reference().Value)
	}

	var hostNetworkProps mo.HostNetworkSystem
	err = hns.Properties(ctx, hns.Reference(), nil, &hostNetworkProps)
	if err != nil {
		fmt.Printf("read func - had an error getting the network system properties: %s", err)
	}

	dns_config := hostNetworkProps.DnsConfig.GetHostDnsConfig()
	d.Set("dns_hostname", dns_config.HostName)
	d.Set("dns_servers", dns_config.Address)
	d.Set("search_domains", dns_config.SearchDomain)
	d.Set("domain_name", dns_config.DomainName)

	return nil
}

func resourceVSphereHostConfigDNSUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChanges("dns_hostname", "dns_servers", "search_domains", "domain_name") {
		c := meta.(*Client).vimClient
		ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer cancel()

		host, _, err := hostsystem.FromHostnameOrID(c, d)
		if err != nil {
			return fmt.Errorf("update func - error getting host ID FromHostnameOrID")
		}

		hns, err := hostNetworkSystemFromHostSystemID(c, host.Reference().Value)
		if err != nil {
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
			Dhcp:         false,
			HostName:     d.Get("dns_hostname").(string),
			DomainName:   d.Get("domain_name").(string),
			Address:      holder_dns_servers,
			SearchDomain: holder_search_domains,
		}

		err = hns.UpdateDnsConfig(ctx, host_dns_config)
		if err != nil {
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

	// this should get removed for new import helper func
	host, _, err := hostsystem.FromHostnameOrID(c, d)
	if err != nil {
		return nil, fmt.Errorf("import func - error getting host ID FromHostnameOrID")
	}

	hns, err := hostNetworkSystemFromHostSystemID(c, host.Reference().Value)
	if err != nil {
		return nil, fmt.Errorf("import func - error getting host network system: %s", err)
	}

	var hostNetworkProps mo.HostNetworkSystem
	err = hns.Properties(ctx, hns.Reference(), nil, &hostNetworkProps)
	if err != nil {
		return nil, fmt.Errorf("import func - had an error getting the network system properties: %s", err)
	}

	dns_config := hostNetworkProps.DnsConfig.GetHostDnsConfig()
	// update this to the new ID value from our new import helper function
	d.SetId(d.Id())
	// put some logic here to set either "host_system_id" or "hostname" and swap d.Id() with new ID from import helper func
	d.Set("host_system_id", d.Id())
	d.Set("hostname", dns_config.HostName)
	d.Set("dns_servers", dns_config.Address)
	d.Set("search_domains", dns_config.SearchDomain)
	d.Set("domain_name", dns_config.DomainName)

	return []*schema.ResourceData{d}, nil
}
