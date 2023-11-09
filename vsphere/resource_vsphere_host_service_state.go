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
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/govmomi/vim25/types"
)

var (
	serviceKeyList = []string{
		string(hostservicestate.HostServiceKeyDCUI),
		string(hostservicestate.HostServiceKeyShell),
		string(hostservicestate.HostServiceKeySSH),
		string(hostservicestate.HostServiceKeyAttestd),
		string(hostservicestate.HostServiceKeyDPD),
		string(hostservicestate.HostServiceKeyKMXD),
		string(hostservicestate.HostServiceKeyLoadBasedTeaming),
		string(hostservicestate.HostServiceKeyActiveDirectory),
		string(hostservicestate.HostServiceKeyNTPD),
		string(hostservicestate.HostServiceKeySmartCard),
		string(hostservicestate.HostServiceKeyPTPD),
		string(hostservicestate.HostServiceKeyCIMServer),
		string(hostservicestate.HostServiceKeySLPD),
		string(hostservicestate.HostServiceKeySNMPD),
		string(hostservicestate.HostServiceKeyVLTD),
		string(hostservicestate.HostServiceKeySyslogServer),
		string(hostservicestate.HostServiceKeyHAAgent),
		string(hostservicestate.HostServiceKeyVcenterAgent),
		string(hostservicestate.HostServiceKeyXORG),
	}
)

// func resourceVsphereHostServiceState() *schema.Resource {
// 	return &schema.Resource{
// 		Create: resourceVSphereHostServiceStateCreate,
// 		Read:   resourceVSphereHostServiceStateRead,
// 		Update: resourceVSphereHostServiceStateUpdate,
// 		Delete: resourceVSphereHostServiceStateDelete,
// 		Importer: &schema.ResourceImporter{
// 			StateContext: resourceVSphereHostServiceStateImport,
// 		},

// 		Schema: map[string]*schema.Schema{
// 			"host_system_id": {
// 				Type:        schema.TypeString,
// 				Required:    true,
// 				ForceNew:    true,
// 				Description: "Host id of machine that will update service",
// 			},
// 			"key": {
// 				Type:         schema.TypeString,
// 				Required:     true,
// 				Description:  "Key for service to update state on given host.  Options: " + hostservicestate.GetServiceKeyMsg(serviceKeyList),
// 				ValidateFunc: validation.StringInSlice(serviceKeyList, false),
// 			},
// 			"running": {
// 				Type:        schema.TypeBool,
// 				Optional:    true,
// 				Default:     false,
// 				Description: "Determines whether the service should be on or off.  Default: 'off'",
// 			},
// 			"policy": {
// 				Type:        schema.TypeString,
// 				Optional:    true,
// 				Default:     types.HostServicePolicyOff,
// 				Description: "The policy of the service.  Valid options are 'on', 'off', or 'automatic'.  Default: 'off'",
// 				ValidateFunc: validation.StringInSlice(
// 					[]string{
// 						string(types.HostServicePolicyOn),
// 						string(types.HostServicePolicyOff),
// 						string(types.HostServicePolicyAutomatic),
// 					},
// 					false,
// 				),
// 			},
// 		},
// 	}
// }

// func resourceVSphereHostServiceStateRead(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*Client).vimClient
// 	hostID := d.Get("host_system_id").(string)
// 	ss, err := hostservicestate.GetServiceState(client, hostID, d.Get("key").(string), provider.DefaultAPITimeout)
// 	if err != nil {
// 		return fmt.Errorf(
// 			"error trying to retrieve service state for host '%s': %s",
// 			hostID,
// 			err,
// 		)
// 	}

// 	d.Set("host_system_id", hostID)
// 	d.Set("key", ss.Key)
// 	d.Set("running", ss.Running)
// 	d.Set("policy", ss.Policy)

// 	return nil
// }

// func resourceVSphereHostServiceStateCreate(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*Client).vimClient
// 	err := hostservicestate.UpdateServiceState(d, client, true)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func resourceVSphereHostServiceStateUpdate(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*Client).vimClient
// 	err := hostservicestate.UpdateServiceState(d, client, false)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func resourceVSphereHostServiceStateDelete(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*Client).vimClient
// 	hostID := d.Get("host_system_id").(string)
// 	err := hostservicestate.SetServiceState(
// 		client,
// 		hostservicestate.ServiceState{
// 			HostSystemID: hostID,
// 			Key:          hostservicestate.HostServiceKey(d.Get("key").(string)),
// 			Policy:       types.HostServicePolicyOff,
// 			Running:      false,
// 		},
// 		provider.DefaultAPITimeout,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("error trying to set service state for host '%s': %s", hostID, err)
// 	}

// 	return nil
// }

// func resourceVSphereHostServiceStateImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
// 	client := meta.(*Client).vimClient
// 	id := strings.Split(d.Id(), ":")

// 	if len(id) != 2 {
// 		return nil, fmt.Errorf("invalid import format.  Format should be: <host_system_id>:<key>")
// 	}

// 	d.Set("host_system_id", id[0])
// 	d.Set("key", id[1])

// 	_, err := hostservicestate.GetServiceState(client, id[0], id[1], provider.DefaultAPITimeout)
// 	if err != nil {
// 		return nil, fmt.Errorf(
// 			"error trying to retrieve service state for host '%s': %s",
// 			id[0],
// 			err,
// 		)
// 	}

// 	d.SetId(fmt.Sprintf("%s:%s", id[0], id[1]))
// 	return []*schema.ResourceData{d}, nil
// }

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
							Description:  "Key for service to update state on given host.  Options: " + hostservicestate.GetServiceKeyMsg(serviceKeyList),
							ValidateFunc: validation.StringInSlice(serviceKeyList, false),
						},
						"running": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Determines whether the service should be on or off.  Default: 'off'",
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
	log.Printf("[DEBUG] hitting service state read")

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
	client := meta.(*Client).vimClient

	err := hostservicestate.UpdateServiceState(d, client, true)
	if err != nil {
		return err
	}

	return nil
}

func resourceVSphereHostServiceStateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	err := hostservicestate.UpdateServiceState(d, client, false)
	if err != nil {
		return err
	}

	return nil
}

func resourceVSphereHostServiceStateDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	srvs := d.Get("service").([]interface{})

	for _, v := range srvs {
		srv := v.(map[string]interface{})
		srv["running"] = false
		srv["policy"] = "off"
		err := hostservicestate.SetServiceState(
			client,
			hostID,
			srv,
			provider.DefaultAPITimeout,
		)
		if err != nil {
			return fmt.Errorf(
				"error trying to set service state for host '%s' with key '%s': %s",
				hostID,
				srv["key"].(string),
				err,
			)
		}
	}

	return nil
}

func resourceVSphereHostServiceStateImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Client).vimClient
	_, err := hostsystem.FromID(client, d.Id())
	if err != nil {
		return nil, fmt.Errorf("error while trying to retrieve host '%s': %s", d.Id(), err)
	}

	hsList, err := hostservicestate.GetHostServies(client, d.Id(), provider.DefaultAPITimeout)
	if err != nil {
		return nil, fmt.Errorf("error while trying to retrieve host services: %s", err)
	}

	srvs := make([]interface{}, 0, len(hsList))

	for _, hostSrv := range hsList {
		if hostSrv.Running || hostSrv.Policy != "off" {
			srvs = append(srvs, map[string]interface{}{
				"key":     hostSrv.Key,
				"policy":  hostSrv.Policy,
				"ruuning": true,
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
