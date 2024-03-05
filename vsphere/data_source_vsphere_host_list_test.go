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

func TestAccDataSourceVSphereHostList_basic(t *testing.T) {
	resourceName := "data.vsphere_host_list.h1"

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
				Config: testAccDataSourceVSphereHostListConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						resourceName,
						"id",
						regexp.MustCompile("^datacenter-"),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereHostListConfig() string {
	return fmt.Sprintf(
		`
		%s

		data "vsphere_host_list" "h1" {
			datacenter_id = data.vsphere_datacenter.rootdc1.id
		}
		`,
		testhelper.ConfigDataRootDC1(),
	)
}
