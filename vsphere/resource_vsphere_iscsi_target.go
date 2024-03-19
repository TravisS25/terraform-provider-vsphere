package vsphere

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/iscsi"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "ID of the host system to attach iscsi adapter to",
				ExactlyOneOf: []string{"hostname"},
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Hostname of host system to attach iscsi adapter to",
			},
			"adapter_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Iscsi adapter the iscsi targets will be added to.  This should be in the form of 'vmhb<unique_name>'",
			},
			"dynamic_target": {
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
	adapterID := d.Get("adapter_id").(string)

	host, hr, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host for iscsi: %s", err)
	}

	hss, err := hostsystem.GetHostStorageSystemFromHost(client, host)
	if err != nil {
		return err
	}

	hssProps, err := hostsystem.HostStorageSystemProperties(hss)
	if err != nil {
		return err
	}

	dynamicTargets := d.Get("dynamic_target").(*schema.Set).List()
	hbaDynamicTargets := make([]types.HostInternetScsiHbaSendTarget, 0, len(dynamicTargets))

	for _, v := range dynamicTargets {
		dynamicTarget := v.(map[string]interface{})
		outgoingCreds := iscsi.ExtractChapCredsFromTarget(dynamicTarget, true)
		incomingCreds := iscsi.ExtractChapCredsFromTarget(dynamicTarget, false)
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

		hbaDynamicTargets = append(hbaDynamicTargets, types.HostInternetScsiHbaSendTarget{
			Address:                  dynamicTarget["ip"].(string),
			Port:                     int32(dynamicTarget["port"].(int)),
			AuthenticationProperties: authSettings,
		})
	}

	if len(hbaDynamicTargets) > 0 {
		if err = iscsi.AddInternetScsiDynamicTargets(
			client,
			host.Name(),
			adapterID,
			hssProps,
			hbaDynamicTargets,
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
			host.Name(),
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

	d.SetId(iscsi.GetIscsiTargetID(hr.Value, adapterID))

	return nil
}

func resourceVSphereIscsiTargetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, _, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host for iscsi read: %s", err)
	}

	return iscsiTargetRead(client, d, host, d.Get("adapter_id").(string), true)
}

func resourceVSphereIscsiTargetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	host, _, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host for iscsi create: %s", err)
	}

	adapterID := d.Get("adapter_id").(string)
	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, host)
	if err != nil {
		return err
	}

	if _, err = iscsi.GetIscsiAdater(hssProps, host.Name(), adapterID); err != nil {
		return err
	}

	if d.HasChange("dynamic_target") {
		oldVal, newVal := d.GetChange("dynamic_target")
		oldList := oldVal.(*schema.Set).List()
		newList := newVal.(*schema.Set).List()

		removeTargets, addTargets := structure.ExtractResourceDiff(oldList, newList)
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
			if err = iscsi.RemoveInternetScsiDynamicTargets(
				client,
				host.Name(),
				adapterID,
				hssProps,
				hbaRemoveTargets,
			); err != nil {
				return err
			}
		}

		if len(hbaAddTargets) > 0 {
			if err = iscsi.AddInternetScsiDynamicTargets(
				client,
				host.Name(),
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

		removeTargets, addTargets := structure.ExtractResourceDiff(oldList, newList)
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
				host.Name(),
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
				host.Name(),
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
	host, _, err := hostsystem.FromHostnameOrID(client, d)
	if err != nil {
		return fmt.Errorf("error retrieving host for iscsi create: %s", err)
	}

	adapterID := d.Get("adapter_id").(string)

	hss, err := hostsystem.GetHostStorageSystemFromHost(client, host)
	if err != nil {
		return err
	}

	hssProps, err := hostsystem.HostStorageSystemProperties(hss)
	if err != nil {
		return err
	}

	dynamicTargets := d.Get("dynamic_target").(*schema.Set).List()
	staticTargets := d.Get("static_target").(*schema.Set).List()

	removeDynamicTargets := make([]types.HostInternetScsiHbaSendTarget, 0, len(dynamicTargets))
	removeStaticTargets := make([]types.HostInternetScsiHbaStaticTarget, 0, len(staticTargets))

	for _, dynamicTarget := range dynamicTargets {
		target := dynamicTarget.(map[string]interface{})
		removeDynamicTargets = append(removeDynamicTargets, types.HostInternetScsiHbaSendTarget{
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

	if len(removeDynamicTargets) > 0 {
		if err = iscsi.RemoveInternetScsiDynamicTargets(
			client,
			host.Name(),
			adapterID,
			hssProps,
			removeDynamicTargets,
		); err != nil {
			return err
		}
	}

	if len(removeStaticTargets) > 0 {
		if err = iscsi.RemoveInternetScsiStaticTargets(
			client,
			host.Name(),
			adapterID,
			hssProps,
			removeStaticTargets,
		); err != nil {
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

	client := meta.(*Client).vimClient
	host, hr, err := hostsystem.CheckIfHostnameOrID(client, idSplit[0])
	if err != nil {
		return nil, fmt.Errorf("error retrieving host '%s' on import: %s", idSplit[0], err)
	}

	if err = iscsiTargetRead(client, d, host, idSplit[1], false); err != nil {
		return nil, fmt.Errorf("error reading iscsi target properties on import for host '%s': %s", host.Name(), err)
	}

	d.SetId(iscsi.GetIscsiTargetID(hr.Value, idSplit[1]))
	d.Set(hr.IDName, hr.Value)
	return []*schema.ResourceData{d}, nil
}

func resourceVSphereIscsiTargetCustomDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	var host *object.HostSystem
	var err error

	client := meta.(*Client).vimClient

	if d.Get("host_system_id") != "" {
		host, _, err = hostsystem.CheckIfHostnameOrID(client, d.Get("host_system_id").(string))
	} else {
		host, _, err = hostsystem.CheckIfHostnameOrID(client, d.Get("hostname").(string))
	}

	if err != nil {
		return fmt.Errorf("error retrieving host on custom diff: %s", err)
	}

	adapterID := d.Get("adapter_id").(string)

	if adapterID != "" {
		hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, host)
		if err != nil {
			return err
		}

		adapter, err := iscsi.GetIscsiAdater(hssProps, host.Name(), adapterID)
		if err != nil {
			return err
		}

		switch adapter.(type) {
		case *types.HostInternetScsiHba:
			break
		default:
			return fmt.Errorf("'adapter_id' belongs to a device that does NOT allow static or dynamic discovery")
		}
	}

	staticTargets, staticOk := d.GetOk("static_target")
	dynamicTargets, dynamicOk := d.GetOk("dynamic_target")

	if !staticOk && !dynamicOk {
		return fmt.Errorf("must set at least one 'dynamic_target' or 'static_target' attribute")
	}

	if staticOk {
		dupMap := map[string]struct{}{}
		strFmt := "%s:%s:%s"

		for _, v := range staticTargets.(*schema.Set).List() {
			st := v.(map[string]interface{})

			if _, ok := dupMap[fmt.Sprintf(strFmt, st["ip"], st["port"], st["name"])]; ok {
				return fmt.Errorf(
					"duplicate ip, port and name found for static target;  ip: %s, port: %d, name: %s",
					st["ip"],
					st["port"],
					st["name"],
				)
			} else {
				dupMap[fmt.Sprintf(strFmt, st["ip"], st["port"], st["name"])] = struct{}{}
			}
		}
	}

	if dynamicOk {
		dupMap := map[string]struct{}{}
		strFmt := "%s:%s"

		for _, v := range dynamicTargets.(*schema.Set).List() {
			st := v.(map[string]interface{})

			if _, ok := dupMap[fmt.Sprintf(strFmt, st["ip"], st["port"])]; ok {
				return fmt.Errorf(
					"duplicate ip and port found for dynamic target;  ip: %s, port: %d",
					st["ip"],
					st["port"],
				)
			} else {
				dupMap[fmt.Sprintf(strFmt, st["ip"], st["port"])] = struct{}{}
			}
		}
	}

	return nil
}

func iscsiTargetRead(client *govmomi.Client, d *schema.ResourceData, host *object.HostSystem, adapterID string, isRead bool) error {
	hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, host)
	if err != nil {
		return fmt.Errorf("error retrieving host system storage properties for host '%s': %s", host.Name(), err)
	}

	baseAdapter, err := iscsi.GetIscsiAdater(hssProps, host.Name(), adapterID)
	if err != nil {
		return fmt.Errorf("error retrieving base adapter for host '%s': %s", host.Name(), err)
	}

	adapter := baseAdapter.(*types.HostInternetScsiHba)
	dynamicTargets := make([]interface{}, 0, len(adapter.ConfiguredSendTarget))
	staticTargets := make([]interface{}, 0, len(adapter.ConfiguredStaticTarget))

	for _, dynamicTarget := range adapter.ConfiguredSendTarget {
		target := map[string]interface{}{
			"ip":   dynamicTarget.Address,
			"port": dynamicTarget.Port,
		}

		if isRead {
			currentDynamicTargets := d.Get("dynamic_target").(*schema.Set).List()
			for _, v := range currentDynamicTargets {
				currentDynamicTarget := v.(map[string]interface{})

				if currentDynamicTarget["ip"].(string) == target["ip"].(string) &&
					int32(currentDynamicTarget["port"].(int)) == target["port"].(int32) {
					target["chap"] = currentDynamicTarget["chap"]
				}
			}
		} else {
			target["chap"] = []interface{}{
				map[string]interface{}{
					"outgoing_creds": []interface{}{
						map[string]interface{}{
							"username": dynamicTarget.AuthenticationProperties.ChapName,
							"password": dynamicTarget.AuthenticationProperties.ChapSecret,
						},
					},
					"incoming_creds": []interface{}{
						map[string]interface{}{
							"username": dynamicTarget.AuthenticationProperties.MutualChapName,
							"password": dynamicTarget.AuthenticationProperties.MutualChapSecret,
						},
					},
				},
			}
		}

		dynamicTargets = append(dynamicTargets, target)
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

	d.Set("adapter_id", adapterID)
	d.Set("dynamic_target", dynamicTargets)
	d.Set("static_target", staticTargets)

	return nil
}
