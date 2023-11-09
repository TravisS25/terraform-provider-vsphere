package hostservicestate

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/types"
)

// ServiceState represents the state of a given service for a given esxi host
// type ServiceState struct {
// 	HostSystemID string                  `json:"hostSystemID"`
// 	Key          HostServiceKey          `json:"key"`
// 	Policy       types.HostServicePolicy `json:"policy"`
// 	Running      bool                    `json:"running"`
// }

type ServiceState struct {
	HostSystemID string                  `json:"hostSystemID"`
	Key          HostServiceKey          `json:"key"`
	Policy       types.HostServicePolicy `json:"policy"`
	Running      bool                    `json:"running"`
}

// HostServiceKey represents the key value of a service for esxi host
type HostServiceKey string

const (
	HostServiceKeyDCUI             HostServiceKey = "DCUI"
	HostServiceKeyShell            HostServiceKey = "TSM"
	HostServiceKeySSH              HostServiceKey = "TSM-SSH"
	HostServiceKeyAttestd          HostServiceKey = "attestd"
	HostServiceKeyDPD              HostServiceKey = "dpd"
	HostServiceKeyKMXD             HostServiceKey = "kmxd"
	HostServiceKeyLoadBasedTeaming HostServiceKey = "lbtd"
	HostServiceKeyActiveDirectory  HostServiceKey = "lwsmd"
	HostServiceKeyNTPD             HostServiceKey = "ntpd"
	HostServiceKeySmartCard        HostServiceKey = "pcscd"
	HostServiceKeyPTPD             HostServiceKey = "ptpd"
	HostServiceKeyCIMServer        HostServiceKey = "sfcbd-watchdog"
	HostServiceKeySLPD             HostServiceKey = "slpd"
	HostServiceKeySNMPD            HostServiceKey = "snmpd"
	HostServiceKeyVLTD             HostServiceKey = "vltd"
	HostServiceKeySyslogServer     HostServiceKey = "vmsyslogd"
	HostServiceKeyHAAgent          HostServiceKey = "vmware-fdm"
	HostServiceKeyVcenterAgent     HostServiceKey = "vpxa"
	HostServiceKeyXORG             HostServiceKey = "xorg"
)

// // GetServiceState retrieves the service state of the given host
// func GetServiceState(client *govmomi.Client, hostID, key string, timeout time.Duration) (*ServiceState, error) {
// 	//hostID := d.Get("host_system_id").(string)

// 	// Find host and get reference to it.
// 	host, err := hostsystem.FromID(client, hostID)
// 	if err != nil {
// 		return nil, fmt.Errorf("error while trying to retrieve host '%s': %s", hostID, err)
// 	}

// 	if host.ConfigManager() != nil {
// 		ctx, cancel := context.WithTimeout(context.Background(), timeout)
// 		defer cancel()

// 		hss, err := host.ConfigManager().ServiceSystem(ctx)
// 		if err != nil {
// 			return nil, fmt.Errorf("error while trying to obtain host service system for host %s: %s", host.Name(), err)
// 		}

// 		hsList, err := hss.Service(ctx)
// 		if err != nil {
// 			return nil, fmt.Errorf("error while trying to obtain list of host services for host %s: %s", host.Name(), err)
// 		}

// 		for _, hostSrv := range hsList {
// 			if strings.EqualFold(hostSrv.Key, key) {
// 				return &ServiceState{
// 					HostSystemID: hostID,
// 					Key:          HostServiceKey(hostSrv.Key),
// 					Policy:       types.HostServicePolicy(hostSrv.Policy),
// 					Running:      hostSrv.Running,
// 				}, nil
// 			}
// 		}

// 		return nil, fmt.Errorf("could not find service with key: %s", key)
// 	}

// 	return nil, fmt.Errorf("could not obtain config manager for host %s", host.Name())
// }

// // SetServiceState sets the state of a given service
// func SetServiceState(client *govmomi.Client, ss ServiceState, timeout time.Duration) error {
// 	// Find host and get reference to it.
// 	host, err := hostsystem.FromID(client, ss.HostSystemID)
// 	if err != nil {
// 		return fmt.Errorf("error while trying to retrieve host '%s': %s", ss.HostSystemID, err)
// 	}

// 	if ss.Key == "" {
// 		return fmt.Errorf("service key must be set for host: '%s'", host.Name())
// 	}
// 	if ss.Policy == "" {
// 		return fmt.Errorf("service policy must be set for host: '%s'", host.Name())
// 	}

// 	if host.ConfigManager() != nil {
// 		ctx, cancel := context.WithTimeout(context.Background(), timeout)
// 		defer cancel()

// 		hss, err := host.ConfigManager().ServiceSystem(ctx)
// 		if err != nil {
// 			return fmt.Errorf("error while trying to obtain host service system for host %s: %s", host.Name(), err)
// 		}

// 		if ss.Running {
// 			if err = hss.Start(ctx, string(ss.Key)); err != nil {
// 				return fmt.Errorf("error while trying to start %s service for host %s: %s", ss.Key, host.Name(), err)
// 			}
// 		} else {
// 			if err = hss.Stop(ctx, string(ss.Key)); err != nil {
// 				return fmt.Errorf("error while trying to stop %s service for host %s.  Error: %s", ss.Key, host.Name(), err)
// 			}
// 		}

// 		if err = hss.UpdatePolicy(ctx, string(ss.Key), string(ss.Policy)); err != nil {
// 			return fmt.Errorf("error while trying to update policy for %s service for host %s.  Error: %s", ss.Key, host.Name(), err)
// 		}

// 		return nil
// 	}

// 	return fmt.Errorf("could not obtain config manager for host %s to set state for service %s", host.Name(), ss.Key)
// }

