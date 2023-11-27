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

func TestAccDataSourceVSphereIscsiSoftwareAdapter_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariablesF(
				t,
				[]string{"TF_VAR_VSPHERE_DATACENTER", "TF_VAR_VSPHERE_CLUSTER", "TF_VAR_VSPHERE_ESXI1"},
			)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereIscsiSoftwareAdapterConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_iscsi_software_adapter.h1",
						"id",
						regexp.MustCompile("^host-"),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereIscsiSoftwareAdapterConfig() string {
	return fmt.Sprintf(
		`
		%s

		resource "vsphere_iscsi_software_adapter" "h1" {
			host_system_id = data.vsphere_host.roothost1.id
		}

		data "vsphere_iscsi_software_adapter" "h1" {
			host_system_id = vsphere_iscsi_software_adapter.h1.host_system_id
		}

		`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost1(),
		),
	)
}
