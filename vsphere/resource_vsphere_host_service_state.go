// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostservicestate"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/govmomi/vim25/types"
)

var (
// serviceKeyMap is simply a mapping where the key is what the user chooses
// and the value is the actual key for service
//
// Reason for this is some of the keys are hard to identify by their acronym when
// referencing from a gui point of view so this mapping makes it more clear
//
//	serviceKeyMap = map[string]hostservicestate.HostServiceKey{
//		"dcui":               hostservicestate.HostServiceKeyDCUI,
//		"shell":              hostservicestate.HostServiceKeyShell,
//		"ssh":                hostservicestate.HostServiceKeySSH,
//		"attestd":            hostservicestate.HostServiceKeyAttestd,
//		"dpd":                hostservicestate.HostServiceKeyDPD,
//		"kmxd":               hostservicestate.HostServiceKeyKMXD,
//		"load_based_teaming": hostservicestate.HostServiceKeyLoadBasedTeaming,
//		"active_directory":   hostservicestate.HostServiceKeyActiveDirectory,
//		"ntpd":               hostservicestate.HostServiceKeyNTPD,
//		"smart_card":         hostservicestate.HostServiceKeySmartCard,
//		"ptpd":               hostservicestate.HostServiceKeyPTPD,
//		"cim_server":         hostservicestate.HostServiceKeyCIMServer,
//		"slpd":               hostservicestate.HostServiceKeySLPD,
//		"snmpd":              hostservicestate.HostServiceKeySNMPD,
//		"vltd":               hostservicestate.HostServiceKeyVLTD,
//		"syslog_server":      hostservicestate.HostServiceKeySyslogServer,
//		"ha_agent":           hostservicestate.HostServiceKeyHAAgent,
//		"vcenter_agent":      hostservicestate.HostServiceKeyVcenterAgent,
//		"xorg":               hostservicestate.HostServiceKeyXORG,
//	}
)

func resourceVSphereHostServiceState() *schema.Resource {
	srvKeyOptMsg := ""
	// srvKeyList := make([]string, 0, len(serviceKeyMap))
	// for k := range serviceKeyMap {
	// 	srvKeyOptMsg += "'" + k + "', "
	// 	srvKeyList = append(srvKeyList, k)
	// }

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
			"service_key": {
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
	ss, err := hostservicestate.GetServiceState(d, meta, client, provider.DefaultAPITimeout)
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
	err := hostservicestate.UpdateServiceState(d, meta, client, true)
	if err != nil {
		return err
	}

	return resourceVSphereDatacenterRead(d, meta)
}

func resourceVSphereHostServiceStateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	err := hostservicestate.UpdateServiceState(d, meta, client, false)
	if err != nil {
		return err
	}

	return nil
}

func resourceVSphereHostServiceStateDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	// Find host and get reference to it.
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return fmt.Errorf("error while trying to retrieve host '%s': %s", hostID, err)
	}

	if err = hostservicestate.SetServiceState(
		host,
		provider.DefaultAPITimeout,
		hostservicestate.ServiceState{
			Key:     hostservicestate.HostServiceKey(d.Get("key").(string)),
			Policy:  types.HostServicePolicyOff,
			Running: false,
		},
	); err != nil {
		return fmt.Errorf("error trying to set service state for host '%s': %s", hostID, err)
	}

	return nil
}

func resourceVSphereHostServiceStateImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	ss, err := hostservicestate.GetServiceState(d, meta, client, provider.DefaultAPITimeout)
	if err != nil {
		return nil, fmt.Errorf(
			"error trying to retrieve service state for host '%s': %s",
			hostID,
			err,
		)
	}

	d.SetId(fmt.Sprintf("%s:%s", hostID, ss.Key))
	return []*schema.ResourceData{d}, nil
}
