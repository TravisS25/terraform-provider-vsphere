// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceVSphereVcenterSyslog_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereVcenterSyslogDestroy("data.vsphere_vcenter_syslog.syslog"),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereVcenterSyslogConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_vcenter_syslog.syslog",
						"id",
						regexp.MustCompile(vAppSyslogID),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereVcenterSyslogConfig() string {
	return `data "vsphere_vcenter_syslog" "syslog" {}`
}
