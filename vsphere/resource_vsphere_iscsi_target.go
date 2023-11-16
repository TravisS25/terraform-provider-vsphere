package vsphere

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"send_target": {
				Type: schema.TypeSet,
				//Required: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						iscsi.IPResourceKey:   iscsi.IPSchema(),
						iscsi.PortResourceKey: iscsi.PortSchema(),
						iscsi.ChapResourceKey: iscsi.ChapSchema(),
					},
				},
			},
			"static_target": {
				Type: schema.TypeSet,
				//Required: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						iscsi.IPResourceKey:   iscsi.IPSchema(),
						iscsi.PortResourceKey: iscsi.PortSchema(),
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The iqn of the storage device",
						},
						iscsi.ChapResourceKey: iscsi.ChapSchema(),
					},
				},
			},
			// OLD
			//
			// "target": {
			// 	Type:     schema.TypeSet,
			// 	Required: true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"ip": {
			// 				Type:         schema.TypeString,
			// 				Required:     true,
			// 				Description:  "IP of the iscsi target",
			// 				ValidateFunc: validation.IsCIDR,
			// 			},
			// 			"port": {
			// 				Type:         schema.TypeInt,
			// 				Default:      3260,
			// 				Description:  "Port of the iscsi target",
			// 				ValidateFunc: validation.IsPortNumber,
			// 			},
			// 			"name": {
			// 				Type:        schema.TypeString,
			// 				Description: "The iqn of the storage device if 'discovery_type' is 'static'",
			// 			},
			// 			"discovery_type": {
			// 				Type:         schema.TypeString,
			// 				Optional:     true,
			// 				ForceNew:     true,
			// 				Default:      "dynamic",
			// 				Description:  "Determines what type of iscsi to create.  Valid options are 'dynamic' and 'static'",
			// 				ValidateFunc: validation.StringInSlice([]string{"dynamic", "static"}, true),
			// 			},
			// 			// default - chap can be optional, if optional, DO NOT inherit and auth method should be none
			// 			"chap": {
			// 				Type:        schema.TypeList,
			// 				MaxItems:    1,
			// 				Description: "The chap credentials for iscsi devices",
			// 				Optional:    true,
			// 				Elem: &schema.Resource{
			// 					Schema: map[string]*schema.Schema{
			// 						"method": {
			// 							Type:         schema.TypeString,
			// 							Default:      "unidirectional",
			// 							Description:  "Chap auth method.  Valid options are 'unidirectional' and 'bidirectional'",
			// 							ValidateFunc: validation.StringInSlice([]string{"unidirectional", "bidirectional"}, true),
			// 						},
			// 						"outgoing_creds": {
			// 							Type:        schema.TypeList,
			// 							Required:    true,
			// 							MaxItems:    1,
			// 							Description: "Outgoing creds for iscsi device",
			// 							Elem: &schema.Resource{
			// 								Schema: map[string]*schema.Schema{
			// 									"username": {
			// 										Type:        schema.TypeString,
			// 										Required:    true,
			// 										Description: "Username to auth against iscsi device",
			// 									},
			// 									"password": {
			// 										Type:        schema.TypeString,
			// 										Required:    true,
			// 										Description: "Password to auth against iscsi device",
			// 										Sensitive:   true,
			// 									},
			// 								},
			// 							},
			// 						},
			// 						"incoming_creds": {
			// 							Type: schema.TypeList,
			// 							//Required:    true,
			// 							MaxItems:    1,
			// 							Description: "Incoming creds for iscsi device to auth host",
			// 							Elem: &schema.Resource{
			// 								Schema: map[string]*schema.Schema{
			// 									"username": {
			// 										Type:        schema.TypeString,
			// 										Required:    true,
			// 										Description: "Username to auth against host",
			// 									},
			// 									"password": {
			// 										Type:        schema.TypeString,
			// 										Required:    true,
			// 										Description: "Password to auth against host",
			// 										Sensitive:   true,
			// 									},
			// 								},
			// 							},
			// 						},
			// 					},
			// 				},
			// 			},
			// 		},
			// 	},
			// },
		},
	}
}

func resourceVSphereIscsiTargetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)

	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
	if err != nil {
		return err
	}

	targetList := d.Get("target").(*schema.Set).List()
	inherited := false

	for _, v := range targetList {
		target := v.(map[string]interface{})
		authSettings := &types.HostInternetScsiHbaAuthenticationProperties{
			ChapInherited:       &inherited,
			MutualChapInherited: &inherited,
		}

		if c, ok := target["chap"]; ok {
			chap := c.([]interface{})
			outgoingCreds := chap[0].(map[string]interface{})["outgoing_creds"].([]interface{})[0].(map[string]interface{})

			authSettings.ChapName = outgoingCreds["username"].(string)
			authSettings.ChapSecret = outgoingCreds["password"].(string)

			if incomingCredsList, ok := chap[0].(map[string]interface{})["incoming_creds"]; ok {
				incomingCreds := incomingCredsList.([]interface{})[0].(map[string]interface{})
				authSettings.MutualChapName = incomingCreds["username"].(string)
				authSettings.MutualChapSecret = incomingCreds["password"].(string)
			}
		}

		if d.Get("discovery_type") == "static" {
			targets := make([]types.HostInternetScsiHbaStaticTarget, 0, len(targetList))
			targets = append(targets, types.HostInternetScsiHbaStaticTarget{
				Address:                  target["ip"].(string),
				Port:                     target["port"].(int32),
				IScsiName:                target["name"].(string),
				AuthenticationProperties: authSettings,
			})

			if err = iscsi.AddInternetScsiStaticTargets(
				client,
				hostID,
				d.Get("adapter_id").(string),
				hssProps,
				targets,
			); err != nil {
				return err
			}
		} else {
			targets := make([]types.HostInternetScsiHbaSendTarget, 0, len(targetList))
			targets = append(targets, types.HostInternetScsiHbaSendTarget{
				Address:                  target["ip"].(string),
				Port:                     target["port"].(int32),
				AuthenticationProperties: authSettings,
			})

			if err = iscsi.AddInternetScsiSendTargets(
				client,
				hostID,
				d.Get("adapter_id").(string),
				hssProps,
				targets,
			); err != nil {
				return err
			}
		}
	}

	d.SetId(fmt.Sprintf("%s:%s", hostID, d.Get("adapter_id").(string)))

	return nil
}

func resourceVSphereIscsiTargetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	adapterID := d.Get("adapter_id").(string)

	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
	if err != nil {
		return err
	}

	baseAdapter, err := iscsi.GetIscsiAdater(hssProps, hostID, adapterID)
	if err != nil {
		return err
	}

	adapter := baseAdapter.(*types.HostInternetScsiHba)

	sendTargets := make([]map[string]interface{}, 0, len(adapter.ConfiguredSendTarget))
	staticTargets := make([]map[string]interface{}, 0, len(adapter.ConfiguredStaticTarget))

	for _, sendTarget := range adapter.ConfiguredSendTarget {
		target := map[string]interface{}{
			"ip":   sendTarget.Address,
			"port": sendTarget.Port,
		}

		if c, ok := d.GetOk("chap"); ok {
			target["chap"] = c
		}

		sendTargets = append(sendTargets, target)
	}

	for _, staticTarget := range adapter.ConfiguredStaticTarget {
		target := map[string]interface{}{
			"ip":   staticTarget.Address,
			"port": staticTarget.Port,
			"name": staticTarget.IScsiName,
		}

		if c, ok := d.GetOk("chap"); ok {
			target["chap"] = c
		}

		sendTargets = append(sendTargets, target)
	}

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
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	adapterID := d.Get("adapter_id").(string)

	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
	if err != nil {
		return err
	}

	adapter, err := iscsi.GetIscsiAdater(hssProps, hostID, adapterID)
	if err != nil {
		return err
	}

	switch adapter.(type) {
	case *types.HostInternetScsiHba:
		break
	default:
		return fmt.Errorf("adapter_id does not belong to a device that allows static or dynamic discovery")
	}

	_, staticOk := d.GetOk("static_target")
	_, sendOK := d.GetOk("send_target")

	if !staticOk && !sendOK {
		return fmt.Errorf("must set at least one 'send_target' or 'static_target' attribute")
	}

	// errList := ""
	// targets := d.Get("target").([]interface{})

	// for _, v := range targets {
	// 	target := v.(map[string]interface{})

	// 	if target["discovery_type"] == "static" {
	// 		name, ok := target["name"]

	// 		if name == "" || !ok {
	// 			errList += fmt.Sprintf(
	// 				"target with ip '%s' must set 'name' attribute when discovery type is 'static'\n",
	// 				target["ip"],
	// 			)
	// 		}
	// 	}

	// 	if c, ok := target["chap"]; ok {
	// 		chap := c.([]interface{})[0].(map[string]interface{})
	// 		if chap["method"].(string) == "bidirectional" {
	// 			if _, ok := chap["incoming_creds"]; !ok {
	// 				errList += fmt.Sprintf(
	// 					"target with ip '%s' must set 'incoming_creds' attribute as the 'method' attribute is set to 'bidirectional'\n",
	// 					target["ip"],
	// 				)
	// 			}
	// 		}
	// 	}
	// }

	// if errList != "" {
	// 	return fmt.Errorf(errList)
	// }

	return nil
}

