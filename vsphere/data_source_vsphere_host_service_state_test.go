// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostservicestate"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/vmware/govmomi/vim25/types"
)

func TestAccDataSourceVSphereHostServiceState_basic(t *testing.T) {
	resourceName := "data.vsphere_host_service_state.h1"
	policy := types.HostServicePolicyOn

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereHostServiceStateEnvCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereHostServiceStateConfig(policy),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						resourceName,
						"id",
						regexp.MustCompile("^host-"),
					),
					testAccDataSourceVSphereHostServiceStateDataCheck(resourceName, policy),
				),
			},
		},
	})
}

func testAccDataSourceVSphereHostServiceStateConfig(policy types.HostServicePolicy) string {
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

		data "vsphere_host_service_state" "h1" {
			host_system_id = vsphere_host_service_state.h1.id
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

func testAccDataSourceVSphereHostServiceStateEnvCheck(t *testing.T) {
	envVars := []string{"TF_VAR_VSPHERE_DATACENTER", "TF_VAR_VSPHERE_CLUSTER", "TF_VAR_VSPHERE_ESXI1", "TF_VAR_VSPHERE_SERVICE_KEY_1"}

	for _, v := range envVars {
		if os.Getenv(v) == "" {
			t.Fatalf("Must set env variable '%s'", v)
		}
	}
}

func testAccDataSourceVSphereHostServiceStateDataCheck(resourceName string, policy types.HostServicePolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%s key not found on the server", resourceName)
		}

		client := testAccProvider.Meta().(*Client).vimClient

		hsList, err := hostservicestate.GetHostServies(client, rs.Primary.ID, provider.DefaultAPITimeout)
		if err != nil {
			return fmt.Errorf("error trying to get host services from host '%s'", err)
		}

		for _, hs := range hsList {
			if hs.Key == os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_1") {
				if !hs.Running {
					return fmt.Errorf("service '%s' is not running", os.Getenv("TF_VAR_VSPHERE_SERVICE_KEY_1"))
				}
				if hs.Policy != string(policy) {
					return fmt.Errorf("service '%s' does not have policy '%s'; got '%s'", hs.Key, policy, hs.Policy)
				}
			}
		}

		return nil
	}
}
