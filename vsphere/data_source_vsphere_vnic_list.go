// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
)

func dataSourceVSphereVnicList() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereVnicListRead,
		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "Host id of machine to grab vnic information",
				ExactlyOneOf: []string{"hostname"},
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Hostname of machine to grab vnic information",
			},
			"vnics": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of vnics of given hostname/host_system_id",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Device name of vnic",
						},
						"port": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Port of vnic",
						},
						"port_group": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Port group of vnic",
						},
						"spec": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Spec for current vnic",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mac": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Mac of vnic",
									},
									"mtu": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "MTU of vnic",
									},
									"ip": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "Port group of vnic",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"dhcp": {
													Type:        schema.TypeBool,
													Computed:    true,
													Description: "Determines if vnic gets address based on dhcp",
												},
												"ip_address": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "IP of vnic",
												},
												"subnet_mask": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Subnet mask of ip for vnic",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereVnicListRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, hr, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host on vnic list read: %s", err)
	}

	hostProps, err := hostsystem.Properties(host)
	if err != nil {
		return fmt.Errorf("error retrieving host properties for host %q: %s", host.Name(), err)
	}

	vnics := make([]map[string]interface{}, 0, len(hostProps.Config.Network.Vnic))

	for _, vnic := range hostProps.Config.Network.Vnic {
		vnics = append(vnics, map[string]interface{}{
			"device":     vnic.Device,
			"port":       vnic.Port,
			"port_group": vnic.Portgroup,
			"spec": []map[string]interface{}{
				{
					"mac": vnic.Spec.Mac,
					"mtu": vnic.Spec.Mtu,
					"ip": []map[string]interface{}{
						{
							"dhcp":        vnic.Spec.Ip.Dhcp,
							"ip_address":  vnic.Spec.Ip.IpAddress,
							"subnet_mask": vnic.Spec.Ip.SubnetMask,
						},
					},
				},
			},
		})
	}

	d.SetId(hr.Value)
	d.Set("vnics", vnics)
	return nil
}
