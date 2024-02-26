// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	vcenterssh "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/ssh"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
	"golang.org/x/crypto/ssh"
)

const (
	vsphereVcenterSnmpID = "tf-vcenter-snmp"
	snmpMonitoringPath   = "/appliance/techpreview/monitoring/snmp"
)

func resourceVSphereVcenterSNMP() *schema.Resource {
	return &schema.Resource{
		Create:        resourceVSphereVcenterSNMPCreate,
		Read:          resourceVSphereVcenterSNMPRead,
		Update:        resourceVSphereVcenterSNMPUpdate,
		Delete:        resourceVSphereVcenterSNMPDelete,
		CustomizeDiff: getSNMPCustomDiff(false),
		Importer: &schema.ResourceImporter{
			StateContext: resourceVSphereVcenterSNMPImport,
		},

		Schema: map[string]*schema.Schema{
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User of host.  Only required if using snmp v3",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password of host.  Only required if using snmp v3",
			},
			"known_hosts_path": {
				Type:     schema.TypeString,
				Optional: true,
				Description: `File path to 'known_hosts' file that must contain the hostname of esxi host.
				This is used to verify a host against their current public ssh key.  Must be full path
				`,
			},
			"ssh_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     22,
				Description: "Port to connect to esxi host for ssh",
			},
			"ssh_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     8,
				Description: "Number in seconds it should take to establish connection before timing out",
			},
			"engine_id": {
				Type:        schema.TypeString,
				Description: "Sets SNMPv3 engine id",
				Required:    true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 32),
					validation.StringMatch(
						regexp.MustCompile("^[0-9a-fA-F]+$"),
						"Must be hexadecimal characters",
					),
				),
			},
			"authentication_protocol": {
				Type:         schema.TypeString,
				Description:  "Protocol used ensure the identity of users of SNMP v3",
				Optional:     true,
				Default:      "none",
				ValidateFunc: validation.StringInSlice([]string{"none", "MD5", "SHA1"}, false),
			},
			"privacy_protocol": {
				Type:         schema.TypeString,
				Description:  "Protocol used to allow encryption of SNMP v3 messages",
				Optional:     true,
				Default:      "none",
				ValidateFunc: validation.StringInSlice([]string{"none", "AES128"}, false),
			},
			"log_level": {
				Type:         schema.TypeString,
				Description:  "Log level the host snmp agent will output",
				Optional:     true,
				Default:      "warning",
				ValidateFunc: validation.StringInSlice([]string{"debug", "info", "warning", "error"}, false),
			},
			"snmp_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     161,
				Description: "Port for the agent listen on",
			},
			"read_only_communities": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Communities that are read only.  Only valid for version 1 and 2",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"remote_user": {
				Type:        schema.TypeSet,
				Description: "Set of users to use for auth against snmp agent",
				Optional:    true,
				MaxItems:    5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of user",
						},
						"authentication_password": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							Description:  "Password to use to auth user",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(".{8,}"), "Must be at least 8 characters"),
						},
						"privacy_secret": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							Description:  "Secret to use for encryption of messages",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(".{16}"), "Must be exactly 16 characters"),
						},
					},
				},
			},
			"trap_target": {
				Type:        schema.TypeSet,
				Description: "Targets to send snmp message",
				Optional:    true,
				MaxItems:    3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Hostname of receiver for notifications from host",
						},
						"port": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Port of receiver for notifications from host",
						},
						"community": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Community of receiver for notifications from host",
						},
					},
				},
			},
		},
	}
}

func resourceVSphereVcenterSNMPCreate(d *schema.ResourceData, meta interface{}) error {
	err := vsphereVcenterSNMPUpdate(d, meta)
	if err != nil {
		return err
	}

	d.SetId(vsphereVcenterSnmpID)
	return nil
}

func resourceVSphereVcenterSNMPRead(d *schema.ResourceData, meta interface{}) error {
	return vsphereVcenterSNMPRead(d, meta)
}

func resourceVSphereVcenterSNMPUpdate(d *schema.ResourceData, meta interface{}) error {
	return vsphereVcenterSNMPUpdate(d, meta)
}

