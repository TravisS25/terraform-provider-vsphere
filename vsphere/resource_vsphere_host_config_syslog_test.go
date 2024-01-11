package vsphere

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostconfig"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/vmware/govmomi/object"
)

const (
	hostConfigSyslogResourceName = "vsphere_host_config_syslog.h1"
	hostConfigSyslogLogLvl       = "info"
	hostConfigSyslogNewLogLvl    = "debug"
)

func TestAccResourceVSphereHostConfigSyslog_UsingHostSystemID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariablesF(t, []string{"ESX_LOG_HOST", "TF_VAR_VSPHERE_ESXI1"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostConfigSyslogDestroy(hostConfigSyslogResourceName, false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostConfigSyslogConfig(hostConfigSyslogLogLvl, false),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostConfigSyslogValidate(hostConfigSyslogResourceName, hostConfigSyslogLogLvl, false),
				),
			},
			{
				Config: testAccResourceVSphereHostConfigSyslogConfig(hostConfigSyslogNewLogLvl, false),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostConfigSyslogValidate(hostConfigSyslogResourceName, hostConfigSyslogNewLogLvl, false),
				),
			},
			{
				ResourceName: hostConfigSyslogResourceName,
				Config:       testAccResourceVSphereHostConfigSyslogConfig(hostConfigSyslogNewLogLvl, false),
				ImportState:  true,
			},
		},
	})
}

func TestAccResourceVSphereHostConfigSyslog_UsingHostname(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariablesF(t, []string{"ESX_LOG_HOST", "TF_VAR_VSPHERE_ESXI1"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostConfigSyslogDestroy(hostConfigSyslogResourceName, true),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostConfigSyslogConfig(hostConfigSyslogLogLvl, true),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostConfigSyslogValidate(hostConfigSyslogResourceName, hostConfigSyslogLogLvl, true),
				),
			},
			{
				Config: testAccResourceVSphereHostConfigSyslogConfig(hostConfigSyslogNewLogLvl, true),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostConfigSyslogValidate(hostConfigSyslogResourceName, hostConfigSyslogNewLogLvl, true),
				),
			},
			{
				ResourceName: hostConfigSyslogResourceName,
				Config:       testAccResourceVSphereHostConfigSyslogConfig(hostConfigSyslogNewLogLvl, true),
				ImportState:  true,
			},
		},
	})
}

func testAccResourceVSphereHostConfigSyslogDestroy(name string, useHostname bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("'%s' key not found on the server", name)
		}

		client := testAccProvider.Meta().(*Client).vimClient
		hostID := rs.Primary.ID
		ctx := context.Background()

		var host *object.HostSystem
		var err error

		if useHostname {
			host, err = hostsystem.FromHostname(client, hostID)
		} else {
			host, err = hostsystem.FromID(client, hostID)
		}

		if err != nil {
			return err
		}

		optManager, err := hostconfig.GetOptionManager(client, host)
		if err != nil {
			return err
		}

		hostOpts, err := optManager.Query(ctx, hostconfig.SyslogHostKey)
		if err != nil {
			return fmt.Errorf("error querying for log host on host '%s': %s", hostID, err)
		}

		if len(hostOpts) > 0 && hostOpts[0].GetOptionValue().Value != "" {
			return fmt.Errorf(
				"log host '%s' still exists after delete for host: '%s'",
				hostOpts[0].GetOptionValue().Value,
				hostID,
			)
		}

		return nil
	}
}

func testAccResourceVSphereHostConfigSyslogConfig(logLvl string, useHostname bool) string {
	idStr := "host_system_id = data.vsphere_host.roothost1.id"

	if useHostname {
		idStr = `hostname = "` + os.Getenv("TF_VAR_VSPHERE_ESXI1") + `"`
	}

	return fmt.Sprintf(
		`
		%s

		resource "vsphere_host_config_syslog" "h1" {
			%s
			log_host = "%s"
			log_level = "%s"
		}
		`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost1(),
		),
		idStr,
		os.Getenv("ESX_LOG_HOST"),
		logLvl,
	)
}

func testAccResourceVSphereHostConfigSyslogValidate(resourceName, logLvl string, useHostname bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%s key not found on the server", resourceName)
		}

		client := testAccProvider.Meta().(*Client).vimClient
		hostID := rs.Primary.ID
		ctx := context.Background()

		var host *object.HostSystem
		var err error

		if useHostname {
			host, err = hostsystem.FromHostname(client, hostID)
		} else {
			host, err = hostsystem.FromID(client, hostID)
		}

		if err != nil {
			return err
		}

		optManager, err := hostconfig.GetOptionManager(client, host)
		if err != nil {
			return err
		}

		if err = testHostConfigSyslogKey(ctx, optManager, hostconfig.SyslogLogLevelKey, hostID, logLvl); err != nil {
			return err
		}

		return nil
	}
}

func testHostConfigSyslogKey(ctx context.Context, optManager *object.OptionManager, key, hostID string, expectedVal interface{}) error {
	queryOpts, err := optManager.Query(ctx, key)
	if err != nil {
		return fmt.Errorf("error querying against key '%s' on host '%s': %s", key, hostID, err)
	}

	if len(queryOpts) == 0 {
		return fmt.Errorf("no values returned from key '%s' on host: '%s'", key, hostID)
	}

	if !reflect.DeepEqual(queryOpts[0].GetOptionValue().Value, expectedVal) {
		return fmt.Errorf(
			"expected key '%s' to have value: '%s'; got: '%s'; for host '%s'",
			key,
			expectedVal,
			queryOpts[0].GetOptionValue().Value,
			hostID,
		)
	}

	return nil
}
