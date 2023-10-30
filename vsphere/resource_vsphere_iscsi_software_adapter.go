// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"strings"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/customattribute"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	iscsiAdapterName = "internetscsihba"
)

func resourceVSphereIscsiSoftwareAdapter() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereDatacenterCreate,
		Read:   resourceVSphereDatacenterRead,
		Update: resourceVSphereDatacenterUpdate,
		Delete: resourceVSphereDatacenterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVSphereDatacenterImport,
		},

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Host to enable iscsi software adapter",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Default:     true,
				Description: "Determines whether to enable iscsi software adpater.  Default: true",
			},
			"iscsi_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The unique iqn name for the iscsi software adapter if enabled.  If left blank, vmware will generate the iqn name",
			},

			// Add tags schema
			vSphereTagAttributeKey: tagsSchema(),

			// Custom Attributes
			customattribute.ConfigKey: customattribute.ConfigSchema(),
		},
	}
}

func resourceVSphereIscsiSoftwareAdapterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		if viapi.IsManagedObjectNotFoundError(err) {
			return fmt.Errorf("could not find host with id %s", hostID)
		}

		return fmt.Errorf("error while searching host %s: %s ", hostID, err)
	}

	hsProps, err := hostsystem.Properties(hs)
	if err != nil {
		return fmt.Errorf("error trying to retrieve host system properties: %s", err)
	}

	hss := object.NewHostStorageSystem(client.Client, *hsProps.ConfigManager.StorageSystem)
	hssProps, err := hostsystem.HostStorageSystemProperties(hss)
	if err != nil {
		return fmt.Errorf("error trying to retrieve host storage system properties: %s", err)
	}

	enabled := d.Get("enabled").(bool)

	if _, err = methods.UpdateSoftwareInternetScsiEnabled(
		context.Background(),
		client.Client,
		&types.UpdateSoftwareInternetScsiEnabled{
			This:    hssProps.Reference(),
			Enabled: enabled,
		},
	); err != nil {
		return fmt.Errorf("error while trying to enable/disable iscsi software adapter: %s", err)
	}

	if enabled {
		if name, ok := d.GetOk("iscsi_name"); ok {
			for _, v := range hssProps.StorageDeviceInfo.HostBusAdapter {
				if strings.Contains(strings.ToLower(v.GetHostHostBusAdapter().Key), iscsiAdapterName) {
					fmt.Printf("found the adapter name!\n")

					if _, err = methods.UpdateInternetScsiName(context.Background(), client.Client, &types.UpdateInternetScsiName{
						This:           hss.Reference(),
						IScsiHbaDevice: v.GetHostHostBusAdapter().Device,
						IScsiName:      name.(string),
					}); err != nil {
						return fmt.Errorf("could not update iscsi name: %s", err)
					}
				}
			}
		}
	}

	return resourceVSphereDatacenterRead(d, meta)
}

func resourceVSphereIscsiSoftwareAdapterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		if viapi.IsManagedObjectNotFoundError(err) {
			return fmt.Errorf("could not find host with id %s", hostID)
		}

		return fmt.Errorf("error while searching host %s. Error: %s ", hostID, err)
	}

	d.Set("host_system_id", hostID)

	hsProps, err := hostsystem.Properties(hs)
	if err != nil {
		return fmt.Errorf("error trying to retrieve host system properties: %s", err)
	}

	hss := object.NewHostStorageSystem(client.Client, *hsProps.ConfigManager.StorageSystem)
	hssProps, err := hostsystem.HostStorageSystemProperties(hss)
	if err != nil {
		return fmt.Errorf("error trying to retrieve host storage system properties: %s", err)
	}

	d.Set("enabled", hssProps.StorageDeviceInfo.SoftwareInternetScsiEnabled)

	if hssProps.StorageDeviceInfo.SoftwareInternetScsiEnabled {
		for _, v := range hssProps.StorageDeviceInfo.HostBusAdapter {
			if strings.Contains(strings.ToLower(v.GetHostHostBusAdapter().Key), iscsiAdapterName) {
				d.Set("iscsi_name", v.(*types.HostInternetScsiHba).IScsiName)
			}
		}
	}

	return nil
}

func resourceVSphereIscsiSoftwareAdapterUpdate(d *schema.ResourceData, meta interface{}) error {
	var err error
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		if viapi.IsManagedObjectNotFoundError(err) {
			return fmt.Errorf("could not find host with id %s", hostID)
		}

		return fmt.Errorf("error while searching host %s. Error: %s ", hostID, err)
	}

	hsProps, err := hostsystem.Properties(hs)
	if err != nil {
		return fmt.Errorf("error trying to retrieve host system properties: %s", err)
	}

	hss := object.NewHostStorageSystem(client.Client, *hsProps.ConfigManager.StorageSystem)
	hssProps, err := hostsystem.HostStorageSystemProperties(hss)
	if err != nil {
		return fmt.Errorf("error trying to retrieve host storage system properties: %s", err)
	}

	if d.HasChange("enabled") {
		_, enabledVal := d.GetChange("enabled")

		if _, err = methods.UpdateSoftwareInternetScsiEnabled(
			context.Background(),
			client.Client,
			&types.UpdateSoftwareInternetScsiEnabled{
				This:    hssProps.Reference(),
				Enabled: enabledVal.(bool),
			},
		); err != nil {
			return fmt.Errorf("error while trying to enable/disable iscsi software adapter: %s", err)
		}

		if enabledVal.(bool) && d.HasChange("iscsi_name") {
			_, iscsiName := d.GetChange("iscsi_name")

			for _, v := range hssProps.StorageDeviceInfo.HostBusAdapter {
				if strings.Contains(strings.ToLower(v.GetHostHostBusAdapter().Key), iscsiAdapterName) {
					fmt.Printf("found the adapter name!\n")

					if _, err = methods.UpdateInternetScsiName(context.Background(), client.Client, &types.UpdateInternetScsiName{
						This:           hss.Reference(),
						IScsiHbaDevice: v.GetHostHostBusAdapter().Device,
						IScsiName:      iscsiName.(string),
					}); err != nil {
						return fmt.Errorf("could not update iscsi name: %s", err)
					}
				}
			}
		}
	}

	return nil
}

func resourceVSphereIscsiSoftwareAdapterDelete(d *schema.ResourceData, meta interface{}) error {
	var err error
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		if viapi.IsManagedObjectNotFoundError(err) {
			return fmt.Errorf("could not find host with id %s", hostID)
		}

		return fmt.Errorf("error while searching host %s. Error: %s ", hostID, err)
	}

	hsProps, err := hostsystem.Properties(hs)
	if err != nil {
		return fmt.Errorf("error trying to retrieve host system properties: %s", err)
	}

	hss := object.NewHostStorageSystem(client.Client, *hsProps.ConfigManager.StorageSystem)
	hssProps, err := hostsystem.HostStorageSystemProperties(hss)
	if err != nil {
		return fmt.Errorf("error trying to retrieve host storage system properties: %s", err)
	}

	if _, err = methods.UpdateSoftwareInternetScsiEnabled(
		context.Background(),
		client.Client,
		&types.UpdateSoftwareInternetScsiEnabled{
			This:    hssProps.Reference(),
			Enabled: false,
		},
	); err != nil {
		return fmt.Errorf("error while trying to delete iscsi software adapter: %s", err)
	}

	return nil
}

func resourceVSphereIscsiSoftwareAdapterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return nil, nil
}