func resourceVSphereVcenterSNMPDelete(d *schema.ResourceData, meta interface{}) error {
	var err error

	client := meta.(*Client).restClient

	if _, err = viapi.RestRequest[[]interface{}](
		client,
		http.MethodPost,
		"/appliance/techpreview/monitoring/snmp/disable",
		nil,
	); err != nil {
		return fmt.Errorf("error enabling snmp for vcenter: %s", err)
	}

	cb := ssh.InsecureIgnoreHostKey()

	if d.Get("known_hosts_path").(string) != "" {
		cb = vcenterssh.GetDefaultHostKeyCallback(d.Get("known_hosts_path").(string))
	}

	if _, err = vcenterssh.RunCommand(
		fmt.Sprintf(
			`
			snmp.set \
				--authentication %s \
				--privacy %s \
				--communities %s \
				--loglevel %s \
				--port %d \
				--remoteusers %s \
				--targets %s
			`,
			"none",
			"none",
			"reset",
			"warning",
			161,
			"reset",
			"reset",
		),
		client.URL().Hostname(),
		d.Get("ssh_port").(int),
		vcenterssh.GetDefaultClientConfig(
			d.Get("user").(string),
			d.Get("password").(string),
			d.Get("ssh_timeout").(int),
			cb,
		),
	); err != nil {
		return fmt.Errorf("error updating snmp settings for vcenter on host: %s", err)
	}

	return nil
}

func resourceVSphereVcenterSNMPImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if d.Id() != vsphereVcenterSnmpID {
		return nil, fmt.Errorf("invalid import.  Import should simply be '%s'", vsphereVcenterSnmpID)
	}

	err := snmpSSHImport(d, false)
	if err != nil {
		return nil, err
	}

	restClient := meta.(*Client).restClient
	valRes, err := viapi.RestRequest[map[string]interface{}](
		restClient,
		http.MethodGet,
		snmpMonitoringPath,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error retrieving snmp settings for vcenter: %s", err)
	}

	ruList := valRes["remoteusers"].([]interface{})
	remoteUsers := make([]map[string]interface{}, 0, len(ruList))

	for _, u := range ruList {
		user := u.(map[string]interface{})
		remoteUsers = append(remoteUsers, map[string]interface{}{
			"name": user["username"],
		})
	}

	d.SetId(vsphereVcenterSnmpID)
	d.Set("remote_user", remoteUsers)
	return []*schema.ResourceData{d}, nil
}

func vsphereVcenterSNMPRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).restClient

	valRes, err := viapi.RestRequest[map[string]interface{}](
		client,
		http.MethodGet,
		snmpMonitoringPath,
		nil,
	)
	if err != nil {
		return fmt.Errorf("error retrieving snmp response from vcenter: %s", err)
	}

	targetList := valRes["targets"].([]interface{})
	trapTargets := make([]map[string]interface{}, 0, len(targetList))
	for _, t := range targetList {
		target := t.(map[string]interface{})
		trapTargets = append(trapTargets, map[string]interface{}{
			"hostname":  target["ip"],
			"port":      target["port"],
			"community": target["community"],
		})
	}

	// For some reason, the api represents a "default" or empty value as array with len of 1
	// with an empty string, so below is a check to only set if first element does not equal
	// empty string
	if len(valRes["communities"].([]interface{})) > 0 && valRes["communities"].([]interface{})[0] != "" {
		d.Set("read_only_communities", valRes["communities"])
	}

	d.Set("engine_id", valRes["engineid"])
	d.Set("authentication_protocol", valRes["authentication"])
	d.Set("privacy_protocol", valRes["privacy"])
	d.Set("log_level", valRes["loglevel"])
	d.Set("snmp_port", valRes["port"])
	d.Set("trap_target", trapTargets)
	return nil
}

