// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hostsystem

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

var (
	ErrHostnameNotFound = errors.New("could not find host with given hostname")
)

// SystemOrDefault returns a HostSystem from a specific host name and
// datacenter. If the user is connecting over ESXi, the default host system is
// used.
func SystemOrDefault(client *govmomi.Client, name string, dc *object.Datacenter) (*object.HostSystem, error) {
	finder := find.NewFinder(client.Client, false)
	finder.SetDatacenter(dc)

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	t := client.ServiceContent.About.ApiType
	switch t {
	case "HostAgent":
		return finder.DefaultHostSystem(ctx)
	case "VirtualCenter":
		if name != "" {
			return finder.HostSystem(ctx, name)
		}
		return finder.DefaultHostSystem(ctx)
	}
	return nil, fmt.Errorf("unsupported ApiType: %s", t)
}

// FromID locates a HostSystem by its managed object reference ID.
func FromID(client *govmomi.Client, id string) (*object.HostSystem, error) {
	log.Printf("[DEBUG] Locating host system ID %s", id)
	finder := find.NewFinder(client.Client, false)

	ref := types.ManagedObjectReference{
		Type:  "HostSystem",
		Value: id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	hs, err := finder.ObjectReference(ctx, ref)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Host system found: %s", hs.Reference().Value)
	return hs.(*object.HostSystem), nil
}

// FromHostname locates a HostSystem by hostname
// Will return error type "ErrHostnameNotFound" if no host is found
func FromHostname(client *govmomi.Client, hostname string) (*object.HostSystem, error) {
	log.Printf("[DEBUG] Locating host system with hostname %s", hostname)
	finder := find.NewFinder(client.Client, false)

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	dcs, err := finder.DatacenterList(ctx, "*")
	if err != nil {
		return nil, err
	}

	counter := 0

	var host *object.HostSystem

	for _, dc := range dcs {
		viewMgr := view.NewManager(client.Client)
		dcView, err := viewMgr.CreateContainerView(ctx, dc.Reference(), []string{"HostSystem"}, true)
		if err != nil {
			fmt.Printf("error trying to create container: %s", err)
			os.Exit(1)
		}

		var moHosts []mo.HostSystem
		if err = dcView.RetrieveWithFilter(ctx, []string{"HostSystem"}, []string{"name"}, &moHosts, property.Filter{}); err != nil {
			fmt.Printf("error trying to retrieve hosts: %s", err)
			os.Exit(1)
		}

		// Loop through hosts for given dc and determine if exists
		// Throws error if can't find
		for _, h := range moHosts {
			if h.Name == hostname {
				counter++

				if counter > 1 {
					return nil, fmt.Errorf("more than one host with hostname '%s' was found", hostname)
				}

				f := find.NewFinder(client.Client, true)
				ref, err := f.ObjectReference(ctx, h.Self)
				if err != nil {
					return nil, fmt.Errorf("error trying to retrieve host object reference: %s", err)
				}

				host = ref.(*object.HostSystem)
			}
		}
	}

	if host != nil {
		return host, nil
	}

	return nil, ErrHostnameNotFound
}

// FromHostnameOrID locates HostSystem by either id or hostname depending on what's passed through ResourceData
// Returns hostsystem along with identifier used to get hostsystem passed from ResourceData
func FromHostnameOrID(client *govmomi.Client, d *schema.ResourceData) (*object.HostSystem, string, error) {
	var hostID string
	var host *object.HostSystem
	var err error

	if d.Get("host_system_id").(string) != "" {
		hostID = d.Get("host_system_id").(string)
		host, err = FromID(client, d.Get("host_system_id").(string))
	} else {
		hostID = d.Get("hostname").(string)
		host, err = FromHostname(client, d.Get("hostname").(string))
	}

	if err != nil {
		return nil, "", fmt.Errorf("error while trying to retrieve host '%s': %s", hostID, err)
	}

	return host, hostID, nil
}

// Properties is a convenience method that wraps fetching the HostSystem MO
// from its higher-level object.
func Properties(host *object.HostSystem) (*mo.HostSystem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	var props mo.HostSystem
	if err := host.Properties(ctx, host.Reference(), nil, &props); err != nil {
		return nil, err
	}
	return &props, nil
}

// HostStorageSystemProperties is a convenience method that wraps fetching the HostStorageSystem MO
// from its higher-level object.
func HostStorageSystemProperties(hss *object.HostStorageSystem) (*mo.HostStorageSystem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	var props mo.HostStorageSystem
	if err := hss.Properties(ctx, hss.Reference(), nil, &props); err != nil {
		return nil, err
	}
	return &props, nil
}

// GetHostStorageSystemPropertiesFromHost is util helper that grabs the storage system properties for given host
func GetHostStorageSystemPropertiesFromHost(client *govmomi.Client, hostID string) (*mo.HostStorageSystem, error) {
	hss, err := GetHostStorageSystemFromHost(client, hostID)
	if err != nil {
		return nil, err
	}

	hssProps, err := HostStorageSystemProperties(hss)
	if err != nil {
		return nil, fmt.Errorf("error trying to retrieve host storage system properties for host '%s': %s", hostID, err)
	}

	return hssProps, nil
}

// GetHostStorageSystemFromHost is util helper that grabs the storage system properties for given host
func GetHostStorageSystemFromHost(client *govmomi.Client, hostID string) (*object.HostStorageSystem, error) {
	hs, err := FromID(client, hostID)
	if err != nil {
		if viapi.IsManagedObjectNotFoundError(err) {
			return nil, fmt.Errorf("could not find host with id '%s'", hostID)
		}

		return nil, fmt.Errorf("error while searching host '%s': %s", hostID, err)
	}

	hsProps, err := Properties(hs)
	if err != nil {
		return nil, fmt.Errorf("error trying to retrieve host system properties for host '%s': %s", hostID, err)
	}

	return object.NewHostStorageSystem(client.Client, *hsProps.ConfigManager.StorageSystem), nil
}

// ResourcePool is a convenience method that wraps fetching the host system's
// root resource pool
func ResourcePool(host *object.HostSystem) (*object.ResourcePool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	return host.ResourcePool(ctx)
}

// hostSystemNameFromID returns the name of a host via its its managed object
// reference ID.
func hostSystemNameFromID(client *govmomi.Client, id string) (string, error) {
	hs, err := FromID(client, id)
	if err != nil {
		return "", err
	}
	return hs.Name(), nil
}

// NameOrID is a convenience method mainly for helping displaying friendly
// errors where space is important - it displays either the host name or the ID
// if there was an error fetching it.
func NameOrID(client *govmomi.Client, id string) string {
	name, err := hostSystemNameFromID(client, id)
	if err != nil {
		return id
	}
	return name
}

// HostInMaintenance checks a HostSystem's maintenance mode and returns true if the
// the host is in maintenance mode.
func HostInMaintenance(host *object.HostSystem) (bool, error) {
	hostObject, err := Properties(host)
	if err != nil {
		return false, err
	}

	return hostObject.Runtime.InMaintenanceMode, nil
}

// EnterMaintenanceMode puts a host into maintenance mode. If evacuate is set
// to true, all powered off VMs will be removed from the host, or the task will
// block until this is the case, depending on whether or not DRS is on or off
// for the host's cluster. This parameter is ignored on direct ESXi.
func EnterMaintenanceMode(host *object.HostSystem, timeout time.Duration, evacuate bool) error {
	if err := viapi.VimValidateVirtualCenter(host.Client()); err != nil {
		evacuate = false
	}

	maintMode, err := HostInMaintenance(host)
	if err != nil {
		return err
	}
	if maintMode {
		log.Printf("[DEBUG] Host %q is already in maintenance mode", host.Name())
		return nil
	}

	log.Printf("[DEBUG] Host %q is entering maintenance mode (evacuate: %t)", host.Name(), evacuate)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	task, err := host.EnterMaintenanceMode(ctx, int32(timeout.Seconds()), evacuate, nil)
	if err != nil {
		return err
	}

	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	var to mo.Task
	err = task.Properties(context.TODO(), task.Reference(), nil, &to)
	if err != nil {
		log.Printf("[DEBUG] Failed while getting task results: %s", err)
		return err
	}

	if to.Info.State != "success" {
		return fmt.Errorf("error while putting host(%s) in maintenance mode: %s", host.Reference(), to.Info.Error)
	}
	return nil
}

// ExitMaintenanceMode takes a host out of maintenance mode.
func ExitMaintenanceMode(host *object.HostSystem, timeout time.Duration) error {
	maintMode, err := HostInMaintenance(host)
	if err != nil {
		return err
	}
	if !maintMode {
		log.Printf("[DEBUG] Host %q is already not in maintenance mode", host.Name())
		return nil
	}

	log.Printf("[DEBUG] Host %q is exiting maintenance mode", host.Name())

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	task, err := host.ExitMaintenanceMode(ctx, int32(timeout.Seconds()))
	if err != nil {
		return err
	}

	err = task.Wait(ctx)
	if err != nil {
		return err
	}
	var to mo.Task
	err = task.Properties(context.TODO(), task.Reference(), nil, &to)
	if err != nil {
		log.Printf("[DEBUG] Failed while getting task results: %s", err)
		return err
	}

	if to.Info.State != "success" {
		return fmt.Errorf("error while getting host(%s) out of maintenance mode: %s", host.Reference(), to.Info.Error)
	}
	return nil
}

// GetConnectionState returns the host's connection state (see vim.HostSystem.ConnectionState)
func GetConnectionState(host *object.HostSystem) (types.HostSystemConnectionState, error) {
	hostProps, err := Properties(host)
	if err != nil {
		return "", err
	}

	return hostProps.Runtime.ConnectionState, nil
}