// func UpdateServiceState(d *schema.ResourceData, client *govmomi.Client, isCreate bool) error {
// 	hostID := d.Get("host_system_id").(string)
// 	key := HostServiceKey(d.Get("key").(string))
// 	policy := types.HostServicePolicy(d.Get("policy").(string))

// 	if isCreate {
// 		d.SetId(fmt.Sprintf("%s:%s", hostID, key))
// 		d.Set("host_system_id", hostID)
// 	}

// 	err := SetServiceState(
// 		client,
// 		ServiceState{
// 			HostSystemID: hostID,
// 			Key:          key,
// 			Policy:       policy,
// 			Running:      d.Get("running").(bool),
// 		},
// 		provider.DefaultAPITimeout,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("error trying to set service state for host '%s': %s", hostID, err)
// 	}

// 	return nil
// }

// GetServiceState retrieves the service state of the given host
func GetServiceState(client *govmomi.Client, hostID string, key HostServiceKey, timeout time.Duration) (map[string]interface{}, error) {
	// Find host and get reference to it.
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return nil, fmt.Errorf("error while trying to retrieve host '%s': %s", hostID, err)
	}

	if host.ConfigManager() != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		hss, err := host.ConfigManager().ServiceSystem(ctx)
		if err != nil {
			return nil, fmt.Errorf("error while trying to obtain host service system for host %s: %s", host.Name(), err)
		}

		hsList, err := hss.Service(ctx)
		if err != nil {
			return nil, fmt.Errorf("error while trying to obtain list of host services for host %s: %s", host.Name(), err)
		}

		for _, hostSrv := range hsList {
			if strings.EqualFold(hostSrv.Key, string(key)) {
				return map[string]interface{}{
					"key":     hostSrv.Key,
					"policy":  hostSrv.Policy,
					"running": hostSrv.Running,
				}, nil
			}
		}

		return nil, fmt.Errorf("could not find service with key: '%s' on host '%s'", key, hostID)
	}

	return nil, fmt.Errorf("could not obtain config manager for host %s", host.Name())
}

// SetServiceState sets the state of a given service
func SetServiceState(client *govmomi.Client, hostID string, ss map[string]interface{}, timeout time.Duration) error {
	// Find host and get reference to it.
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return fmt.Errorf("error while trying to retrieve host '%s': %s", hostID, err)
	}

	key := ss["key"].(string)
	policy := ss["policy"].(string)
	running := ss["running"].(bool)

	if key == "" {
		return fmt.Errorf("service key must be set for host: '%s'", hostID)
	}
	if policy == "" {
		return fmt.Errorf("service policy must be set for host: '%s'", hostID)
	}

	if host.ConfigManager() != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		hss, err := host.ConfigManager().ServiceSystem(ctx)
		if err != nil {
			return fmt.Errorf("error while trying to obtain host service system for host %s: %s", host.Name(), err)
		}

		if running {
			if err = hss.Start(ctx, key); err != nil {
				return fmt.Errorf("error while trying to start %s service for host %s: %s", key, host.Name(), err)
			}
		} else {
			if err = hss.Stop(ctx, key); err != nil {
				return fmt.Errorf("error while trying to stop %s service for host %s.  Error: %s", key, host.Name(), err)
			}
		}

		if err = hss.UpdatePolicy(ctx, key, policy); err != nil {
			return fmt.Errorf("error while trying to update policy for %s service for host %s.  Error: %s", key, host.Name(), err)
		}

		return nil
	}

	return fmt.Errorf("could not obtain config manager for host %s to set state for service %s", host.Name(), key)
}

func UpdateServiceState(d *schema.ResourceData, client *govmomi.Client, isCreate bool) error {
	var err error

	hostID := d.Get("host_system_id").(string)

	if isCreate {
		d.SetId(hostID)
		d.Set("host_system_id", hostID)
		srvs := d.Get("service").([]interface{})

		for _, v := range srvs {
			srv := v.(map[string]interface{})

			if err = SetServiceState(
				client,
				hostID,
				srv,
				provider.DefaultAPITimeout,
			); err != nil {
				return fmt.Errorf("error trying to set service state for host '%s' with key '%s': %s", hostID, srv["key"], err)
			}
		}

	} else {
		oldVal, newVal := d.GetChange("service")
		oldList := oldVal.([]interface{})
		newList := newVal.([]interface{})

		log.Printf("old values: %+v", oldList)
		log.Printf("new values: %+v", newList)

		for _, v := range oldList {
			oldSrv := v.(map[string]interface{})
			found := false

			for _, t := range newList {
				newSrv := t.(map[string]interface{})

				if newSrv["key"].(string) == oldSrv["key"].(string) {
					found = true
					break
				}
			}

			if !found {
				oldSrv["running"] = false
				oldSrv["policy"] = string(types.HostServicePolicyOff)

				if err = SetServiceState(
					client,
					hostID,
					oldSrv,
					provider.DefaultAPITimeout,
				); err != nil {
					return fmt.Errorf("error trying to set service state for host '%s' with key '%s': %s", hostID, oldSrv["key"], err)
				}
			}
		}

		for _, v := range newList {
			newSrv := v.(map[string]interface{})

			if err = SetServiceState(
				client,
				hostID,
				newSrv,
				provider.DefaultAPITimeout,
			); err != nil {
				return fmt.Errorf("error trying to set service state for host '%s' with key '%s': %s", hostID, newSrv["key"], err)
			}
		}
	}

	return nil
}

func GetServiceKeyMsg(keyList []string) string {
	srvKeyOptMsg := "srvKeyOptMsg"
	for _, key := range keyList {
		srvKeyOptMsg += "'" + key + "', "
	}

	return srvKeyOptMsg
}
