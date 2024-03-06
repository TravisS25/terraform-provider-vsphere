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

func TestAccDataSourceVSphereVcenterSNMP_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariablesF(
				t,
				[]string{
					"TF_VAR_VSPHERE_VCENTER_SSH_USER",
					"TF_VAR_VSPHERE_VCENTER_SSH_PASSWORD",
				},
			)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereVcenterSNMPConfig(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_vcenter_snmp.vcenter",
						"id",
						regexp.MustCompile("tf-vcenter-snmp"),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereVcenterSNMPConfig(useHostname bool) string {
	resourceStr :=
		`
	%s

	resource "vsphere_vcenter_snmp" "vcenter" {
		user = "%s"
		password = "%s"
		read_only_communities = ["public"]
		engine_id = "80001ADC0517464555781707920697"
		authentication_protocol = "SHA1"
		privacy_protocol = "AES128"
		remote_user {
			name = "user"
			authentication_password = "password"
			privacy_secret = "123456789abcdefg"
		}
		trap_target {
			hostname = "example.com"
			port = 161
			community = "public"
		}
	}

	data "vsphere_vcenter_snmp" "vcenter" {
		user = "%s"
		password = "%s"
	}
	`

	if useHostname {
		return fmt.Sprintf(
			resourceStr,
			testhelper.CombineConfigs(
				testhelper.ConfigDataRootDC1(),
				testhelper.ConfigDataRootComputeCluster1(),
				testhelper.ConfigDataRootHost1(),
			),
			os.Getenv("TF_VAR_VSPHERE_VCENTER_SSH_USER"),
			os.Getenv("TF_VAR_VSPHERE_VCENTER_SSH_PASSWORD"),
			os.Getenv("TF_VAR_VSPHERE_VCENTER_SSH_USER"),
			os.Getenv("TF_VAR_VSPHERE_VCENTER_SSH_PASSWORD"),
		)
	}

	return fmt.Sprintf(
		resourceStr,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost1(),
		),
		os.Getenv("TF_VAR_VSPHERE_VCENTER_SSH_USER"),
		os.Getenv("TF_VAR_VSPHERE_VCENTER_SSH_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_VCENTER_SSH_USER"),
		os.Getenv("TF_VAR_VSPHERE_VCENTER_SSH_PASSWORD"),
	)
}
