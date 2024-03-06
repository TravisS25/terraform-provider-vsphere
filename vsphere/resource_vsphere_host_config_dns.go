// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func resourceVSphereHostConfigDNS() *schema.Resource {
	return &schema.Resource{
		Create:        resourceVSphereHostConfigDNSCreate,
		Read:          resourceVSphereHostConfigDNSRead,
		Update:        resourceVSphereHostConfigDNSUpdate,
		Delete:        resourceVSphereHostConfigDNSDelete,
		CustomizeDiff: resourceVSphereHostConfigDNSCustomDiff,
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
			"soft_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"dns_hostname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"dns_servers": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"search_domains": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceVSphereHostConfigDNSCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, hr, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host on create: %s", err)
	}

	if err = hostConfigDNSUpdate(client, d, host); err != nil {
		return fmt.Errorf("error creating dns settings on host '%s': %s", host.Name(), err)
	}

	d.SetId(hr.Value)
	return nil
}

func resourceVSphereHostConfigDNSRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, _, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host on read: %s", err)
	}
	return hostConfigDNSRead(client, d, host)
}

func resourceVSphereHostConfigDNSUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, _, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host on create: %s", err)
	}

	if err = hostConfigDNSUpdate(client, d, host); err != nil {
		return fmt.Errorf("error updating dns settings on host '%s': %s", host.Name(), err)
	}

	return nil
}

func resourceVSphereHostConfigDNSDelete(d *schema.ResourceData, meta interface{}) error {
	if !d.Get("soft_delete").(bool) {
		ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer cancel()

		client := meta.(*Client).vimClient
		host, _, err := hostsystem.FromHostnameOrID(client, d)
		if err != nil {
			return fmt.Errorf("error retrieving host on delete: %s", err)
		}

		hns, err := hostNetworkSystemFromHostSystemID(client, host.Reference().Value)
		if err != nil {
			return fmt.Errorf("error retrieving host network system on host '%s': %s", host.Name(), err)
		}

		if err = hns.UpdateDnsConfig(
			ctx,
			&types.HostDnsConfig{
				Dhcp:         false,
				HostName:     d.Get("dns_hostname").(string),
				DomainName:   "",
				Address:      []string{},
				SearchDomain: []string{},
			},
		); err != nil {
			return fmt.Errorf("error updating dns config: %s", err)
		}
	}

	return nil
}

func resourceVSphereHostConfigDNSCustomDiff(ctx context.Context, rd *schema.ResourceDiff, meta interface{}) error {
	if strings.Contains(rd.Get("dns_hostname").(string), ".") {
		return fmt.Errorf("'dns_hostname' should simply be the hostname itself, NOT a FQDN")
	}

	return nil
}

func resourceVSphereHostConfigDNSImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Client).vimClient
	_, hr, err := hostsystem.CheckIfHostnameOrID(client, d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving host on import: %s", err)
	}

	d.SetId(hr.Value)
	d.Set(hr.IDName, hr.Value)
	d.Set("soft_delete", true)
	return []*schema.ResourceData{d}, nil
}

func hostConfigDNSRead(client *govmomi.Client, d *schema.ResourceData, host *object.HostSystem) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	hns, err := hostNetworkSystemFromHostSystemID(client, host.Reference().Value)
	if err != nil {
		return fmt.Errorf("error retrieving host network system from host '%s': %s", host.Name(), err)
	}

	var hostNetworkProps mo.HostNetworkSystem
	if err = hns.Properties(ctx, hns.Reference(), nil, &hostNetworkProps); err != nil {
		fmt.Printf("error retrieving network system properties from host '%s': %s", host.Name(), err)
	}

	dnsCfg := hostNetworkProps.DnsConfig.GetHostDnsConfig()
	d.Set("dns_hostname", dnsCfg.HostName)
	d.Set("dns_servers", dnsCfg.Address)
	d.Set("search_domains", dnsCfg.SearchDomain)
	d.Set("domain_name", dnsCfg.DomainName)
	return nil
}

func hostConfigDNSUpdate(client *govmomi.Client, d *schema.ResourceData, host *object.HostSystem) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	hns, err := hostNetworkSystemFromHostSystemID(client, host.Reference().Value)
	if err != nil {
		return fmt.Errorf("error retrieving host network system on host '%s': %s", host.Name(), err)
	}

	dnsServers := []string{}
	for _, v := range d.Get("dns_servers").(*schema.Set).List() {
		dnsServers = append(dnsServers, v.(string))
	}

	searchDomains := []string{}
	for _, v := range d.Get("search_domains").(*schema.Set).List() {
		searchDomains = append(searchDomains, v.(string))
	}

	if err = hns.UpdateDnsConfig(
		ctx,
		&types.HostDnsConfig{
			Dhcp:         false,
			HostName:     d.Get("dns_hostname").(string),
			DomainName:   d.Get("domain_name").(string),
			Address:      dnsServers,
			SearchDomain: searchDomains,
		},
	); err != nil {
		return fmt.Errorf("error updating dns config: %s", err)
	}

	return nil
}
