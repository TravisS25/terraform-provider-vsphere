// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iscsi

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
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

// UpdateIscsiName is util helper that updates iscsi name for adapter
func UpdateIscsiName(host, device, name string, c *govmomi.Client, hssProps types.ManagedObjectReference) error {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	log.Printf("[INFO] updating iscsi software adapter")
	_, err := methods.UpdateInternetScsiName(ctx, c, &types.UpdateInternetScsiName{
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

	log.Printf("[INFO] enabling iscsi software adapter")
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
