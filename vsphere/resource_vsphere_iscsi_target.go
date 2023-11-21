package vsphere

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/iscsi"
	"github.com/vmware/govmomi/vim25/types"
)

func resourceVSphereIscsiTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereIscsiTargetCreate,
		Read:   resourceVSphereIscsiTargetRead,
		Delete: resourceVSphereIscsiTargetDelete,
		Update: resourceVSphereIscsiTargetUpdate,
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
				Type:     schema.TypeSet,
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
				Type:     schema.TypeSet,
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
		},
	}
}

func resourceVSphereIscsiTargetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	adapterID := d.Get("adapter_id").(string)

	hss, err := hostsystem.GetHostStorageSystemFromHost(client, hostID)
	if err != nil {
		return err
	}

	hssProps, err := hostsystem.HostStorageSystemProperties(hss)
	if err != nil {
		return err
	}

	sendTargets := d.Get("send_target").(*schema.Set).List()
	hbaSendTargets := make([]types.HostInternetScsiHbaSendTarget, 0, len(sendTargets))

	for _, v := range sendTargets {
		sendTarget := v.(map[string]interface{})
		outgoingCreds := iscsi.ExtractChapCredsFromTarget(sendTarget, true)
		incomingCreds := iscsi.ExtractChapCredsFromTarget(sendTarget, false)

		authSettings := &types.HostInternetScsiHbaAuthenticationProperties{}

		if len(outgoingCreds["username"].(string)) > 0 {
			authSettings.ChapAuthEnabled = true
			authSettings.ChapName = outgoingCreds["username"].(string)
			authSettings.ChapSecret = outgoingCreds["password"].(string)
			authSettings.ChapAuthenticationType = string(types.HostInternetScsiHbaChapAuthenticationTypeChapRequired)
		}

		if len(incomingCreds["username"].(string)) > 0 {
			authSettings.ChapAuthEnabled = true
			authSettings.MutualChapName = incomingCreds["username"].(string)
			authSettings.MutualChapSecret = incomingCreds["password"].(string)
			authSettings.MutualChapAuthenticationType = string(types.HostInternetScsiHbaChapAuthenticationTypeChapRequired)
		}

		hbaSendTargets = append(hbaSendTargets, types.HostInternetScsiHbaSendTarget{
			Address:                  sendTarget["ip"].(string),
			Port:                     int32(sendTarget["port"].(int)),
			AuthenticationProperties: authSettings,
		})
	}

	if len(hbaSendTargets) > 0 {
		if err = iscsi.AddInternetScsiSendTargets(
			client,
			hostID,
			adapterID,
			hssProps,
			hbaSendTargets,
		); err != nil {
			return err
		}
	}

	staticTargets := d.Get("static_target").(*schema.Set).List()
	hbaStaticTargets := make([]types.HostInternetScsiHbaStaticTarget, 0, len(staticTargets))

	for _, v := range staticTargets {
		staticTarget := v.(map[string]interface{})
		outgoingCreds := iscsi.ExtractChapCredsFromTarget(staticTarget, true)
		incomingCreds := iscsi.ExtractChapCredsFromTarget(staticTarget, false)
		authSettings := &types.HostInternetScsiHbaAuthenticationProperties{}

		if len(outgoingCreds["username"].(string)) > 0 {
			authSettings.ChapAuthEnabled = true
			authSettings.ChapName = outgoingCreds["username"].(string)
			authSettings.ChapSecret = outgoingCreds["password"].(string)
			authSettings.ChapAuthenticationType = string(types.HostInternetScsiHbaChapAuthenticationTypeChapRequired)
		}

		if len(incomingCreds["username"].(string)) > 0 {
			authSettings.ChapAuthEnabled = true
			authSettings.MutualChapName = incomingCreds["username"].(string)
			authSettings.MutualChapSecret = incomingCreds["password"].(string)
			authSettings.MutualChapAuthenticationType = string(types.HostInternetScsiHbaChapAuthenticationTypeChapRequired)
		}

		hbaStaticTargets = append(hbaStaticTargets, types.HostInternetScsiHbaStaticTarget{
			Address:                  staticTarget["ip"].(string),
			Port:                     int32(staticTarget["port"].(int)),
			IScsiName:                staticTarget["name"].(string),
			AuthenticationProperties: authSettings,
		})
	}

	if len(hbaStaticTargets) > 0 {
		if err = iscsi.AddInternetScsiStaticTargets(
			client,
			hostID,
			adapterID,
			hssProps,
			hbaStaticTargets,
		); err != nil {
			return err
		}
	}

	if err = iscsi.RescanStorageDevices(hss); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s:%s", hostID, adapterID))

	return nil
}

