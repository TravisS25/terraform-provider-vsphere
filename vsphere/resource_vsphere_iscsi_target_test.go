// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/iscsi"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/vmware/govmomi/vim25/types"
)

func TestAccResourceVSphereIscsiTarget_basic(t *testing.T) {
	resourceName := "vsphere_iscsi_target.target"

	staticTargetIP := "172.20.0.1"
	sendTargetIP := "172.20.1.1"

	newStaticTargetIP := "172.21.0.1"
	newSendTargetIP := "172.21.1.1"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariablesF(
				t,
				[]string{
					"TF_VAR_VSPHERE_DATACENTER",
					"TF_VAR_VSPHERE_CLUSTER",
					"TF_VAR_VSPHERE_ESXI1",
					"TF_VAR_VSPHERE_ISCSI_ADAPTER_ID",
				},
			)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereIscsiTargetDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereIscsiTargetConfig(staticTargetIP, sendTargetIP),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereIscsiTargetValidation(resourceName, staticTargetIP, sendTargetIP),
				),
			},
			{
				Config: testAccResourceVSphereIscsiTargetConfig(newStaticTargetIP, newSendTargetIP),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereIscsiTargetValidation(resourceName, newStaticTargetIP, newSendTargetIP),
				),
			},
			{
				ResourceName: resourceName,
				Config:       testAccResourceVSphereIscsiTargetConfig(newStaticTargetIP, newSendTargetIP),
				ImportState:  true,
			},
		},
	})
}

func testAccVSphereIscsiTargetValidation(resourceName, staticTargetIP, sendTargetIP string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%s key not found on the server", resourceName)
		}
		hostID := strings.Split(rs.Primary.ID, ":")[0]
		client := testAccProvider.Meta().(*Client).vimClient
		hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
		if err != nil {
			return err
		}

		adapterID := os.Getenv("TF_VAR_VSPHERE_ISCSI_ADAPTER_ID")
		baseAdapter, err := iscsi.GetIscsiAdater(hssProps, hostID, adapterID)
		if err != nil {
			return err
		}

		adapter := baseAdapter.(*types.HostInternetScsiHba)

		if len(adapter.ConfiguredSendTarget) == 0 {
			return fmt.Errorf("there are no send targets for adapter '%s'", adapterID)
		}
		if len(adapter.ConfiguredStaticTarget) == 0 {
			return fmt.Errorf("there are no static targets for adapter '%s'", adapterID)
		}

		if adapter.ConfiguredSendTarget[0].Address != sendTargetIP {
			return fmt.Errorf(
				"invalid ip for send target.  expected '%s'; got '%s'",
				sendTargetIP,
				adapter.ConfiguredSendTarget[0].Address,
			)
		}
		if adapter.ConfiguredStaticTarget[0].Address != staticTargetIP {
			return fmt.Errorf(
				"invalid ip for static target.  expected '%s'; got '%s'",
				staticTargetIP,
				adapter.ConfiguredStaticTarget[0].Address,
			)
		}

		return nil
	}
}

func testAccVSphereIscsiTargetDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		hostID := strings.Split(rs.Primary.ID, ":")[0]
		client := testAccProvider.Meta().(*Client).vimClient
		hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, hostID)
		if err != nil {
			return err
		}

		adapterID := os.Getenv("TF_VAR_VSPHERE_ISCSI_ADAPTER_ID")
		baseAdapter, err := iscsi.GetIscsiAdater(hssProps, hostID, adapterID)
		if err != nil {
			return err
		}

		adapter := baseAdapter.(*types.HostInternetScsiHba)

		if len(adapter.ConfiguredSendTarget) > 0 {
			return fmt.Errorf("send targets still exists for adapter '%s'", adapterID)
		}
		if len(adapter.ConfiguredStaticTarget) > 0 {
			return fmt.Errorf("static targets still exists for adapter '%s'", adapterID)
		}

		return nil
	}
}

func testAccResourceVSphereIscsiTargetConfig(staticTargetIP, sendTargetIP string) string {
	return fmt.Sprintf(
		`
	%s

	resource "vsphere_iscsi_target" "target" {
		host_system_id = data.vsphere_host.roothost1.id
		adapter_id     = "%s"

		static_target{
			ip = "%s"
			name = "iqn.1998-01.com.static_test_1"
			chap {
				outgoing_creds {
					username = "user"
					password = "password"
				}
				incoming_creds {
					username = "user"
					password = "password"
				}
			}
		}
		send_target{
			ip = "%s"
			chap {
				outgoing_creds {
					username = "user"
					password = "password"
				}
				incoming_creds {
					username = "user"
					password = "password"
				}
			}
		}
	}
	`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost1(),
		),
		os.Getenv("TF_VAR_VSPHERE_ISCSI_ADAPTER_ID"),
		staticTargetIP,
		sendTargetIP,
	)
}
