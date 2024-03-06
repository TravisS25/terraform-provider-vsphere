package vsphere

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVSphereVcenterSNMP() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereVcenterSNMPRead,

		Schema: map[string]*schema.Schema{
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "User of vcenter",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password of vcenter",
			},
			"known_hosts_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "File path to 'known_hosts' file that will contain the hostname of esxi host.  Must be full path",
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
				Computed:    true,
			},
			"authentication_protocol": {
				Type:        schema.TypeString,
				Description: "Protocol used ensure the identity of users of SNMP v3",
				Computed:    true,
			},
			"privacy_protocol": {
				Type:        schema.TypeString,
				Description: "Protocol used to allow encryption of SNMP v3 messages",
				Computed:    true,
			},
			"log_level": {
				Type:        schema.TypeString,
				Description: "Log level the host snmp agent will output",
				Computed:    true,
			},
			"remote_user": {
				Type:        schema.TypeSet,
				Description: "Set of users to use for auth against snmp agent",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of user",
						},
					},
				},
			},
			"snmp_port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Port for the agent listen on",
			},
			"read_only_communities": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Communities that are read only.  Only valid for version 1 and 2",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"trap_target": {
				Type:        schema.TypeSet,
				Description: "Targets to send snmp message",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Hostname of receiver for notifications from host",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Port of receiver for notifications from host",
						},
						"community": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Community of receiver for notifications from host",
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereVcenterSNMPRead(d *schema.ResourceData, meta interface{}) error {
	err := vsphereVcenterSNMPRead(d, meta)
	if err != nil {
		return fmt.Errorf("error trying to read snmp settings in data source for vcenter: %s", err)
	}

	d.SetId(vsphereVcenterSnmpID)
	return nil
}
