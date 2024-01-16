package hostservicestate

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
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
	HostServiceKeyVCenterAgent     HostServiceKey = "vpxa"
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
		string(HostServiceKeyXORG),
	}

	// These are keys that terraform should NEVER manage and are keys that will both
	// NOT be an option for "service" attribute in "vsphere_host_service_state" resource
	// and will be excluded when querying the data source "vsphere_host_service_state"
	ExcludeServiceKeyList = []string{
		string(HostServiceKeySyslogServer),
		string(HostServiceKeyHAAgent),
		string(HostServiceKeyVCenterAgent),
	}
)

// GetServiceState retrieves the service state of the given host
func GetServiceState(client *govmomi.Client, host *object.HostSystem, key HostServiceKey, timeout time.Duration) (map[string]interface{}, error) {
	hsList, err := GetHostServies(client, host, timeout)
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

	return nil, fmt.Errorf("could not find service with key '%s' on host '%s'", key, host.Name())
}

// GetHostServies retrieves all of the services for a given host
func GetHostServies(client *govmomi.Client, host *object.HostSystem, timeout time.Duration) ([]types.HostService, error) {
	if host.ConfigManager() != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		hss, err := host.ConfigManager().ServiceSystem(ctx)
		if err != nil {
			return nil, fmt.Errorf("error while trying to obtain host service system for host '%s': %s", host.Name(), err)
		}

		log.Printf("[INFO] querying services for host '%s'", host.Name())

		hsList, err := hss.Service(ctx)
		if err != nil {
			return nil, fmt.Errorf("error while trying to obtain list of host services for host '%s': %s", host.Name(), err)
		}

		return hsList, nil
	}

	return nil, fmt.Errorf("could not obtain config manager for host %s", host.Name())
}

// SetServiceState sets the state of a given service
func SetServiceState(client *govmomi.Client, host *object.HostSystem, ss map[string]interface{}, timeout time.Duration, running bool) error {
	key := ss["key"].(string)
	policy := ss["policy"].(string)

	// Checks to make sure key and policy are not empty
	if key == "" {
		return fmt.Errorf("service key must be set for host: '%s'", host.Name())
	}
	if policy == "" {
		return fmt.Errorf("service policy must be set for host: '%s'", host.Name())
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
			log.Printf("[INFO] starting '%s' service for host '%s'", key, host.Name())

			if err = hss.Start(ctx, key); err != nil {
				return fmt.Errorf("error while trying to start %s service for host %s: %s", key, host.Name(), err)
			}
		} else {
			log.Printf("[INFO] stopping '%s' service for host '%s'", key, host.Name())

			if err = hss.Stop(ctx, key); err != nil {
				return fmt.Errorf("error while trying to stop %s service for host %s: %s", key, host.Name(), err)
			}
		}

		log.Printf("[INFO] updating service '%s' with policy '%s' for host '%s'", key, policy, host.Name())

		if err = hss.UpdatePolicy(ctx, key, policy); err != nil {
			return fmt.Errorf("error while trying to update policy for %s service for host '%s': %s", key, host.Name(), err)
		}

		return nil
	}

	return fmt.Errorf("could not obtain config manager for host '%s' to set state for service '%s'", host.Name(), key)
}

func GetServiceKeyMsg() string {
	srvKeyOptMsg := ""
	for _, key := range ServiceKeyList {
		srvKeyOptMsg += "'" + key + "', "
	}

	return srvKeyOptMsg
}
