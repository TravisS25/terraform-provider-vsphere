package vsphere

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVSphereVcenterSyslog() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereVcenterSyslogRead,

		Schema: map[string]*schema.Schema{
			"log_server": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The log servers to forward logs to",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Hostname of log server",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Port of log server",
						},
						"protocol": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Protocol of log server",
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereVcenterSyslogRead(d *schema.ResourceData, meta interface{}) error {
	err := vsphereVcenterSyslogForwardingRead(d, meta)
	if err != nil {
		return fmt.Errorf("error retrieving log configuration: %s", err)
	}

	d.SetId(vAppSyslogID)
	return nil
}