// ----------------------------------------------------------------------------

// OLD

// func resourceVSphereIscsiTargetCreate(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*Client).vimClient
// 	hostID := d.Get("host_system_id").(string)

// 	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
// 	if err != nil {
// 		return err
// 	}

// 	targetList := d.Get("target").(*schema.Set).List()
// 	inherited := false

// 	for _, v := range targetList {
// 		target := v.(map[string]interface{})
// 		authSettings := &types.HostInternetScsiHbaAuthenticationProperties{
// 			ChapInherited:       &inherited,
// 			MutualChapInherited: &inherited,
// 		}

// 		if c, ok := target["chap"]; ok {
// 			chap := c.([]interface{})
// 			outgoingCreds := chap[0].(map[string]interface{})["outgoing_creds"].([]interface{})[0].(map[string]interface{})

// 			authSettings.ChapName = outgoingCreds["username"].(string)
// 			authSettings.ChapSecret = outgoingCreds["password"].(string)

// 			if incomingCredsList, ok := chap[0].(map[string]interface{})["incoming_creds"]; ok {
// 				incomingCreds := incomingCredsList.([]interface{})[0].(map[string]interface{})
// 				authSettings.MutualChapName = incomingCreds["username"].(string)
// 				authSettings.MutualChapSecret = incomingCreds["password"].(string)
// 			}
// 		}

// 		if d.Get("discovery_type") == "static" {
// 			targets := make([]types.HostInternetScsiHbaStaticTarget, 0, len(targetList))
// 			targets = append(targets, types.HostInternetScsiHbaStaticTarget{
// 				Address:                  target["ip"].(string),
// 				Port:                     target["port"].(int32),
// 				IScsiName:                target["name"].(string),
// 				AuthenticationProperties: authSettings,
// 			})

// 			if err = iscsi.AddInternetScsiStaticTargets(
// 				client,
// 				hostID,
// 				d.Get("adapter_id").(string),
// 				hssProps,
// 				targets,
// 			); err != nil {
// 				return err
// 			}
// 		} else {
// 			targets := make([]types.HostInternetScsiHbaSendTarget, 0, len(targetList))
// 			targets = append(targets, types.HostInternetScsiHbaSendTarget{
// 				Address:                  target["ip"].(string),
// 				Port:                     target["port"].(int32),
// 				AuthenticationProperties: authSettings,
// 			})

// 			if err = iscsi.AddInternetScsiSendTargets(
// 				client,
// 				hostID,
// 				d.Get("adapter_id").(string),
// 				hssProps,
// 				targets,
// 			); err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	// targetList := d.Get("target").([]interface{})
// 	// inherited := false

// 	// for _, v := range targetList {
// 	// 	target := v.(map[string]interface{})
// 	// 	authSettings := &types.HostInternetScsiHbaAuthenticationProperties{
// 	// 		ChapInherited:       &inherited,
// 	// 		MutualChapInherited: &inherited,
// 	// 	}

// 	// 	if c, ok := target["chap"]; ok {
// 	// 		chap := c.([]interface{})
// 	// 		outgoingCreds := chap[0].(map[string]interface{})["outgoing_creds"].([]interface{})[0].(map[string]interface{})

// 	// 		authSettings.ChapName = outgoingCreds["username"].(string)
// 	// 		authSettings.ChapSecret = outgoingCreds["password"].(string)

