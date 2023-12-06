// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereHostConfigDateTime_basic(t *testing.T) {
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
				},
			)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereHostConfigDateTimeConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_host_config_date_time.h1",
						"id",
						regexp.MustCompile("^host-"),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereHostConfigDateTimeConfig() string {
	return fmt.Sprintf(
		`
		%s

		resource "vsphere_host_config_date_time" "h1" {
			host_system_id = data.vsphere_host.roothost1.id
			ntp_servers = ["0.us.pool.ntp.org"]
		}

		data "vsphere_host_config_date_time" "h1" {
			host_system_id = vsphere_host_config_date_time.h1.host_system_id
		}

		`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost1(),
		),
	)
}
