package hostconfig

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	SyslogHostKey = "Syslog.global.logHost"
)

func HostConfigSyslogRead(ctx context.Context, d *schema.ResourceData, client *govmomi.Client, hostID string) error {
	optManager, err := GetOptionManager(ctx, d, client, hostID)
	if err != nil {
		return err
	}

	queryOpts, err := optManager.Query(ctx, SyslogHostKey)
	if err != nil {
		return fmt.Errorf("error querying for log host on host '%s': %s", hostID, err)
	}

	if len(queryOpts) > 0 {
		d.Set("log_host", queryOpts[0].GetOptionValue().Value)
	}

	return nil
}

func UpdateHostConfigSyslog(ctx context.Context, d *schema.ResourceData, client *govmomi.Client, hostID string, isDelete bool) error {
	optManager, err := GetOptionManager(ctx, d, client, hostID)
	if err != nil {
		return err
	}

	var logHost *string

	if !isDelete {
		lh := d.Get("log_host").(string)
		logHost = &lh
	}

	if err = optManager.Update(
		ctx,
		[]types.BaseOptionValue{
			&types.OptionValue{
				Key:   SyslogHostKey,
				Value: logHost,
			},
		},
	); err != nil {
		return fmt.Errorf("error trying to update syslog options for host '%s': %s", hostID, err)
	}

	return nil
}
