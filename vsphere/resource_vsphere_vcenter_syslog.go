package vsphere

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
	"github.com/vmware/govmomi/vapi/appliance/logging"
)

const (
	vAppSyslogID = "tf-vcenter-syslog"
)

func resourceVSphereVcenterSyslog() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereVcenterSyslogCreate,
		Read:   resourceVSphereVcenterSyslogRead,
		Delete: resourceVSphereVcenterSyslogDelete,
		Update: resourceVSphereVcenterSyslogUpdate,
		Importer: &schema.ResourceImporter{
			State: resourceVSphereVcenterSyslogImport,
		},

		Schema: map[string]*schema.Schema{
			"log_server": {
				Type:        schema.TypeSet,
				Required:    true,
				MaxItems:    3,
				Description: "The log servers to forward logs to",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Host to forward syslog logs",
						},
						"port": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Port of host to forward logs to",
						},
						"protocol": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "tls",
							Description: "Protocol to send to host",
							ValidateFunc: validation.StringInSlice(
								[]string{"TLS", "TCP", "RELP", "UDP"},
								false,
							),
						},
					},
				},
			},
		},
	}
}

func resourceVSphereVcenterSyslogCreate(d *schema.ResourceData, meta interface{}) error {
	err := vsphereVCenterSyslogForwardingUpdate(d, meta, true)
	if err != nil {
		return fmt.Errorf("error creating syslog configurations: %s", err)
	}

	d.SetId(vAppSyslogID)
	return nil
}

func resourceVSphereVcenterSyslogRead(d *schema.ResourceData, meta interface{}) error {
	err := vsphereVcenterSyslogForwardingRead(d, meta)
	if err != nil {
		return fmt.Errorf("error retrieving log configuration info in read function: %s", err)
	}

	return nil
}

func resourceVSphereVcenterSyslogUpdate(d *schema.ResourceData, meta interface{}) error {
	err := vsphereVCenterSyslogForwardingUpdate(d, meta, true)
	if err != nil {
		return fmt.Errorf("error updating syslog configurations: %s", err)
	}

	return nil
}

func resourceVSphereVcenterSyslogDelete(d *schema.ResourceData, meta interface{}) error {
	err := vsphereVCenterSyslogForwardingUpdate(d, meta, false)
	if err != nil {
		return fmt.Errorf("error deleting syslog configurations: %s", err)
	}

	return nil
}

func resourceVSphereVcenterSyslogImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if d.Id() != vAppSyslogID {
		return nil, fmt.Errorf("invalid import.  Import should simply be '%s'", vAppSyslogID)
	}

	err := vsphereVcenterSyslogForwardingRead(d, meta)
	if err != nil {
		return nil, fmt.Errorf("error retrieving log configuration info in import function: %s", err)
	}

	d.SetId(vAppSyslogID)
	return []*schema.ResourceData{d}, nil
}

func vsphereVcenterSyslogForwardingRead(d *schema.ResourceData, meta interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	client := meta.(*Client).restClient
	lm := logging.NewManager(client)

	logs, err := lm.Forwarding(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving log configuration: %s", err)
	}

	logList := make([]interface{}, 0, len(logs))

	for _, log := range logs {
		logList = append(logList, map[string]interface{}{
			"hostname": log.Hostname,
			"port":     log.Port,
			"protocol": log.Protocol,
		})
	}

	d.Set("log_server", logList)
	return nil
}

func vsphereVCenterSyslogForwardingUpdate(d *schema.ResourceData, meta interface{}, isUpdate bool) error {
	client := meta.(*Client).restClient
	var reqBody map[string]interface{}

	if isUpdate {
		reqBody = map[string]interface{}{
			"cfg_list": d.Get("log_server").(*schema.Set).List(),
		}
	} else {
		reqBody = map[string]interface{}{
			"cfg_list": []interface{}{},
		}
	}

	err := viapi.RestUpdateRequest(client, http.MethodPut, "/appliance/logging/forwarding", reqBody)
	if err != nil {
		return fmt.Errorf("error on syslog update request: %s", err)
	}

	return nil
}
