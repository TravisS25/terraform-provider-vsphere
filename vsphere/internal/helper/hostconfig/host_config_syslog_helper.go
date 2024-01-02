package hostconfig

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	SyslogHostKey     = "Syslog.global.logHost"
	SyslogLogLevelKey = "Syslog.global.logLevel"
)

func HostConfigSyslogRead(ctx context.Context, d *schema.ResourceData, client *govmomi.Client, hostID string) error {
	optManager, err := GetOptionManager(ctx, client, hostID)
	if err != nil {
		return err
	}

	hostOpts, err := optManager.Query(ctx, SyslogHostKey)
	if err != nil {
		return fmt.Errorf("error querying for log host on host '%s': %s", hostID, err)
	}

	if len(hostOpts) > 0 {
		d.Set("log_host", hostOpts[0].GetOptionValue().Value)
	}

	logLvlOpts, err := optManager.Query(ctx, SyslogLogLevelKey)
	if err != nil {
		return fmt.Errorf("error querying for log level on host '%s': %s", hostID, err)
	}

	if len(logLvlOpts) > 0 {
		d.Set("log_level", logLvlOpts[0].GetOptionValue().Value)
	}

	return nil
}

func UpdateHostConfigSyslog(ctx context.Context, d *schema.ResourceData, client *govmomi.Client, hostID string, isDelete bool) error {
	optManager, err := GetOptionManager(ctx, client, hostID)
	if err != nil {
		return err
	}

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

	for _, v := range optValues {
		if err = optManager.Update(
			ctx,
			[]types.BaseOptionValue{v},
		); err != nil {
			return fmt.Errorf("error trying to update syslog options for host '%s': %s", hostID, err)
		}
	}

	return nil
}
