package vsphere

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostservicestate"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/vmware/govmomi/vim25/types"
)

func TestAccResourceVSphereHostConfigDateTime_basic(t *testing.T) {
	policy := types.HostServicePolicyOn
	newPolicy := types.HostServicePolicyOff
	resourceName := "vsphere_host_config_date_time.h1"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			//testAccCheckEnvVariables(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostConfigDateTimeDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereTwoHostConfigDateTimeConfig(policy),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostConfigDateTimeValidateServicesRunning(resourceName, true),
				),
			},
			{
				Config: testAccResourceVSphereOneHostConfigDateTimeConfig(newPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostConfigDateTimeValidateServicesRunning(resourceName, false),
				),
			},
			{
				ResourceName: resourceName,
				Config:       testAccResourceVSphereOneHostConfigDateTimeConfig(newPolicy),
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

		hsList, err := hostservicestate.GetHostServies(client, rs.Primary.ID, provider.DefaultAPITimeout)
		if err != nil {
			return fmt.Errorf("error trying to get host services from host '%s'", err)
		}

		srvKey1Running := false
		srvKey2Running := false

		for _, hs := range hsList {
			if hs.Key == os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_1") && hs.Running {
				srvKey1Running = true
			}
			if hs.Key == os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_2") && hs.Running {
				srvKey2Running = true
			}
		}

		if srvKey1Running {
			return fmt.Errorf("service '%s' is still running", os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_1"))
		}

		if srvKey2Running {
			return fmt.Errorf("service '%s' is still running", os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_2"))
		}

		return nil
	}
}

func testAccResourceVSphereOneHostConfigDateTimeConfig(policy types.HostServicePolicy) string {
	return fmt.Sprintf(
		`
		%s

		resource "vsphere_host_service_state" "h1" {
			host_system_id = data.vsphere_host.roothost1.id
			service {
				key = "%s"
				policy = "%s"
			}
		}
		`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost1(),
		),
		os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_1"),
		policy,
	)
}

func testAccResourceVSphereTwoHostConfigDateTimeConfig(policy types.HostServicePolicy) string {
	return fmt.Sprintf(
		`
	%s

	%s

	%s

	resource "vsphere_host_service_state" "h1" {
		host_system_id = data.vsphere_host.roothost1.id
		service {
			key = "%s"
			policy = "%s"
		}
		service {
			key = "%s"
			policy = "%s"
		}
	}
	`,
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigDataRootHost1(),
		os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_1"),
		policy,
		os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_2"),
		policy,
	)
}

func testAccResourceVSphereHostConfigDateTimeValidateServicesRunning(name string, twoServicesRunning bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		client := testAccProvider.Meta().(*Client).vimClient

		hsList, err := hostservicestate.GetHostServies(client, rs.Primary.ID, provider.DefaultAPITimeout)
		if err != nil {
			return fmt.Errorf("error trying to get host services from host '%s'", err)
		}

		srvKey1Running := false
		srvKey2Running := false

		for _, hs := range hsList {
			if hs.Key == os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_1") && hs.Running {
				srvKey1Running = true
			}
			if hs.Key == os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_2") && hs.Running {
				srvKey2Running = true
			}
		}

		if !srvKey1Running {
			return fmt.Errorf("service '%s' is not running", os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_1"))
		}

		if !srvKey2Running && twoServicesRunning {
			return fmt.Errorf("service '%s' is not running", os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_2"))
		}

		if srvKey2Running && !twoServicesRunning {
			return fmt.Errorf("service '%s' is running when it should be turned off", os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_2"))
		}

		return nil
	}
}
