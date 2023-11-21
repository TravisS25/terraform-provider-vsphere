// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereIscsiTarget_basic(t *testing.T) {
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
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereIscsiTargetConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_iscsi_target.h1",
						"id",
						regexp.MustCompile("^host-"),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereIscsiTargetConfig() string {
	return fmt.Sprintf(
		`
		%s

		resource "vsphere_iscsi_target" "h1" {
			host_system_id = data.vsphere_host.roothost1.id
			adapter_id = "%s"

			send_target{
				ip = "172.16.0.1"
			}
		}

		data "vsphere_iscsi_target" "h1" {
			host_system_id = vsphere_iscsi_target.h1.host_system_id
			adapter_id = "%s"
		}

		`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost1(),
		),
		os.Getenv("TF_VAR_VSPHERE_ISCSI_ADAPTER_ID"),
		os.Getenv("TF_VAR_VSPHERE_ISCSI_ADAPTER_ID"),
	)
}
