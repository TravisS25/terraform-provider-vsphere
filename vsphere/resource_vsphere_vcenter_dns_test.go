// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/viapi"
)

func TestAccResourceVSphereVcenterDNS_basic(t *testing.T) {
	resourceName := "vsphere_vcenter_dns.dns"
	createServer := "172.22.208.11"

	updateServerStr := ""
	envServers := strings.Split(os.Getenv("TF_VAR_VSPHERE_VCENTER_DNS_SERVERS"), ",")
	updateServers := make([]string, 0, len(envServers))

	for i, s := range envServers {
		srv := strings.TrimSpace(s)
		updateServers = append(updateServers, srv)
		updateServerStr += `"` + srv + `"`

		if i != len(envServers)-1 {
			updateServerStr += ", "
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariablesF(t, []string{"TF_VAR_VSPHERE_VCENTER_DNS_SERVERS"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereVcenterDNSValidation(resourceName, updateServers),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVcenterDNSConfig(`"` + createServer + `"`),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVcenterDNSValidation(resourceName, []string{createServer}),
				),
			},
			{
				Config: testAccResourceVSphereVcenterDNSConfig(updateServerStr),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVcenterDNSValidation(resourceName, updateServers),
				),
			},
			{
				ResourceName: resourceName,
				Config:       testAccResourceVSphereVcenterDNSConfig(updateServerStr),
				ImportState:  true,
			},
		},
	})
}

func testAccResourceVSphereVcenterDNSValidation(resourceName string, givenServers []string) resource.TestCheckFunc {
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

			counter := 0

			for _, givenSrv := range givenServers {
				for _, resSrv := range servers {
					if givenSrv == resSrv {
						counter++
					}
				}
			}

			if counter != len(givenServers) || len(servers) != len(givenServers) {
				return fmt.Errorf(
					"given servers do not match api response servers: given servers: %v; api response servers: %v",
					givenServers,
					servers,
				)
			}
		} else {
			return fmt.Errorf("did not receive server list")
		}

		return nil
	}
}

func testAccResourceVSphereVcenterDNSConfig(serverStr string) string {
	return fmt.Sprintf(
		`
		resource "vsphere_vcenter_dns" "dns" {
			servers = [%s]
		}
		`,
		serverStr,
	)
}
