package hostconfig

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
)

func GetOptionManager(ctx context.Context, client *govmomi.Client, hostID string) (*object.OptionManager, error) {
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return nil, err
	}

	optManager, err := host.ConfigManager().OptionManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving option manager for host '%s': %s", hostID, err)
	}

	return optManager, nil
}
