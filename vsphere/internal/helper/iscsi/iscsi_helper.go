// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iscsi

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	ChapResourceKey = "chap"
	PortResourceKey = "port"
	IPResourceKey   = "ip"
)

// GetIscsiSoftwareAdater is util helper that loops through storage adapters and grabs the
// iscsi software adapter
//
// Returns error if iscsi software adapter can not be found (usually due to adapter not being enabled)
func GetIscsiSoftwareAdater(hssProps *mo.HostStorageSystem, host string) (*types.HostInternetScsiHba, error) {
	for _, v := range hssProps.StorageDeviceInfo.HostBusAdapter {
		if strings.Contains(strings.ToLower(v.GetHostHostBusAdapter().Key), "internetscsihba") {
			return v.(*types.HostInternetScsiHba), nil
		}
	}

	return nil, fmt.Errorf("could not find iscsi software adapter for host '%s'", host)
}

func GetIscsiAdater(hssProps *mo.HostStorageSystem, host, adapterID string) (types.BaseHostHostBusAdapter, error) {
	for _, adapter := range hssProps.StorageDeviceInfo.HostBusAdapter {
		if adapter.GetHostHostBusAdapter().Device == adapterID {
			return adapter, nil
		}
	}

	return nil, fmt.Errorf("could not find iscsi adapter device '%s' for host '%s'", adapterID, host)
}

// UpdateIscsiName is util helper that updates iscsi name for adapter
func UpdateIscsiName(host, device, name string, c *govmomi.Client, hssProps types.ManagedObjectReference) error {
	_, err := methods.UpdateInternetScsiName(context.Background(), c, &types.UpdateInternetScsiName{
		This:           hssProps.Reference(),
		IScsiHbaDevice: device,
		IScsiName:      name,
	})

	if err != nil {
		return fmt.Errorf("could not update iscsi name for host '%s': %s", host, err)
	}

	return nil
}

// UpdateSoftwareInternetScsi is util helper that enables/disables the iscsi software adapter
func UpdateSoftwareInternetScsi(client *govmomi.Client, ref types.ManagedObjectReference, host string, enabled bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	_, err := methods.UpdateSoftwareInternetScsiEnabled(
		ctx,
		client.Client,
		&types.UpdateSoftwareInternetScsiEnabled{
			This:    ref,
			Enabled: enabled,
		},
	)

	if err != nil {
		msg := "error while trying to %s iscsi software adapter for host '%s': %s"

		if enabled {
			msg = fmt.Sprintf(msg, "enable", host, err)
		} else {
			msg = fmt.Sprintf(msg, "disable", host, err)
		}
		return fmt.Errorf(msg)
	}

	return nil
}

func AddInternetScsiStaticTargets(
	client *govmomi.Client,
	host,
	adapterID string,
	hssProps *mo.HostStorageSystem,
	targets []types.HostInternetScsiHbaStaticTarget,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	if _, err := methods.AddInternetScsiStaticTargets(ctx, client, &types.AddInternetScsiStaticTargets{
		This:           hssProps.Reference(),
		IScsiHbaDevice: adapterID,
		Targets:        targets,
	}); err != nil {
		return fmt.Errorf("error trying to add static targets for iscsi adapter: %s", err)
	}

	return nil
}

func AddInternetScsiSendTargets(
	client *govmomi.Client,
	host,
	adapterID string,
	hssProps *mo.HostStorageSystem,
	targets []types.HostInternetScsiHbaSendTarget,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	if _, err := methods.AddInternetScsiSendTargets(ctx, client, &types.AddInternetScsiSendTargets{
		This:           hssProps.Reference(),
		IScsiHbaDevice: adapterID,
		Targets:        targets,
	}); err != nil {
		return fmt.Errorf("error trying to add send targets for iscsi software adapter: %s", err)
	}

	return nil
}

/////////////////////////
// Schemas Helpers
/////////////////////////

func ChapSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Description: "The chap credentials for iscsi devices",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"method": {
					Type:         schema.TypeString,
					Default:      "unidirectional",
					Description:  "Chap auth method.  Valid options are 'unidirectional' and 'bidirectional'",
					ValidateFunc: validation.StringInSlice([]string{"unidirectional", "bidirectional"}, true),
				},
				"outgoing_creds": {
					Type:        schema.TypeList,
					Required:    true,
					MaxItems:    1,
					Description: "Outgoing creds for iscsi device",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"username": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Username to auth against iscsi device",
							},
							"password": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Password to auth against iscsi device",
								Sensitive:   true,
							},
						},
					},
				},
				"incoming_creds": {
					Type: schema.TypeList,
					//Required:    true,
					MaxItems:    1,
					Description: "Incoming creds for iscsi device to auth host",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"username": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Username to auth against host",
							},
							"password": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Password to auth against host",
								Sensitive:   true,
							},
						},
					},
				},
			},
		},
	}
}

func IPSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		Description:  "IP of the iscsi target",
		ValidateFunc: validation.IsCIDR,
	}
}

func PortSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeInt,
		Default:      3260,
		Description:  "Port of the iscsi target",
		ValidateFunc: validation.IsPortNumber,
	}
}