func resourceVSphereIscsiTargetRead(d *schema.ResourceData, meta interface{}) error {
	return iscsiTargetRead(
		d,
		meta,
		d.Get("host_system_id").(string),
		d.Get("adapter_id").(string),
		true,
	)
}

func resourceVSphereIscsiTargetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	adapterID := d.Get("adapter_id").(string)

	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
	if err != nil {
		return err
	}

	if _, err = iscsi.GetIscsiAdater(hssProps, hostID, adapterID); err != nil {
		return err
	}

	if d.HasChange("send_target") {
		oldVal, newVal := d.GetChange("send_target")
		oldList := oldVal.(*schema.Set).List()
		newList := newVal.(*schema.Set).List()

		removeTargets, addTargets := iscsi.ExtractTargetUpdates(oldList, newList)
		hbaRemoveTargets := make([]types.HostInternetScsiHbaSendTarget, 0, len(removeTargets))

		for _, v := range removeTargets {
			removeTarget := v.(map[string]interface{})
			hbaRemoveTargets = append(hbaRemoveTargets, types.HostInternetScsiHbaSendTarget{
				Address: removeTarget["ip"].(string),
				Port:    int32(removeTarget["port"].(int)),
			})
		}

		hbaAddTargets := make([]types.HostInternetScsiHbaSendTarget, 0, len(addTargets))

		for _, v := range addTargets {
			addTarget := v.(map[string]interface{})
			outgoingCreds := iscsi.ExtractChapCredsFromTarget(addTarget, true)
			incomingCreds := iscsi.ExtractChapCredsFromTarget(addTarget, false)
			authSettings := &types.HostInternetScsiHbaAuthenticationProperties{}

			if len(outgoingCreds["username"].(string)) > 0 {
				authSettings.ChapAuthEnabled = true
				authSettings.ChapName = outgoingCreds["username"].(string)
				authSettings.ChapSecret = outgoingCreds["password"].(string)
				authSettings.ChapAuthenticationType = string(types.HostInternetScsiHbaChapAuthenticationTypeChapRequired)
			}

			if len(incomingCreds["username"].(string)) > 0 {
				authSettings.ChapAuthEnabled = true
				authSettings.MutualChapName = incomingCreds["username"].(string)
				authSettings.MutualChapSecret = incomingCreds["password"].(string)
				authSettings.MutualChapAuthenticationType = string(types.HostInternetScsiHbaChapAuthenticationTypeChapRequired)
			}

			hbaAddTargets = append(hbaAddTargets, types.HostInternetScsiHbaSendTarget{
				Address:                  addTarget["ip"].(string),
				Port:                     int32(addTarget["port"].(int)),
				AuthenticationProperties: authSettings,
			})
		}

		if len(hbaRemoveTargets) > 0 {
			if err = iscsi.RemoveInternetScsiSendTargets(
				client,
				hostID,
				adapterID,
				hssProps,
				hbaRemoveTargets,
			); err != nil {
				return err
			}
		}

		if len(hbaAddTargets) > 0 {
			if err = iscsi.AddInternetScsiSendTargets(
				client,
				hostID,
				adapterID,
				hssProps,
				hbaAddTargets,
			); err != nil {
				return err
			}
		}
	}

	if d.HasChange("static_target") {
		oldVal, newVal := d.GetChange("static_target")
		oldList := oldVal.(*schema.Set).List()
		newList := newVal.(*schema.Set).List()

		removeTargets, addTargets := iscsi.ExtractTargetUpdates(oldList, newList)
		hbaRemoveTargets := make([]types.HostInternetScsiHbaStaticTarget, 0, len(removeTargets))

		for _, v := range removeTargets {
			removeTarget := v.(map[string]interface{})
			hbaRemoveTargets = append(hbaRemoveTargets, types.HostInternetScsiHbaStaticTarget{
				Address:   removeTarget["ip"].(string),
				Port:      int32(removeTarget["port"].(int)),
				IScsiName: removeTarget["name"].(string),
			})
		}

		hbaAddTargets := make([]types.HostInternetScsiHbaStaticTarget, 0, len(addTargets))

		for _, v := range addTargets {
			addTarget := v.(map[string]interface{})
			outgoingCreds := iscsi.ExtractChapCredsFromTarget(addTarget, true)
			incomingCreds := iscsi.ExtractChapCredsFromTarget(addTarget, false)
			authSettings := &types.HostInternetScsiHbaAuthenticationProperties{}

			if len(outgoingCreds["username"].(string)) > 0 {
				authSettings.ChapAuthEnabled = true
				authSettings.ChapName = outgoingCreds["username"].(string)
				authSettings.ChapSecret = outgoingCreds["password"].(string)
				authSettings.ChapAuthenticationType = string(types.HostInternetScsiHbaChapAuthenticationTypeChapRequired)
			}

			if len(incomingCreds["username"].(string)) > 0 {
				authSettings.ChapAuthEnabled = true
				authSettings.MutualChapName = incomingCreds["username"].(string)
				authSettings.MutualChapSecret = incomingCreds["password"].(string)
				authSettings.MutualChapAuthenticationType = string(types.HostInternetScsiHbaChapAuthenticationTypeChapRequired)
			}

			hbaAddTargets = append(hbaAddTargets, types.HostInternetScsiHbaStaticTarget{
				Address:                  addTarget["ip"].(string),
				Port:                     int32(addTarget["port"].(int)),
				IScsiName:                addTarget["name"].(string),
				AuthenticationProperties: authSettings,
			})
		}

		if len(hbaRemoveTargets) > 0 {
			if err = iscsi.RemoveInternetScsiStaticTargets(
				client,
				hostID,
				adapterID,
				hssProps,
				hbaRemoveTargets,
			); err != nil {
				return err
			}
		}

		if len(hbaAddTargets) > 0 {
			if err = iscsi.AddInternetScsiStaticTargets(
				client,
				hostID,
				adapterID,
				hssProps,
				hbaAddTargets,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceVSphereIscsiTargetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Get("host_system_id").(string)
	adapterID := d.Get("adapter_id").(string)

	hss, err := hostsystem.GetHostStorageSystemFromHost(client, hostID)
	if err != nil {
		return err
	}

	hssProps, err := hostsystem.HostStorageSystemProperties(hss)
	if err != nil {
		return err
	}

	sendTargets := d.Get("send_target").(*schema.Set).List()
	staticTargets := d.Get("static_target").(*schema.Set).List()

	removeSendTargets := make([]types.HostInternetScsiHbaSendTarget, 0, len(sendTargets))
	removeStaticTargets := make([]types.HostInternetScsiHbaStaticTarget, 0, len(staticTargets))

	for _, sendTarget := range sendTargets {
		target := sendTarget.(map[string]interface{})
		removeSendTargets = append(removeSendTargets, types.HostInternetScsiHbaSendTarget{
			Address: target["ip"].(string),
			Port:    int32(target["port"].(int)),
		})
	}

	for _, staticTarget := range staticTargets {
		target := staticTarget.(map[string]interface{})
		removeStaticTargets = append(removeStaticTargets, types.HostInternetScsiHbaStaticTarget{
			Address:   target["ip"].(string),
			Port:      int32(target["port"].(int)),
			IScsiName: target["name"].(string),
		})
	}

	if len(removeSendTargets) > 0 {
		if err = iscsi.RemoveInternetScsiSendTargets(client, hostID, adapterID, hssProps, removeSendTargets); err != nil {
			return err
		}
	}

	if len(removeStaticTargets) > 0 {
		if err = iscsi.RemoveInternetScsiStaticTargets(client, hostID, adapterID, hssProps, removeStaticTargets); err != nil {
			return err
		}
	}

	if err = iscsi.RescanStorageDevices(hss); err != nil {
		return err
	}

	return nil
}

func resourceVSphereIscsiTargetImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idSplit := strings.Split(d.Id(), ":")

	if len(idSplit) != 2 {
		return nil, fmt.Errorf("invalid import format; should be '<host_system_id>:<adapter_id>'")
	}

	hostID := idSplit[0]
	adapterID := idSplit[1]

	err := iscsiTargetRead(d, meta, hostID, adapterID, false)
	if err != nil {
		return nil, err
	}

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
		return fmt.Errorf("'adapter_id' belongs to a device that does NOT allow static or dynamic discovery")
	}

	staticTargets, staticOk := d.GetOk("static_target")
	sendTargets, sendOK := d.GetOk("send_target")

	if !staticOk && !sendOK {
		return fmt.Errorf("must set at least one 'send_target' or 'static_target' attribute")
	}

	if staticOk {
		dupMap := map[string]bool{}
		strFmt := "%s:%s:%s"

		for _, v := range staticTargets.(*schema.Set).List() {
			st := v.(map[string]interface{})

			log.Printf("static foobar: %+v\n", st)

			if _, ok := dupMap[fmt.Sprintf(strFmt, st["ip"], st["port"], st["name"])]; ok {
				return fmt.Errorf(
					"duplicate ip, port and name found for static target;  ip: %s, port: %d, name: %s",
					st["ip"],
					st["port"],
					st["name"],
				)
			} else {
				dupMap[fmt.Sprintf(strFmt, st["ip"], st["port"], st["name"])] = true
			}
		}
	}

	if sendOK {
		dupMap := map[string]bool{}
		strFmt := "%s:%s"

		for _, v := range sendTargets.(*schema.Set).List() {
			st := v.(map[string]interface{})
			log.Printf("send foobar: %+v\n", st)

			if _, ok := dupMap[fmt.Sprintf(strFmt, st["ip"], st["port"])]; ok {
				return fmt.Errorf(
					"duplicate ip and port found for send target;  ip: %s, port: %d",
					st["ip"],
					st["port"],
				)
			} else {
				dupMap[fmt.Sprintf(strFmt, st["ip"], st["port"])] = true
			}
		}
	}

	return nil
}

