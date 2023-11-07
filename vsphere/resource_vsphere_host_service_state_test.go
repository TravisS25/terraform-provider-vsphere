package vsphere

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostservicestate"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/vmware/govmomi/vim25/types"
)

func TestAccResourceVSphereHostServiceState_basic(t *testing.T) {
	policy := types.HostServicePolicyOn
	newPolicy := types.HostServicePolicyAutomatic

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_VSPHERE_SERVICE_KEY"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostServiceStateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostServiceStateConfig(policy, true),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostServiceStateExists("vsphere_host_service_state.h1"),
				),
			},
			{
				Config: testAccResourceVSphereHostServiceStateConfig(newPolicy, true),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostServiceStateWithPolicy("vsphere_host_service_state.h1", newPolicy),
				),
			},
			{
				ResourceName:      "vsphere_host_service_state.h1",
				Config:            testAccResourceVSphereHostServiceStateConfig(newPolicy, true),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVSphereHostServiceStateDestroy(s *terraform.State) error {
	message := ""
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_host_service_state" {
			continue
		}
		id := strings.Split(rs.Primary.ID, ":")

		if len(id) != 2 {
			return fmt.Errorf("invalid id for resource 'vsphere_host_service_state'.  Given: %v", id)
		}

		client := testAccProvider.Meta().(*Client).vimClient
		ss, err := hostservicestate.GetServiceState(client, id[0], id[1], provider.DefaultAPITimeout)
		if err != nil {
			return fmt.Errorf("error trying to retrieve service state for host '%s': %s", id[0], err)
		}

		if ss.Policy != types.HostServicePolicyOff {
			message += " service policy should be 'off' /"
		}
		if ss.Running {
			message += " service should not be running"
		}
	}
	if message != "" {
		return errors.New(message)
	}
	return nil
}

func testAccResourceVSphereHostServiceStateConfig(policy types.HostServicePolicy, running bool) string {
	return fmt.Sprintf(
		`
	%s

	%s

	%s

	resource "vsphere_host_service_state" "h1" {
		host_system_id = data.vsphere_host.roothost1.id
		key = "%s"
		running = %v
		policy = "%s"
	}
	`,
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigDataRootHost1(),
		os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY"),
		running,
		policy,
	)
}

func testAccVSphereHostServiceStateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		id := strings.Split(rs.Primary.ID, ":")
		client := testAccProvider.Meta().(*Client).vimClient
		_, err := hostservicestate.GetServiceState(client, id[0], id[1], provider.DefaultAPITimeout)
		if err != nil {
			return fmt.Errorf("error trying to retrieve service state for host '%s': %s", id[0], err)
		}

		return nil
	}
}

func testAccVSphereHostServiceStateWithPolicy(resourceName string, policy types.HostServicePolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%s key not found on the server", resourceName)
		}
		id := strings.Split(rs.Primary.ID, ":")
		client := testAccProvider.Meta().(*Client).vimClient
		ss, err := hostservicestate.GetServiceState(client, id[0], id[1], provider.DefaultAPITimeout)
		if err != nil {
			return fmt.Errorf("error trying to retrieve service state for host '%s': %s", id[0], err)
		}

		if ss.Policy != policy {
			return fmt.Errorf("expected service state: %s; got %s", policy, ss.Policy)
		}

		return nil
	}
}
