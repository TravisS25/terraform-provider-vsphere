package hostservicestate

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/types"
)

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

var (
	ServiceKeyList = []string{
		string(HostServiceKeyDCUI),
		string(HostServiceKeyShell),
		string(HostServiceKeySSH),
		string(HostServiceKeyAttestd),
		string(HostServiceKeyDPD),
		string(HostServiceKeyKMXD),
		string(HostServiceKeyLoadBasedTeaming),
		string(HostServiceKeyActiveDirectory),
		string(HostServiceKeyNTPD),
		string(HostServiceKeySmartCard),
		string(HostServiceKeyPTPD),
		string(HostServiceKeyCIMServer),
		string(HostServiceKeySLPD),
		string(HostServiceKeySNMPD),
		string(HostServiceKeyVLTD),
		string(HostServiceKeySyslogServer),
		string(HostServiceKeyHAAgent),
		string(HostServiceKeyVcenterAgent),
		string(HostServiceKeyXORG),
	}
)

// GetServiceState retrieves the service state of the given host
func GetServiceState(client *govmomi.Client, hostID string, key HostServiceKey, timeout time.Duration) (map[string]interface{}, error) {
	hsList, err := GetHostServies(client, hostID, timeout)
	if err != nil {
		return nil, err
	}

	for _, hostSrv := range hsList {
		if hostSrv.Key == string(key) {
			return map[string]interface{}{
				"key":    hostSrv.Key,
				"policy": hostSrv.Policy,
			}, nil
		}
	}

	return nil, fmt.Errorf("could not find service with key '%s' on host '%s'", key, hostID)
}

// GetHostServies retrieves all of the services for a given host
func GetHostServies(client *govmomi.Client, hostID string, timeout time.Duration) ([]types.HostService, error) {
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
			return nil, fmt.Errorf("error while trying to obtain host service system for host '%s': %s", host.Name(), err)
		}

		log.Printf("[INFO] querying services for host '%s'", hostID)

		hsList, err := hss.Service(ctx)
		if err != nil {
			return nil, fmt.Errorf("error while trying to obtain list of host services for host '%s': %s", host.Name(), err)
		}

		return hsList, nil
	}

	return nil, fmt.Errorf("could not obtain config manager for host %s", host.Name())
}

// SetServiceState sets the state of a given service
func SetServiceState(client *govmomi.Client, hostID string, ss map[string]interface{}, timeout time.Duration, running bool) error {
	// Find host and get reference to it.
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return fmt.Errorf("error while trying to retrieve host '%s': %s", hostID, err)
	}

	key := ss["key"].(string)
	policy := ss["policy"].(string)

	// Checks to make sure key and policy are not empty
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

		// Start service if running is set, else stop service
		if running {
			log.Printf("[INFO] starting '%s' service for host '%s'", key, hostID)

			if err = hss.Start(ctx, key); err != nil {
				return fmt.Errorf("error while trying to start %s service for host %s: %s", key, host.Name(), err)
			}
		} else {
			log.Printf("[INFO] stopping '%s' service for host '%s'", key, hostID)

			if err = hss.Stop(ctx, key); err != nil {
				return fmt.Errorf("error while trying to stop %s service for host %s: %s", key, host.Name(), err)
			}
		}

		log.Printf("[INFO] updating service '%s' with policy '%s' for host '%s'", key, policy, hostID)

		if err = hss.UpdatePolicy(ctx, key, policy); err != nil {
			return fmt.Errorf("error while trying to update policy for %s service for host '%s': %s", key, host.Name(), err)
		}

		return nil
	}

	return fmt.Errorf("could not obtain config manager for host '%s' to set state for service '%s'", host.Name(), key)
}

func GetServiceKeyMsg() string {
	srvKeyOptMsg := "srvKeyOptMsg"
	for _, key := range ServiceKeyList {
		srvKeyOptMsg += "'" + key + "', "
	}

	return srvKeyOptMsg
}
