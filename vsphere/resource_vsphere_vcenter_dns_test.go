// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
)

func TestAccResourceVSphereVcenterDNS_basic(t *testing.T) {
	resourceName := "vsphere_vcenter_dns.dns"
	createServer := "172.16.1.10"
	updateServer := "172.16.2.10"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVcenterDNSValidation(resourceName, "", false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVcenterDNSConfig(createServer),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVcenterDNSValidation(resourceName, createServer, true),
				),
			},
			{
				Config: testAccResourceVSphereVcenterDNSConfig(updateServer),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVcenterDNSValidation(resourceName, updateServer, true),
				),
			},
			{
				ResourceName: resourceName,
				Config:       testAccResourceVSphereVcenterDNSConfig(updateServer),
				ImportState:  true,
			},
		},
	})
}

func testAccResourceVSphereVcenterDNSValidation(resourceName, server string, isUpdate bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%s key not found on the server", resourceName)
		}

		client := testAccProvider.Meta().(*Client).restClient
		bodyRes, err := viapi.RestRequest[map[string]interface{}](client, http.MethodGet, dnsServersPath, nil)
		if err != nil {
			return err
		}

		if serverRes, ok := bodyRes["servers"]; ok {
			servers := serverRes.([]interface{})

			if isUpdate {
				if len(servers) != 1 {
					return fmt.Errorf("should have 1 dns server; got %+v", servers)
				}

				if servers[0] != server {
					return fmt.Errorf("server ip should be: '%s'; got '%s'", server, servers[0])
				}
			} else if len(servers) != 0 {
				return fmt.Errorf("should not have any dns servers; got %+v", servers)
			}
		} else {
			return fmt.Errorf("did not receive server list")
		}

		return nil
	}
}

func testAccResourceVSphereVcenterDNSConfig(server string) string {
	return fmt.Sprintf(
		`
		resource "vsphere_vcenter_dns" "dns" {
			servers = ["%s"]
		}
		`,
		server,
	)
}
