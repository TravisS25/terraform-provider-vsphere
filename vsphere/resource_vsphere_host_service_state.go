// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"strings"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostservicestate"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/govmomi/vim25/types"
)

func resourceVsphereHostServiceState() *schema.Resource {
	srvKeyOptMsg := ""
	srvKeyList := []string{
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

	for _, key := range srvKeyList {
		srvKeyOptMsg += "'" + key + "', "
	}

	return &schema.Resource{
		Create: resourceVSphereHostServiceStateCreate,
		Read:   resourceVSphereHostServiceStateRead,
		Update: resourceVSphereHostServiceStateUpdate,
		Delete: resourceVSphereHostServiceStateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVSphereHostServiceStateImport,
		},

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Host id of machine that will update service",
			},
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Key for service to update state on given host.  Options: " + srvKeyOptMsg,
				ValidateFunc: validation.StringInSlice(srvKeyList, false),
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
	}
}

func resourceVSphereHostServiceStateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	ss, err := hostservicestate.GetServiceState(d, client, provider.DefaultAPITimeout)
	if err != nil {
		return fmt.Errorf(
			"error trying to retrieve service state for host '%s': %s",
			hostID,
			err,
		)
	}

	d.Set("host_system_id", hostID)
	d.Set("key", ss.Key)
	d.Set("running", ss.Running)
	d.Set("policy", ss.Policy)

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
	err := hostservicestate.SetServiceState(
		client,
		hostservicestate.ServiceState{
			HostSystemID: hostID,
			Key:          hostservicestate.HostServiceKey(d.Get("key").(string)),
			Policy:       types.HostServicePolicyOff,
			Running:      false,
		},
		provider.DefaultAPITimeout,
	)
	if err != nil {
		return fmt.Errorf("error trying to set service state for host '%s': %s", hostID, err)
	}

	return nil
}

func resourceVSphereHostServiceStateImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Client).vimClient
	id := strings.Split(d.Id(), ":")

	if len(id) != 2 {
		return nil, fmt.Errorf("invalid import format.  Format should be: <host_system_id>:<key>")
	}

	d.Set("host_system_id", id[0])
	d.Set("key", id[1])

	_, err := hostservicestate.GetServiceState(d, client, provider.DefaultAPITimeout)
	if err != nil {
		return nil, fmt.Errorf(
			"error trying to retrieve service state for host '%s': %s",
			id[0],
			err,
		)
	}

	d.SetId(fmt.Sprintf("%s:%s", id[0], id[1]))
	return []*schema.ResourceData{d}, nil
}