func iscsiTargetRead(d *schema.ResourceData, meta interface{}, hostID, adapterID string, isRead bool) error {
	client := meta.(*Client).vimClient

	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
	if err != nil {
		return err
	}

	baseAdapter, err := iscsi.GetIscsiAdater(hssProps, hostID, adapterID)
	if err != nil {
		return err
	}

	adapter := baseAdapter.(*types.HostInternetScsiHba)
	sendTargets := make([]interface{}, 0, len(adapter.ConfiguredSendTarget))
	staticTargets := make([]interface{}, 0, len(adapter.ConfiguredStaticTarget))

	for _, sendTarget := range adapter.ConfiguredSendTarget {
		target := map[string]interface{}{
			"ip":   sendTarget.Address,
			"port": sendTarget.Port,
		}

		if isRead {
			currentSendTargets := d.Get("send_target").(*schema.Set).List()
			for _, v := range currentSendTargets {
				currentSendTarget := v.(map[string]interface{})

				if currentSendTarget["ip"].(string) == target["ip"].(string) &&
					int32(currentSendTarget["port"].(int)) == target["port"].(int32) {
					target["chap"] = currentSendTarget["chap"]
				}
			}
		} else {
			target["chap"] = []interface{}{
				map[string]interface{}{
					"outgoing_creds": []interface{}{
						map[string]interface{}{
							"username": sendTarget.AuthenticationProperties.ChapName,
							"password": sendTarget.AuthenticationProperties.ChapSecret,
						},
					},
					"incoming_creds": []interface{}{
						map[string]interface{}{
							"username": sendTarget.AuthenticationProperties.MutualChapName,
							"password": sendTarget.AuthenticationProperties.MutualChapSecret,
						},
					},
				},
			}
		}

		sendTargets = append(sendTargets, target)
	}

	for _, staticTarget := range adapter.ConfiguredStaticTarget {
		target := map[string]interface{}{
			"ip":   staticTarget.Address,
			"port": staticTarget.Port,
			"name": staticTarget.IScsiName,
		}

		if isRead {
			currentStaticTargets := d.Get("static_target").(*schema.Set).List()

			for _, v := range currentStaticTargets {
				currentStaticTarget := v.(map[string]interface{})

				if currentStaticTarget["ip"].(string) == target["ip"].(string) &&
					int32(currentStaticTarget["port"].(int)) == target["port"].(int32) &&
					currentStaticTarget["name"].(string) == target["name"].(string) {
					target["chap"] = currentStaticTarget["chap"]
				}
			}
		} else {
			target["chap"] = []interface{}{
				map[string]interface{}{
					"outgoing_creds": []interface{}{
						map[string]interface{}{
							"username": staticTarget.AuthenticationProperties.ChapName,
							"password": staticTarget.AuthenticationProperties.ChapSecret,
						},
					},
					"incoming_creds": []interface{}{
						map[string]interface{}{
							"username": staticTarget.AuthenticationProperties.MutualChapName,
							"password": staticTarget.AuthenticationProperties.MutualChapSecret,
						},
					},
				},
			}
		}

		staticTargets = append(staticTargets, target)
	}

	d.Set("host_system_id", hostID)
	d.Set("adapter_id", adapterID)
	d.Set("send_target", sendTargets)
	d.Set("static_target", staticTargets)

	return nil
}
