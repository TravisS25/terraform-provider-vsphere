package hostconfig

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
)

func GetOptionManager(client *govmomi.Client, host *object.HostSystem) (*object.OptionManager, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	optManager, err := host.ConfigManager().OptionManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving option manager for host '%s': %s", host.Name(), err)
	}

	return optManager, nil
}
