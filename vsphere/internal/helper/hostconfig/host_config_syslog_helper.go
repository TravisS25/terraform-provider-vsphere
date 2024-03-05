package hostconfig

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	SyslogHostKey     = "Syslog.global.logHost"
	SyslogLogLevelKey = "Syslog.global.logLevel"
)

func HostConfigSyslogRead(d *schema.ResourceData, client *govmomi.Client, host *object.HostSystem) error {
	optManager, err := GetOptionManager(client, host)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	hostOpts, err := optManager.Query(ctx, SyslogHostKey)
	if err != nil {
		return fmt.Errorf("error querying for log host on host '%s': %s", host.Name(), err)
	}

	if len(hostOpts) > 0 {
		d.Set("log_host", hostOpts[0].GetOptionValue().Value)
	}

	logLvlOpts, err := optManager.Query(ctx, SyslogLogLevelKey)
	if err != nil {
		return fmt.Errorf("error querying for log level on host '%s': %s", host.Name(), err)
	}

	if len(logLvlOpts) > 0 {
		d.Set("log_level", logLvlOpts[0].GetOptionValue().Value)
	}

	return nil
}

func UpdateHostConfigSyslog(d *schema.ResourceData, client *govmomi.Client, host *object.HostSystem, isDelete bool) error {
	optManager, err := GetOptionManager(client, host)
	if err != nil {
		return err
	}

	// Default values
	logHost := ""
	logLvl := "info"

	if !isDelete {
		logHost = d.Get("log_host").(string)
		logLvl = d.Get("log_level").(string)
	}

	optValues := []*types.OptionValue{
		{
			Key:   SyslogHostKey,
			Value: logHost,
		},
		{
			Key:   SyslogLogLevelKey,
			Value: logLvl,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	for _, v := range optValues {
		if err = optManager.Update(
			ctx,
			[]types.BaseOptionValue{v},
		); err != nil {
			return fmt.Errorf("error trying to update syslog setting '%s' for host '%s': %s", v.Key, host.Name(), err)
		}
	}

	return nil
}
