// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	esxissh "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/ssh"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
)

func TestAccResourceVSphereVcenterSNMP_basic(t *testing.T) {
	testAccCheckEnvVariablesF(
		t,
		[]string{
			"TF_VAR_vsphere_vcenter_ssh_user",
			"TF_VAR_vsphere_vcenter_ssh_password",
			"TF_VAR_vsphere_ssh_known_hosts_path",
		},
	)

	resourceName := "vsphere_vcenter_snmp.h1"
	community := "public"
	newCommunity := "new_public"

	_, err := os.OpenFile(
		os.Getenv("TF_VAR_vsphere_ssh_known_hosts_path"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		os.ModePerm,
	)
	if err != nil {
		t.Fatalf("unable to create file: %s", err)
	}

	if _, err = esxissh.GetKnownHostsOutput(
		os.Getenv("TF_VAR_vsphere_ssh_known_hosts_path"),
		os.Getenv("VSPHERE_SERVER"),
	); err != nil && err == esxissh.ErrHostNotFound {
		runKeyScanCommand(t, os.Getenv("VSPHERE_SERVER"))
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVcenterSNMPDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVcenterSNMPConfig(community),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereVcenterSNMPValidation(resourceName, community),
				),
			},
			{
				Config: testAccResourceVSphereVcenterSNMPConfig(newCommunity),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereVcenterSNMPValidation(resourceName, newCommunity),
				),
			},
			{
				ResourceName: resourceName,
				Config:       testAccResourceVSphereVcenterSNMPConfig(resourceName),
				ImportState:  true,
			},
		},
	})
}

func testAccResourceVSphereVcenterSNMPDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}

		resetClient := testAccProvider.Meta().(*Client).restClient
		resVal, err := viapi.RestRequest[map[string]interface{}](
			resetClient,
			http.MethodGet,
			snmpMonitoringPath,
			nil,
		)
		if err != nil {
			return fmt.Errorf("error retrieving snmp settings: %s", err)
		}

		if resVal["authentication"] != "none" {
			return fmt.Errorf("authentication value should be ''; got '%s'", resVal["authentication"])
		}
		if resVal["privacy"] != "none" {
			return fmt.Errorf("authentication value should be ''; got '%s'", resVal["privacy"])
		}

		return nil
	}
}

func testAccVSphereVcenterSNMPValidation(resourceName, community string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%s key not found on the server", resourceName)
		}

		resetClient := testAccProvider.Meta().(*Client).restClient
		resVal, err := viapi.RestRequest[map[string]interface{}](
			resetClient,
			http.MethodGet,
			snmpMonitoringPath,
			nil,
		)
		if err != nil {
			return fmt.Errorf("error retrieving snmp settings: %s", err)
		}

		communities := resVal["communities"].([]interface{})
		if len(communities) == 1 {
			if communities[0] != community {
				return fmt.Errorf("should have community value of '%s'; got '%s'", community, communities[0])
			}
		} else {
			return fmt.Errorf("should have a length of 1 for communties; got '%d'\n", len(communities))
		}

		return nil
	}
}

func testAccResourceVSphereVcenterSNMPConfig(community string) string {
	return fmt.Sprintf(
		`
		%s

		resource "vsphere_vcenter_snmp" "h1" {
			user = "%s"
			password = "%s"
			known_hosts_path = "%s"
			read_only_communities = ["%s"]
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
		`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost1(),
		),
		os.Getenv("TF_VAR_vsphere_esxi_ssh_user"),
		os.Getenv("TF_VAR_vsphere_esxi_ssh_password"),
		os.Getenv("TF_VAR_vsphere_ssh_known_hosts_path"),
		community,
	)
}
