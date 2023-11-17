// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iscsi

import (
	"context"
	"fmt"

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

func GetIscsiAdater(hssProps *mo.HostStorageSystem, host, adapterID string) (types.BaseHostHostBusAdapter, error) {
	for _, adapter := range hssProps.StorageDeviceInfo.HostBusAdapter {
		if adapter.GetHostHostBusAdapter().Device == adapterID {
			return adapter, nil
		}
	}

	return nil, fmt.Errorf("could not find iscsi adapter device '%s' for host '%s'", adapterID, host)
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
		return fmt.Errorf("error trying to add static targets for iscsi adapter '%s': %s", adapterID, err)
	}

	return nil
}

func RemoveInternetScsiStaticTarget(
	client *govmomi.Client,
	host,
	adapterID string,
	hssProps *mo.HostStorageSystem,
	target types.HostInternetScsiHbaStaticTarget,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	if _, err := methods.RemoveInternetScsiStaticTargets(ctx, client, &types.RemoveInternetScsiStaticTargets{
		This:           hssProps.Reference(),
		IScsiHbaDevice: adapterID,
		Targets:        []types.HostInternetScsiHbaStaticTarget{target},
	}); err != nil {
		return fmt.Errorf("error trying to remove static targets from iscsi adapter '%s': %s", adapterID, err)
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
		return fmt.Errorf("error trying to add send targets for iscsi adapter '%s': %s", adapterID, err)
	}

	return nil
}

func RemoveInternetScsiSendTarget(
	client *govmomi.Client,
	host,
	adapterID string,
	hssProps *mo.HostStorageSystem,
	target types.HostInternetScsiHbaSendTarget,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	if _, err := methods.RemoveInternetScsiSendTargets(ctx, client, &types.RemoveInternetScsiSendTargets{
		This:           hssProps.Reference(),
		IScsiHbaDevice: adapterID,
		Targets:        []types.HostInternetScsiHbaSendTarget{target},
	}); err != nil {
		return fmt.Errorf("error trying to remove static targets from iscsi adapter '%s': %s", adapterID, err)
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