func vsphereVcenterSNMPUpdate(d *schema.ResourceData, meta interface{}) error {
	var err error

	client := meta.(*Client).vimClient
	restClient := meta.(*Client).restClient
	cb := ssh.InsecureIgnoreHostKey()

	if d.Get("known_hosts_path").(string) != "" {
		cb = vcenterssh.GetDefaultHostKeyCallback(d.Get("known_hosts_path").(string))
	}

	if _, err = viapi.RestRequest[[]interface{}](
		restClient,
		http.MethodPost,
		"/appliance/techpreview/monitoring/snmp/enable",
		nil,
	); err != nil {
		return fmt.Errorf("error enabling snmp for vcenter: %s", err)
	}

	if _, err = vcenterssh.RunCommand(
		fmt.Sprintf("snmp.set --engineid %s", d.Get("engine_id").(string)),
		client.URL().Hostname(),
		d.Get("ssh_port").(int),
		vcenterssh.GetDefaultClientConfig(
			d.Get("user").(string),
			d.Get("password").(string),
			d.Get("ssh_timeout").(int),
			cb,
		),
	); err != nil {
		return fmt.Errorf("error setting snmp engineid for vcenter with hostname '%s': %s", client.URL().Hostname(), err)
	}

	if _, err = vcenterssh.RunCommand(
		fmt.Sprintf(
			`
			snmp.set \
				--authentication %s \
				--privacy %s
			`,
			d.Get("authentication_protocol").(string),
			d.Get("privacy_protocol").(string),
		),
		client.URL().Hostname(),
		d.Get("ssh_port").(int),
		vcenterssh.GetDefaultClientConfig(
			d.Get("user").(string),
			d.Get("password").(string),
			d.Get("ssh_timeout").(int),
			cb,
		),
	); err != nil {
		return fmt.Errorf("error updating global auth and privacy protocols: %s", err)
	}

	var ttStr string
	ttList := d.Get("trap_target").(*schema.Set).List()

	if len(ttList) == 0 {
		ttStr = "reset"
	} else {
		for i, t := range ttList {
			tt := t.(map[string]interface{})
			port := strconv.Itoa(tt["port"].(int))
			ttStr += tt["hostname"].(string) + "@" + port + "/" + tt["community"].(string)

			if i != len(ttList)-1 {
				ttStr += ","
			}
		}
	}

	var remoteUsersStr string
	remoteUsers := d.Get("remote_user").(*schema.Set).List()

	if len(remoteUsers) == 0 {
		remoteUsersStr = "reset"
	} else {
		for i, u := range remoteUsers {
			user := u.(map[string]interface{})
			resVal, err := viapi.RestRequest[map[string]interface{}](
				restClient,
				http.MethodPost,
				"/appliance/techpreview/monitoring/snmp/hash",
				map[string]interface{}{
					"config": map[string]interface{}{
						"auth_hash":  user["authentication_password"],
						"priv_hash":  user["privacy_secret"],
						"raw_secret": true,
					},
				},
			)
			if err != nil {
				return fmt.Errorf("error hashing auth and privacy for user '%s': %s", user["name"], err)
			}

			remoteUsersStr += user["name"].(string) + "/" + d.Get("authentication_protocol").(string)

			if user["authentication_password"] != "" {
				remoteUsersStr += "/" + resVal["auth_key"].(string)
			} else {
				remoteUsersStr += "/-"
			}

			remoteUsersStr += "/" + d.Get("privacy_protocol").(string)

			if user["privacy_secret"] != "" {
				remoteUsersStr += "/" + resVal["priv_key"].(string)
			} else {
				remoteUsersStr += "/-"
			}

			remoteUsersStr += "/" + d.Get("engine_id").(string)

			if i != len(remoteUsers)-1 {
				remoteUsersStr += ","
			}
		}
	}

	var communityStr string
	communities := d.Get("read_only_communities").(*schema.Set).List()

	if len(communities) == 0 {
		communityStr = "reset"
	} else {
		for i, c := range communities {
			communityStr += c.(string)

			if i != len(communities)-1 {
				communityStr += ","
			}
		}
	}

	if _, err = vcenterssh.RunCommand(
		fmt.Sprintf(
			`
			snmp.set \
				--communities %s \
				--loglevel %s \
				--port %d \
				--remoteusers %s \
				--targets %s
			`,
			communityStr,
			d.Get("log_level").(string),
			d.Get("snmp_port").(int),
			remoteUsersStr,
			ttStr,
		),
		client.URL().Hostname(),
		d.Get("ssh_port").(int),
		vcenterssh.GetDefaultClientConfig(
			d.Get("user").(string),
			d.Get("password").(string),
			d.Get("ssh_timeout").(int),
			cb,
		),
	); err != nil {
		return fmt.Errorf("error updating snmp settings for vcenter on host: %s", err)
	}

	return nil
}
