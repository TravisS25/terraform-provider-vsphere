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

func TestAccDataSourceVSphereHostConfigSNMP_basic(t *testing.T) {
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
					"TF_VAR_vsphere_esxi_ssh_user",
					"TF_VAR_vsphere_esxi_ssh_password",
				},
			)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereHostConfigSNMPConfig(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_host_config_snmp.h1",
						"id",
						regexp.MustCompile("^host-"),
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereHostConfigSNMP_hostname(t *testing.T) {
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
					"TF_VAR_vsphere_esxi_ssh_user",
					"TF_VAR_vsphere_esxi_ssh_password",
				},
			)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereHostConfigSNMPConfig(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_host_config_snmp.h1",
						"id",
						regexp.MustCompile(os.Getenv("TF_VAR_VSPHERE_ESXI1")),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereHostConfigSNMPConfig(useHostname bool) string {
	resourceStr :=
		`
	%s

	resource "vsphere_host_config_snmp" "h1" {
		%s
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

	data "vsphere_host_config_snmp" "h1" {
		%s
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
			"hostname = data.vsphere_host.roothost1.name",
			os.Getenv("TF_VAR_vsphere_esxi_ssh_user"),
			os.Getenv("TF_VAR_vsphere_esxi_ssh_password"),
			"hostname = vsphere_host_config_snmp.h1.hostname",
			os.Getenv("TF_VAR_vsphere_esxi_ssh_user"),
			os.Getenv("TF_VAR_vsphere_esxi_ssh_password"),
		)
	}

	return fmt.Sprintf(
		resourceStr,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost1(),
		),
		"host_system_id = data.vsphere_host.roothost1.id",
		os.Getenv("TF_VAR_vsphere_esxi_ssh_user"),
		os.Getenv("TF_VAR_vsphere_esxi_ssh_password"),
		"host_system_id = vsphere_host_config_snmp.h1.host_system_id",
		os.Getenv("TF_VAR_vsphere_esxi_ssh_user"),
		os.Getenv("TF_VAR_vsphere_esxi_ssh_password"),
	)
}
