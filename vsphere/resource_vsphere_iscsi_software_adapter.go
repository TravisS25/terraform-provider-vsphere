// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/iscsi"
)

func resourceVSphereIscsiSoftwareAdapter() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereIscsiSoftwareAdapterCreate,
		Read:   resourceVSphereIscsiSoftwareAdapterRead,
		Update: resourceVSphereIscsiSoftwareAdapterUpdate,
		Delete: resourceVSphereIscsiSoftwareAdapterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVSphereIscsiSoftwareAdapterImport,
		},

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Host to enable iscsi software adapter",
			},
			"iscsi_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The unique iqn name for the iscsi software adapter if enabled.  If left blank, vmware will generate the iqn name",
			},
			"adapter_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Iscsi adapter name that is created when enabling software adapter.  This will be in the form of 'vmhb<unique_name>'",
			},
		},
	}
}

func resourceVSphereIscsiSoftwareAdapterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	hss, err := hostsystem.GetHostStorageSystemFromHost(client, hostID)
	if err != nil {
		return err
	}

	if err = iscsi.UpdateSoftwareInternetScsi(client, hss.Reference(), hostID, true); err != nil {
		return err
	}

	if err = hss.RescanAllHba(context.Background()); err != nil {
		return fmt.Errorf(
			"error trying to rescan storage adapters after enabling iscsi software adapter for host '%s': %s",
			hostID,
			err,
		)
	}

	hssProps, err := hostsystem.HostStorageSystemProperties(hss)
	if err != nil {
		return err
	}

	adapter, err := iscsi.GetIscsiSoftwareAdater(hssProps, hostID)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s:%s", hostID, adapter.Device))
	d.Set("adapter_id", adapter.Device)

	if name, ok := d.GetOk("iscsi_name"); ok {
		if err = iscsi.UpdateIscsiName(hostID, adapter.Device, name.(string), client, hssProps.Reference()); err != nil {
			return err
		}

		d.Set("iscsi_name", name.(string))
	} else {
		d.Set("iscsi_name", adapter.IScsiName)
	}

	return resourceVSphereIscsiSoftwareAdapterRead(d, meta)
}

func resourceVSphereIscsiSoftwareAdapterRead(d *schema.ResourceData, meta interface{}) error {
	return iscsiSoftwareAdapterRead(d, meta, false)
}

func resourceVSphereIscsiSoftwareAdapterUpdate(d *schema.ResourceData, meta interface{}) error {
	var err error
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
	if err != nil {
		return err
	}

	if d.HasChange("iscsi_name") {
		_, iscsiName := d.GetChange("iscsi_name")
		adapter, err := iscsi.GetIscsiSoftwareAdater(hssProps, hostID)
		if err != nil {
			return err
		}

		if err = iscsi.UpdateIscsiName(hostID, adapter.Device, iscsiName.(string), client, hssProps.Reference()); err != nil {
			return err
		}
	}

	return nil
}

func resourceVSphereIscsiSoftwareAdapterDelete(d *schema.ResourceData, meta interface{}) error {
	var err error
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
	if err != nil {
		return err
	}

	return iscsi.UpdateSoftwareInternetScsi(client, hssProps.Reference(), hostID, false)
}

func resourceVSphereIscsiSoftwareAdapterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Client).vimClient
	hostID := d.Id()
	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
	if err != nil {
		return nil, err
	}

	adapter, err := iscsi.GetIscsiSoftwareAdater(hssProps, hostID)
	if err != nil {
		return nil, err
	}

	d.SetId(fmt.Sprintf("%s:%s", hostID, adapter.Device))
	d.Set("host_system_id", hostID)
	return []*schema.ResourceData{d}, nil
}

func iscsiSoftwareAdapterRead(d *schema.ResourceData, meta interface{}, isDataSource bool) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
	if err != nil {
		return err
	}

	if hssProps.StorageDeviceInfo.SoftwareInternetScsiEnabled {
		adapter, err := iscsi.GetIscsiSoftwareAdater(hssProps, hostID)
		if err != nil {
			return err
		}

		d.Set("iscsi_name", adapter.IScsiName)
	} else if isDataSource {
		return fmt.Errorf("iscsi software adapter is not enabled for host '%s'", hostID)
	}

	return nil
}
