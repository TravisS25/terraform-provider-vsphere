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
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/vmware/govmomi/object"
)

func TestAccResourceVSphereHostConfigSyslog_basic(t *testing.T) {
	resourceName := "vsphere_host_config_syslog.h1"

	logLvl := "info"
	newLogLvl := "debug"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariablesF(t, []string{"ESXI_LOG_HOST"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostConfigSyslogDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostConfigSyslogConfig(logLvl),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostConfigSyslogValidate(resourceName, logLvl),
				),
			},
			{
				Config: testAccResourceVSphereHostConfigSyslogConfig(newLogLvl),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostConfigSyslogValidate(resourceName, newLogLvl),
				),
			},
			{
				ResourceName: resourceName,
				Config:       testAccResourceVSphereHostConfigSyslogConfig(newLogLvl),
				ImportState:  true,
			},
		},
	})
}

func testAccResourceVSphereHostConfigSyslogDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("'%s' key not found on the server", name)
		}

		client := testAccProvider.Meta().(*Client).vimClient
		hostID := rs.Primary.ID
		ctx := context.Background()
		optManager, err := hostconfig.GetOptionManager(ctx, client, hostID)
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

func testAccResourceVSphereHostConfigSyslogConfig(logLvl string) string {
	return fmt.Sprintf(
		`
		%s

		resource "vsphere_host_config_syslog" "h1" {
			host_system_id = data.vsphere_host.roothost1.id
			log_host = "%s"
			log_level = "%s"
		}
		`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost1(),
		),
		os.Getenv("ESXI_LOG_HOST"),
		logLvl,
	)
}

func testAccResourceVSphereHostConfigSyslogValidate(resourceName, logLvl string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%s key not found on the server", resourceName)
		}

		client := testAccProvider.Meta().(*Client).vimClient
		hostID := rs.Primary.ID
		ctx := context.Background()
		optManager, err := hostconfig.GetOptionManager(ctx, client, hostID)
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
