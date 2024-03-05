package vsphere

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/vmware/govmomi/vim25/mo"
)

func TestAccResourceVSphereHostConfigDateTime_basic(t *testing.T) {
	server := "0.us.pool.ntp.org"
	newServer := "1.us.pool.ntp.org"
	resourceName := "vsphere_host_config_date_time.h1"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariablesF(
				t,
				[]string{"TF_VAR_VSPHERE_DATACENTER", "TF_VAR_VSPHERE_CLUSTER", "TF_VAR_VSPHERE_ESXI1"},
			)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostConfigDateTimeDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostConfigDateTimeConfig(server, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostConfigDateTimeValidation(resourceName, server),
				),
			},
			{
				Config: testAccResourceVSphereHostConfigDateTimeConfig(newServer, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostConfigDateTimeValidation(resourceName, newServer),
				),
			},
			{
				ResourceName: resourceName,
				Config:       testAccResourceVSphereHostConfigDateTimeConfig(resourceName, false),
				ImportState:  true,
			},
		},
	})
}

func TestAccResourceVSphereHostConfigDateTime_hostname(t *testing.T) {
	server := "0.us.pool.ntp.org"
	newServer := "1.us.pool.ntp.org"
	resourceName := "vsphere_host_config_date_time.h1"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariablesF(
				t,
				[]string{"TF_VAR_VSPHERE_DATACENTER", "TF_VAR_VSPHERE_CLUSTER", "TF_VAR_VSPHERE_ESXI1"},
			)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostConfigDateTimeDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostConfigDateTimeConfig(server, true),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostConfigDateTimeValidation(resourceName, server),
				),
			},
			{
				Config: testAccResourceVSphereHostConfigDateTimeConfig(newServer, true),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostConfigDateTimeValidation(resourceName, newServer),
				),
			},
			{
				ResourceName: resourceName,
				Config:       testAccResourceVSphereHostConfigDateTimeConfig(resourceName, true),
				ImportState:  true,
			},
		},
	})
}

func testAccResourceVSphereHostConfigDateTimeDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}

		client := testAccProvider.Meta().(*Client).vimClient
		host, _, err := hostsystem.CheckIfHostnameOrID(client, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error retrieving host for 'vsphere_host_config_date_time' on delete test: %s", err)
		}

		hostDt, err := host.ConfigManager().DateTimeSystem(context.Background())
		if err != nil {
			return fmt.Errorf("error trying to get datetime system object from host '%s': %s", host.Name(), err)
		}

		var hostDtProps mo.HostDateTimeSystem
		if err = hostDt.Properties(context.Background(), hostDt.Reference(), nil, &hostDtProps); err != nil {
			return fmt.Errorf("error trying to gather datetime properties from host '%s': %s", host.Name(), err)
		}

		if len(hostDtProps.DateTimeInfo.NtpConfig.Server) > 0 {
			return fmt.Errorf("ntp server not destroyed")
		}

		return nil
	}
}

func testAccVSphereHostConfigDateTimeValidation(resourceName, server string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%s key not found on the server", resourceName)
		}
		client := testAccProvider.Meta().(*Client).vimClient
		host, _, err := hostsystem.CheckIfHostnameOrID(client, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error retrieving host for 'vsphere_host_config_date_time' on delete test: %s", err)
		}

		hostDt, err := host.ConfigManager().DateTimeSystem(context.Background())
		if err != nil {
			return fmt.Errorf("error trying to get datetime system object from host '%s': %s", host.Name(), err)
		}

		var hostDtProps mo.HostDateTimeSystem
		if err = hostDt.Properties(context.Background(), hostDt.Reference(), nil, &hostDtProps); err != nil {
			return fmt.Errorf("error trying to gather datetime properties from host '%s': %s", host.Name(), err)
		}

		if len(hostDtProps.DateTimeInfo.NtpConfig.Server) > 0 {
			if server != hostDtProps.DateTimeInfo.NtpConfig.Server[0] {
				return fmt.Errorf(
					"invalid server:  expected: '%s'; got: '%s'",
					server,
					hostDtProps.DateTimeInfo.NtpConfig.Server[0],
				)
			}
		} else {
			return fmt.Errorf("there are no ntp servers set")
		}

		return nil
	}
}

func testAccResourceVSphereHostConfigDateTimeConfig(server string, useHostname bool) string {
	resourceStr :=
		`
	%s

	resource "vsphere_host_config_date_time" "h1" {
		%s
		ntp_servers = ["%s"]
	}

	`

	if useHostname {
		return fmt.Sprintf(
			resourceStr,
			testhelper.CombineConfigs(
				testhelper.ConfigDataRootDC1(),
				testhelper.ConfigDataRootComputeCluster1(),
				testhelper.ConfigDataRootHost1(),
			),
			"hostname = data.vsphere_host.roothost1.name",
			server,
		)
	}

	return fmt.Sprintf(
		resourceStr,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost1(),
		),
		"host_system_id = data.vsphere_host.roothost1.id",
		server,
	)
}
