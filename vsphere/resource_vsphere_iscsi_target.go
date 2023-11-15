package vsphere

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/iscsi"
	"github.com/vmware/govmomi/vim25/types"
)

func resourceVSphereIscsiTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereIscsiTargetCreate,
		Read:   resourceVSphereIscsiTargetRead,
		Update: resourceVSphereIscsiTargetUpdate,
		Delete: resourceVSphereIscsiTargetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVSphereIscsiTargetImport,
		},

		CustomizeDiff: resourceVSphereIscsiTargetCustomDiff,
		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the host system to attach iscsi adapter to",
			},
			"adapter_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Iscsi adapter the iscsi targets will be added to.  This should be in the form of 'vmhb<unique_name>'",
			},
			"discovery_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "dynamic",
				Description:  "Determines what type of iscsi to create.  Valid options are 'dynamic' and 'static'",
				ValidateFunc: validation.StringInSlice([]string{"dynamic", "static"}, true),
			},
			"target": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "IP of the iscsi target",
							ValidateFunc: validation.IsCIDR,
						},
						"port": {
							Type:         schema.TypeInt,
							Default:      3260,
							Description:  "Port of the iscsi target",
							ValidateFunc: validation.IsPortNumber,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The iqn of the storage device if iscsi type is 'static'",
						},
						// default - chap can be optional, if optinal, DO NOT inhreit and auth method should be none
						"chap": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Description: "The chap credentials for iscsi devices",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"method": {
										Type:         schema.TypeString,
										Default:      "unidirectional",
										Description:  "Chap auth method.  Valid options are 'unidirectional' and 'bidirectional'",
										ValidateFunc: validation.StringInSlice([]string{"unidirectional", "bidirectional"}, true),
									},
									"outgoing_creds": {
										Type:        schema.TypeList,
										Required:    true,
										MaxItems:    1,
										Description: "Outgoing creds for iscsi device",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"username": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Username to auth against iscsi device",
												},
												"password": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Password to auth against iscsi device",
													Sensitive:   true,
												},
											},
										},
									},
									"incoming_creds": {
										Type: schema.TypeList,
										//Required:    true,
										MaxItems:    1,
										Description: "Incoming creds for iscsi device to auth host",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"username": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Username to auth against host",
												},
												"password": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Password to auth against host",
													Sensitive:   true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceVSphereIscsiTargetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	hss, err := hostsystem.GetHostStorageSystemFromHost(client, hostID)
	if err != nil {
		return err
	}

	hssProps, err := hostsystem.HostStorageSystemProperties(hss)
	if err != nil {
		return err
	}

	targetList := d.Get("target").([]interface{})

	if d.Get("discovery_type") == "static" {
		targets := make([]types.HostInternetScsiHbaStaticTarget, 0, len(targetList))

		for _, v := range targetList {
			target := v.(map[string]interface{})
			targets = append(targets, types.HostInternetScsiHbaStaticTarget{
				Address:   target["ip"].(string),
				Port:      target["port"].(int32),
				IScsiName: target["name"].(string),
			})
		}

		if err = iscsi.AddInternetScsiStaticTargets(client, hostID, hssProps, targets); err != nil {
			return err
		}
	} else {
		targets := make([]types.HostInternetScsiHbaSendTarget, 0, len(targetList))

		for _, v := range targetList {
			target := v.(map[string]interface{})
			targets = append(targets, types.HostInternetScsiHbaSendTarget{
				Address: target["ip"].(string),
				Port:    target["port"].(int32),
			})
		}

		if err = iscsi.AddInternetScsiSendTargets(client, hostID, hssProps, targets); err != nil {
			return err
		}
	}

	d.SetId(fmt.Sprintf("%s:%s", hostID, d.Get("adapter_id").(string)))

	return nil
}

func resourceVSphereIscsiTargetRead(d *schema.ResourceData, meta interface{}) error {
	//client := meta.(*Client).vimClient

	return nil
}

func resourceVSphereIscsiTargetUpdate(d *schema.ResourceData, meta interface{}) error {
	//client := meta.(*Client).vimClient

	return nil
}

func resourceVSphereIscsiTargetDelete(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceVSphereIscsiTargetImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	//client := meta.(*Client).vimClient

	return []*schema.ResourceData{d}, nil
}

func resourceVSphereIscsiTargetCustomDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	errList := ""
	targets := d.Get("target").([]interface{})

	// if d.Get("discovery_type") == "static" {
	// 	for _, v := range targets {
	// 		target := v.(map[string]interface{})

	// 		if target["name"] == "" {
	// 			errList += fmt.Sprintf("target with ip '%s' must set 'name' attribute when discovery type is 'static'\n", target["ip"])
	// 		}
	// 	}
	// }

	for _, v := range targets {
		if d.Get("discovery_type") == "static" {
			target := v.(map[string]interface{})

			if target["name"] == "" {
				errList += fmt.Sprintf("target with ip '%s' must set 'name' attribute when discovery type is 'static'\n", target["ip"])
			}
		}

	}

	if errList != "" {
		return fmt.Errorf(errList)
	}

	//client := meta.(*Client).vimClient

	return nil
}
