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
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"VSPHERE_SERVICE_STATE_KEY"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostServiceStateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostServiceStateConfig(os.Getenv("VSPHERE_SERVICE_STATE_KEY")),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
				),
			},
			{
				ResourceName: "vsphere_host.h1",
				Config:       testaccvspherehostconfigImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
				),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVSphereHostServiceStateDestroy(s *terraform.State) error {
	message := ""
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_host" {
			continue
		}
		id := strings.Split(rs.Primary.ID, ":")
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

func testAccResourceVSphereHostServiceStateConfig(serviceKey string) string {
	return fmt.Sprintf(
		`
	%s

	%s

	%s

	resource "vsphere_host_service_state" "h1" {
		host_system_id = data.vsphere_host.roothost1.id
		key = "%s"
		running = true
		policy = "on"
	}
	`,
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigDataRootHost1(),
		serviceKey,
	)
}
