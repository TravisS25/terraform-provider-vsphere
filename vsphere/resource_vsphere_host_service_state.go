// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"log"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostservicestate"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/govmomi/vim25/types"
)

func resourceVsphereHostServiceState() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereHostServiceStateCreate,
		Read:   resourceVSphereHostServiceStateRead,
		Update: resourceVSphereHostServiceStateUpdate,
		Delete: resourceVSphereHostServiceStateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVSphereHostServiceStateImport,
		},
		CustomizeDiff: resourceVSphereHostServiceStateCustomDiff,

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Host id of machine that will update service",
			},
			"service": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The service state object",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Key for service to update state on given host.  Options: " + hostservicestate.GetServiceKeyMsg(),
							ValidateFunc: validation.StringInSlice(hostservicestate.ServiceKeyList, false),
						},
						"policy": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     types.HostServicePolicyOff,
							Description: "The policy of the service.  Valid options are 'on', 'off', or 'automatic'.  Default: 'off'",
							ValidateFunc: validation.StringInSlice(
								[]string{
									string(types.HostServicePolicyOn),
									string(types.HostServicePolicyOff),
									string(types.HostServicePolicyAutomatic),
								},
								false,
							),
						},
					},
				},
			},
		},
	}
}

func resourceVSphereHostServiceStateRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] entering resource_vsphere_host_service_state read function")

	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	srvs := d.Get("service").([]interface{})
	updatedList := make([]interface{}, 0, len(srvs))

	for _, v := range srvs {
		srv := v.(map[string]interface{})
		ss, err := hostservicestate.GetServiceState(
			client,
			hostID,
			hostservicestate.HostServiceKey(srv["key"].(string)),
			provider.DefaultAPITimeout,
		)
		if err != nil {
			return fmt.Errorf(
				"error trying to retrieve service state for host '%s' with key %s: %s",
				hostID,
				srv["key"].(string),
				err,
			)
		}

		updatedList = append(updatedList, ss)
	}

	d.Set("host_system_id", hostID)
	d.Set("service", updatedList)

	return nil
}

func resourceVSphereHostServiceStateCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] entering resource_vsphere_host_service_state create function")

	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	srvs := d.Get("service").([]interface{})

	log.Printf("[INFO] creating host services for host '%s'", hostID)

	for _, v := range srvs {
		srv := v.(map[string]interface{})

		if err := hostservicestate.SetServiceState(
			client,
			hostID,
			srv,
			provider.DefaultAPITimeout,
			true,
		); err != nil {
			return fmt.Errorf(
				"error trying to create service '%s' for host '%s': %s", srv["key"],
				hostID,
				err,
			)
		}
	}

	d.SetId(hostID)

	return nil
}

func resourceVSphereHostServiceStateUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] entering resource_vsphere_host_service_state update function")

	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	log.Printf("[INFO] updating host services for host '%s'", hostID)

	var err error

	oldVal, newVal := d.GetChange("service")
	oldList := oldVal.([]interface{})
	newList := newVal.([]interface{})

	// If we don't find the same service key in new list compared to old list when looping
	// we then destroy/deactivate the service within the old list by stopping the service
	// and setting the policy to "off"
	for _, v := range oldList {
		oldSrv := v.(map[string]interface{})
		found := false

		for _, t := range newList {
			newSrv := t.(map[string]interface{})

			if newSrv["key"].(string) == oldSrv["key"].(string) {
				found = true
				break
			}
		}

		if !found {
			if err = hostservicestate.SetServiceState(
				client,
				hostID,
				oldSrv,
				provider.DefaultAPITimeout,
				false,
			); err != nil {
				return fmt.Errorf(
					"error trying to update old service '%s' for host '%s': %s",
					oldSrv["key"].(string),
					hostID,
					err,
				)
			}
		}
	}

	for _, v := range newList {
		newSrv := v.(map[string]interface{})

		if err = hostservicestate.SetServiceState(
			client,
			hostID,
			newSrv,
			provider.DefaultAPITimeout,
			true,
		); err != nil {
			return fmt.Errorf(
				"error trying to update new service '%s' for host '%s': %s",
				newSrv["key"].(string),
				hostID,
				err,
			)
		}
	}

	return nil
}

func resourceVSphereHostServiceStateDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] entering resource_vsphere_host_service_state delete function")

	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	srvs := d.Get("service").([]interface{})

	log.Printf("[INFO] deleting host services for host '%s'", hostID)

	for _, v := range srvs {
		srv := v.(map[string]interface{})
		err := hostservicestate.SetServiceState(
			client,
			hostID,
			srv,
			provider.DefaultAPITimeout,
			false,
		)
		if err != nil {
			return fmt.Errorf(
				"error trying to delete service '%s' for host '%s': %s",
				srv["key"].(string),
				hostID,
				err,
			)
		}
	}

	return nil
}

func resourceVSphereHostServiceStateImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] entering resource host service state import function")

	client := meta.(*Client).vimClient
	hsList, err := hostservicestate.GetHostServies(client, d.Id(), provider.DefaultAPITimeout)
	if err != nil {
		return nil, fmt.Errorf("error retrieving host services for host '%s': %s", d.Id(), err)
	}

	srvs := make([]interface{}, 0, len(hsList))

	log.Printf("[INFO] importing host service states")

	for _, hostSrv := range hsList {
		foundExcludeKey := false

		for _, excludeKey := range hostservicestate.ExcludeServiceKeyList {
			if hostSrv.Key == excludeKey {
				foundExcludeKey = true
			}
		}

		if !foundExcludeKey && hostSrv.Running {
			srvs = append(srvs, map[string]interface{}{
				"key":    hostSrv.Key,
				"policy": hostSrv.Policy,
			})
		}
	}

	d.SetId(d.Id())
	d.Set("host_system_id", d.Id())
	d.Set("service", srvs)

	return []*schema.ResourceData{d}, nil
}

func resourceVSphereHostServiceStateCustomDiff(ctx context.Context, rd *schema.ResourceDiff, meta interface{}) error {
	srvs := rd.Get("service").([]interface{})
	trackerMap := map[string]bool{}

	for _, val := range srvs {
		srv := val.(map[string]interface{})

		if _, ok := trackerMap[srv["key"].(string)]; ok {
			return fmt.Errorf("duplicate values for 'key' attribute in 'service' resource is not allowed")
		}
		trackerMap[srv["key"].(string)] = true
	}

	return nil
}
