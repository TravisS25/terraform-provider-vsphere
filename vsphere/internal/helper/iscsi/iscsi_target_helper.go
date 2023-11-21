// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iscsi

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
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

func RescanStorageDevices(hss *object.HostStorageSystem) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	if err := hss.RescanAllHba(ctx); err != nil {
		return fmt.Errorf("error trying to rescan storage devices: %s", err)
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

	log.Printf("[INFO] adding iscsi static targets")

	if _, err := methods.AddInternetScsiStaticTargets(ctx, client, &types.AddInternetScsiStaticTargets{
		This:           hssProps.Reference(),
		IScsiHbaDevice: adapterID,
		Targets:        targets,
	}); err != nil {
		return fmt.Errorf(
			"error trying to add static targets for iscsi adapter '%s' on host '%s': %s",
			adapterID,
			host,
			err,
		)
	}

	return nil
}

func RemoveInternetScsiStaticTargets(
	client *govmomi.Client,
	host,
	adapterID string,
	hssProps *mo.HostStorageSystem,
	targets []types.HostInternetScsiHbaStaticTarget,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	log.Printf("[INFO] removing iscsi static targets")

	if _, err := methods.RemoveInternetScsiStaticTargets(ctx, client, &types.RemoveInternetScsiStaticTargets{
		This:           hssProps.Reference(),
		IScsiHbaDevice: adapterID,
		Targets:        targets,
	}); err != nil {
		return fmt.Errorf(
			"error trying to remove static targets from iscsi adapter '%s' on host '%s': %s",
			adapterID,
			host,
			err,
		)
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

	log.Printf("[INFO] adding iscsi send targets")

	if _, err := methods.AddInternetScsiSendTargets(ctx, client, &types.AddInternetScsiSendTargets{
		This:           hssProps.Reference(),
		IScsiHbaDevice: adapterID,
		Targets:        targets,
	}); err != nil {
		return fmt.Errorf(
			"error trying to add send targets for iscsi adapter '%s' on host '%s': %s",
			adapterID,
			host,
			err,
		)
	}

	return nil
}

func RemoveInternetScsiSendTargets(
	client *govmomi.Client,
	host,
	adapterID string,
	hssProps *mo.HostStorageSystem,
	targets []types.HostInternetScsiHbaSendTarget,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	log.Printf("[INFO] removing iscsi send targets")

	if _, err := methods.RemoveInternetScsiSendTargets(ctx, client, &types.RemoveInternetScsiSendTargets{
		This:           hssProps.Reference(),
		IScsiHbaDevice: adapterID,
		Targets:        targets,
	}); err != nil {
		return fmt.Errorf(
			"error trying to remove send targets from iscsi adapter '%s' on host '%s': %s",
			adapterID,
			host,
			err,
		)
	}

	return nil
}

func ExtractChapCredsFromTarget(target map[string]interface{}, outgoingCreds bool) map[string]interface{} {
	chapList := target["chap"].([]interface{})
	chapCreds := map[string]interface{}{
		"username": "",
		"password": "",
	}

	if len(chapList) > 0 {
		chap := chapList[0].(map[string]interface{})

		if outgoingCreds {
			chapCreds["username"] = chap["outgoing_creds"].([]interface{})[0].(map[string]interface{})["username"]
			chapCreds["password"] = chap["outgoing_creds"].([]interface{})[0].(map[string]interface{})["password"]
		} else if len(chap["incoming_creds"].([]interface{})) > 0 {
			chapCreds["username"] = chap["incoming_creds"].([]interface{})[0].(map[string]interface{})["username"]
			chapCreds["password"] = chap["incoming_creds"].([]interface{})[0].(map[string]interface{})["password"]
		}
	}

	return chapCreds
}

func ExtractTargetUpdates(oldList, newList []interface{}) ([]interface{}, []interface{}) {
	dupTargets := make([]interface{}, 0, len(newList))
	removeTargets := make([]interface{}, 0, len(oldList))

	// Looping and comparing if any of the old targets still exist and remove any targets
	// that have changed
	for _, v := range oldList {
		oldTarget := v.(map[string]interface{})
		found := false

		for _, newTarget := range newList {
			if reflect.DeepEqual(oldTarget, newTarget) {
				found = true
			}
		}

		if found {
			dupTargets = append(dupTargets, oldTarget)
		} else {
			removeTargets = append(removeTargets, oldTarget)
		}
	}

	newTargets := make([]interface{}, 0, len(newList))

	for _, v := range newList {
		newTarget := v.(map[string]interface{})
		found := false

		for _, dupTarget := range dupTargets {
			if reflect.DeepEqual(newTarget, dupTarget) {
				found = true
			}
		}

		if !found {
			newTargets = append(newTargets, newTarget)
		}
	}

	return removeTargets, newTargets
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
								//Sensitive:   true,
							},
						},
					},
				},
				"incoming_creds": {
					Type:        schema.TypeList,
					Optional:    true,
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
								//Sensitive:   true,
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
		ValidateFunc: validation.IsIPv4Address,
	}
}

func PortSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeInt,
		Optional:     true,
		Default:      3260,
		Description:  "Port of the iscsi target",
		ValidateFunc: validation.IsPortNumber,
	}
}
