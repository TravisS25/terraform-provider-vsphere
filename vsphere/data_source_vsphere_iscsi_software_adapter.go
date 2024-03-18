// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
)

func dataSourceVSphereIscsiSoftwareAdapter() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereIscsiSoftwareAdapterRead,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "Host to enable iscsi software adapter",
				ExactlyOneOf: []string{"hostname"},
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Hostname of host system to enable software adapter",
			},
			"iscsi_name": {
				Type:        schema.TypeString,
				Description: "The name of the iscsi software adapter for host",
				Computed:    true,
			},
			"adapter_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Iscsi adapter name that is created when enabling software adapter.  This will be in the form of 'vmhb<unique_name>'",
			},
		},
	}
}

func dataSourceVSphereIscsiSoftwareAdapterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, hr, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host for iscsi data source: %s", err)
	}

	if err = iscsiSoftwareAdapterRead(client, d, host, true); err != nil {
		return err
	}

	d.SetId(hr.Value)
	return nil
}
