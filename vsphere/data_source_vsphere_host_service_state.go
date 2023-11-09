// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostservicestate"
)

func dataSourceVSphereHostServiceState() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereHostServiceStateRead,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Host id of machine",
			},
			"service_keys": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Key for service on given host.  Options: " + hostservicestate.GetServiceKeyMsg(serviceKeyList),
				Elem:        &schema.Schema{Type: schema.TypeString},
				// ValidateFunc: validation.StringInSlice(serviceKeyList, false),
			},
			"running": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "State of the service",
			},
			"policy": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy of the service",
			},
		},
	}
}

func dataSourceVSphereHostServiceStateRead(d *schema.ResourceData, meta interface{}) error {
	// err := iscsiSoftwareAdapterRead(d, meta, true)
	// if err != nil {
	// 	return err
	// }

	// d.SetId(d.Get("host_system_id").(string))
	return nil
}
