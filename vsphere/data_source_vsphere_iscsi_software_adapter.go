// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/iscsi"
)

func dataSourceVSphereIscsiSoftwareAdapter() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereIscsiSoftwareAdapterRead,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Description: "The host to gather iscsi information",
				Optional:    true,
			},
		},
	}
}

func dataSourceVSphereIscsiSoftwareAdapterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
	if err != nil {
		return err
	}

	if hssProps.StorageDeviceInfo.SoftwareInternetScsiEnabled {
		if _, err = iscsi.GetIscsiAdater(hssProps, hostID); err != nil {
			return err
		}
	}

	d.SetId(hostID)
	return nil
}
