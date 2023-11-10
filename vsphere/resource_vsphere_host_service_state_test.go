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

func TestAccResourceVSphereHostServiceState_basic(t *testing.T) {
	policy := types.HostServicePolicyOn
	newPolicy := types.HostServicePolicyOff

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccVSphereHostServiceStateEnvCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostServiceStateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostServiceStateConfig(policy),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostServiceStateExists("vsphere_host_service_state.h1"),
				),
			},
			{
				Config: testAccResourceVSphereHostServiceStateConfig(newPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostServiceStateWithPolicy("vsphere_host_service_state.h1", newPolicy),
				),
			},
			{
				ResourceName: "vsphere_host_service_state.h1",
				Config:       testAccResourceVSphereHostServiceStateConfig(newPolicy),
				ImportState:  true,
			},
		},
	})
}

func testAccVSphereHostServiceStateDestroy(s *terraform.State) error {
	resourceName := "vsphere_host_service_state"

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceName {
			continue
		}

		client := testAccProvider.Meta().(*Client).vimClient
		hsList, err := hostservicestate.GetHostServies(client, rs.Primary.ID, provider.DefaultAPITimeout)
		if err != nil {
			return fmt.Errorf("error trying to retrieve services for host '%s': %s", rs.Primary.ID, err)
		}

		for _, hs := range hsList {
			if hs.Key == os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY") {
				if hs.Running {
					return fmt.Errorf("service '%s' should not be running", hs.Key)
				} else {
					return nil
				}
			}
		}

		return fmt.Errorf("could not find service with key '%s'", os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY"))
	}

	return fmt.Errorf("could not find resource '%s'", resourceName)
}

func testAccResourceVSphereHostServiceStateConfig(policy types.HostServicePolicy) string {
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
	}
	`,
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigDataRootHost1(),
		os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY"),
		policy,
	)
}

func testAccVSphereHostServiceStateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		client := testAccProvider.Meta().(*Client).vimClient
		key := hostservicestate.HostServiceKey(os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY"))
		_, err := hostservicestate.GetServiceState(
			client,
			rs.Primary.ID,
			key,
			provider.DefaultAPITimeout,
		)
		if err != nil {
			return fmt.Errorf("error trying to retrieve service state for host '%s': %s", rs.Primary.ID, err)
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
		client := testAccProvider.Meta().(*Client).vimClient
		key := hostservicestate.HostServiceKey(os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY"))
		ss, err := hostservicestate.GetServiceState(
			client,
			rs.Primary.ID,
			key,
			provider.DefaultAPITimeout,
		)
		if err != nil {
			return fmt.Errorf("error trying to retrieve service state for host '%s': %s", rs.Primary.ID, err)
		}

		if ss["policy"].(string) != string(policy) {
			return fmt.Errorf("expected service state: %s; got %s", policy, ss["policy"].(string))
		}

		return nil
	}
}

func testAccVSphereHostServiceStateEnvCheck(t *testing.T) {
	found := false

	for _, v := range hostservicestate.ServiceKeyList {
		if v == os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY") {
			found = true
		}
	}

	if !found {
		t.Fatalf("'TF_VAR_VSPHERE_SERVICE_KEY' env variable must be set to valid service key")
	}
}