// 	// 		if incomingCredsList, ok := chap[0].(map[string]interface{})["incoming_creds"]; ok {
// 	// 			incomingCreds := incomingCredsList.([]interface{})[0].(map[string]interface{})
// 	// 			authSettings.MutualChapName = incomingCreds["username"].(string)
// 	// 			authSettings.MutualChapSecret = incomingCreds["password"].(string)
// 	// 		}
// 	// 	}

// 	// 	if d.Get("discovery_type") == "static" {
// 	// 		targets := make([]types.HostInternetScsiHbaStaticTarget, 0, len(targetList))
// 	// 		targets = append(targets, types.HostInternetScsiHbaStaticTarget{
// 	// 			Address:                  target["ip"].(string),
// 	// 			Port:                     target["port"].(int32),
// 	// 			IScsiName:                target["name"].(string),
// 	// 			AuthenticationProperties: authSettings,
// 	// 		})

// 	// 		if err = iscsi.AddInternetScsiStaticTargets(
// 	// 			client,
// 	// 			hostID,
// 	// 			d.Get("adapter_id").(string),
// 	// 			hssProps,
// 	// 			targets,
// 	// 		); err != nil {
// 	// 			return err
// 	// 		}
// 	// 	} else {
// 	// 		targets := make([]types.HostInternetScsiHbaSendTarget, 0, len(targetList))
// 	// 		targets = append(targets, types.HostInternetScsiHbaSendTarget{
// 	// 			Address:                  target["ip"].(string),
// 	// 			Port:                     target["port"].(int32),
// 	// 			AuthenticationProperties: authSettings,
// 	// 		})

// 	// 		if err = iscsi.AddInternetScsiSendTargets(
// 	// 			client,
// 	// 			hostID,
// 	// 			d.Get("adapter_id").(string),
// 	// 			hssProps,
// 	// 			targets,
// 	// 		); err != nil {
// 	// 			return err
// 	// 		}
// 	// 	}
// 	// }

// 	d.SetId(fmt.Sprintf("%s:%s", hostID, d.Get("adapter_id").(string)))

// 	return nil
// }

// func resourceVSphereIscsiTargetRead(d *schema.ResourceData, meta interface{}) error {
// 	client := meta.(*Client).vimClient
// 	hostID := d.Get("host_system_id").(string)
// 	adapterID := d.Get("adapter_id").(string)

// 	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
// 	if err != nil {
// 		return err
// 	}

// 	baseAdapter, err := iscsi.GetIscsiAdater(hssProps, hostID, adapterID)
// 	if err != nil {
// 		return err
// 	}

// 	adapter := baseAdapter.(*types.HostInternetScsiHba)
// 	targets := d.Get("target").(*schema.Set).List()

// 	return nil
// }

// func resourceVSphereIscsiTargetUpdate(d *schema.ResourceData, meta interface{}) error {
// 	//client := meta.(*Client).vimClient

// 	return nil
// }

// func resourceVSphereIscsiTargetDelete(d *schema.ResourceData, meta interface{}) error {

// 	return nil
// }

// func resourceVSphereIscsiTargetImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
// 	//client := meta.(*Client).vimClient

// 	return []*schema.ResourceData{d}, nil
// }

// func resourceVSphereIscsiTargetCustomDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
// 	errList := ""
// 	targets := d.Get("target").([]interface{})

// 	for _, v := range targets {
// 		target := v.(map[string]interface{})

// 		if target["discovery_type"] == "static" {
// 			name, ok := target["name"]

// 			if name == "" || !ok {
// 				errList += fmt.Sprintf(
// 					"target with ip '%s' must set 'name' attribute when discovery type is 'static'\n",
// 					target["ip"],
// 				)
// 			}
// 		}

// 		if c, ok := target["chap"]; ok {
// 			chap := c.([]interface{})[0].(map[string]interface{})
// 			if chap["method"].(string) == "bidirectional" {
// 				if _, ok := chap["incoming_creds"]; !ok {
// 					errList += fmt.Sprintf(
// 						"target with ip '%s' must set 'incoming_creds' attribute as the 'method' attribute is set to 'bidirectional'\n",
// 						target["ip"],
// 					)
// 				}
// 			}
// 		}
// 	}

// 	if errList != "" {
// 		return fmt.Errorf(errList)
// 	}

// 	return nil
// }
