// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"log"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/vmware/govmomi/vim25/mo"
)

func resourceVSphereHostVirtualSwitch() *schema.Resource {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "The name of the virtual switch.",
			Required:    true,
			ForceNew:    true,
		},
		"host_system_id": {
			Type:         schema.TypeString,
			Description:  "The managed object ID of the host to set the virtual switch up on.",
			Optional:     true,
			ForceNew:     true,
			ExactlyOneOf: []string{"hostname"},
		},
		"hostname": {
			Type:        schema.TypeString,
			Description: "The hostname of host to set the virtual switch up on.",
			Optional:    true,
			ForceNew:    true,
		},
	}
	structure.MergeSchema(s, schemaHostVirtualSwitchSpec())

	// Transform any necessary fields in the schema that need to be updated
	// specifically for this resource.
	s["active_nics"].Required = true
	s["standby_nics"].Optional = true

	s["teaming_policy"].Default = hostNetworkPolicyNicTeamingPolicyModeLoadbalanceSrcID
	s["check_beacon"].Default = false
	s["notify_switches"].Default = true
	s["failback"].Default = true

	s["allow_promiscuous"].Default = false
	s["allow_forged_transmits"].Default = true
	s["allow_mac_changes"].Default = true

	s["shaping_enabled"].Default = false

	return &schema.Resource{
		Create:        resourceVSphereHostVirtualSwitchCreate,
		Read:          resourceVSphereHostVirtualSwitchRead,
		Update:        resourceVSphereHostVirtualSwitchUpdate,
		Delete:        resourceVSphereHostVirtualSwitchDelete,
		CustomizeDiff: resourceVSphereHostVirtualSwitchCustomizeDiff,
		Importer: &schema.ResourceImporter{
			State: resourceVSphereHostVirtualSwitchImport,
		},
		Schema: s,
	}
}

func resourceVSphereHostVirtualSwitchCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	name := d.Get("name").(string)
	host, hr, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return err
	}

	ns, err := hostNetworkSystemFromHostSystem(host)
	if err != nil {
		return fmt.Errorf("error loading host network system: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	spec := expandHostVirtualSwitchSpec(d)
	if err := ns.AddVirtualSwitch(ctx, name, spec); err != nil {
		return fmt.Errorf("error adding host vSwitch: %s", err)
	}

	saveHostVirtualSwitchID(d, hr.Value, name)

	return resourceVSphereHostVirtualSwitchRead(d, meta)
}

func resourceVSphereHostVirtualSwitchRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hsID, name, err := virtualSwitchIDsFromResourceID(d)
	if err != nil {
		return err
	}
	ns, err := hostNetworkSystemFromHostSystemID(client, hsID)
	if err != nil {
		return fmt.Errorf("error loading host network system: %s", err)
	}

	sw, err := hostVSwitchFromName(client, ns, name)
	if err != nil {
		return fmt.Errorf("error fetching virtual switch data: %s", err)
	}

	if err := flattenHostVirtualSwitchSpec(d, &sw.Spec); err != nil {
		return fmt.Errorf("error setting resource data: %s", err)
	}

	return nil
}

func resourceVSphereHostVirtualSwitchUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hsID, name, err := virtualSwitchIDsFromResourceID(d)
	if err != nil {
		return err
	}
	ns, err := hostNetworkSystemFromHostSystemID(client, hsID)
	if err != nil {
		return fmt.Errorf("error loading host network system: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	spec := expandHostVirtualSwitchSpec(d)
	if err := ns.UpdateVirtualSwitch(ctx, name, *spec); err != nil {
		return fmt.Errorf("error updating host vSwitch: %s", err)
	}

	return resourceVSphereHostVirtualSwitchRead(d, meta)
}

func resourceVSphereHostVirtualSwitchDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hsID, name, err := virtualSwitchIDsFromResourceID(d)
	if err != nil {
		return err
	}
	ns, err := hostNetworkSystemFromHostSystemID(client, hsID)
	if err != nil {
		return fmt.Errorf("error loading host network system: %s", err)
	}

	sw, err := hostVSwitchFromName(client, ns, name)
	if err != nil {
		return fmt.Errorf("error fetching virtual switch data: %s", err)
	}

	var moNs mo.HostNetworkSystem

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	if err = ns.Properties(ctx, ns.Reference(), nil, &moNs); err != nil {
		return fmt.Errorf("error fetching host network system properties")
	}

	for _, pg := range moNs.NetworkInfo.Portgroup {
		if pg.Spec.VswitchName == sw.Name && pg.Spec.Name == "Management Network" {
			log.Printf(
				"[DEBUG] Deleting host vswitch '%s' from tf state but not actually removing vswitch from host as this host "+
					"contains 'Management Network' port group which can't be deleted from host",
				sw.Name,
			)
			return nil
		}
	}

	if err := ns.RemoveVirtualSwitch(ctx, name); err != nil {
		return fmt.Errorf("error deleting host vSwitch: %s", err)
	}

	return nil
}

func resourceVSphereHostVirtualSwitchImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Client).vimClient
	hostID, switchName, err := splitHostVirtualSwitchID(d.Id())
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	_, hr, err := hostsystem.CheckIfHostnameOrID(client, hostID)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	if err = d.Set(hr.IDName, hr.Value); err != nil {
		return []*schema.ResourceData{}, err
	}
	if err = d.Set("name", switchName); err != nil {
		return []*schema.ResourceData{}, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceVSphereHostVirtualSwitchCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	// We want to quickly validate that each NIC that is in either active_nics or
	// standby_nics will be a part of the bridge.
	bridgeNics := d.Get("network_adapters").([]interface{})
	activeNics := d.Get("active_nics").([]interface{})
	standbyNics := d.Get("standby_nics").([]interface{})

	for _, v := range activeNics {
		var found bool
		for _, w := range bridgeNics {
			if v == w {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("active NIC entry %q not present in network_adapters list", v)
		}
	}

	for _, v := range standbyNics {
		var found bool
		for _, w := range bridgeNics {
			if v == w {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("standby NIC entry %q not present in network_adapters list", v)
		}
	}

	return nil
}
