package vsphere

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
	"fmt"
	//"strings"
	"os"
	"context"
	"github.com/vmware/govmomi/vim25/mo"
)

func TestAccResourceVSphereHostConfigDNS_basic(t *testing.T) {
	resource_name := "vsphere_host_config_dns.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"host_system_id",  "hostname", "dns_servers", "domain_name", "search_domains"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostConfigDNSDestroy,
		Steps: []resource.TestStep{
			{
				// create the original testing resource
				Config: testAccResourceVSphereHostConfigDNSConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostConfigDNSExists(resource_name),
				),
			},
			{
				// change the originally created resources hostname (create a diff and apply an update)
				Config: testAccResourceVSphereHostConfigDNSConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostConfigDNSwithHostname(resource_name),
				),
			},
			{
				ResourceName: resource_name,
				Config:       testAccResourceVSphereHostConfigDNSConfig(),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVSphereHostConfigDNSDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_host_config_dns" {
			continue
		}
	}
	// the delete function in the resource does not actually delete anything so nothing to check here...
	return nil

}

func testAccResourceVSphereHostConfigDNSConfig() string {
	return fmt.Sprintf(
		`
		resource "vsphere_host_config_dns" "test" {
		  host_system_id = "%s"
		  hostname = "%s"
		  dns_servers = ["%s"]
		  domain_name = "%s"
		  search_domains = ["%s"]
		}
	`,
		os.Getenv("host_system_id"),
		os.Getenv("hostname"),
		os.Getenv("dns_servers"),
		os.Getenv("domain_name"),
		os.Getenv("search_domains"),
	)
}


func testAccVSphereHostConfigDNSExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}

		c := testAccProvider.Meta().(*Client).vimClient
		ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer cancel()

		hns, err := hostNetworkSystemFromHostSystemID(c, os.Getenv("host_system_id"))
		if err != nil{
			return fmt.Errorf("error getting host network system: %s", err)
		}

		var hostNetworkProps mo.HostNetworkSystem
		err = hns.Properties(ctx, hns.Reference(), nil, &hostNetworkProps)
		if err != nil {
			fmt.Printf("had an error getting the network system properties: %s", err)
		}

		dns_config := hostNetworkProps.DnsConfig.GetHostDnsConfig()

		if os.Getenv("hostname") != dns_config.HostName {
			return fmt.Errorf("The configured hostname %s does not match the hostname we expected: %s", dns_config.HostName, os.Getenv("hostname"))
		}

		return nil
	}
}


func testAccVSphereHostConfigDNSwithHostname(resource_name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resource_name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", resource_name)
		}

		c := testAccProvider.Meta().(*Client).vimClient
		ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer cancel()

		hns, err := hostNetworkSystemFromHostSystemID(c, os.Getenv("host_system_id"))
		if err != nil{
			return fmt.Errorf("error getting host network system: %s", err)
		}

		var hostNetworkProps mo.HostNetworkSystem
		err = hns.Properties(ctx, hns.Reference(), nil, &hostNetworkProps)
		if err != nil {
			fmt.Printf("had an error getting the network system properties: %s", err)
		}

		dns_config := hostNetworkProps.DnsConfig.GetHostDnsConfig()

		if os.Getenv("hostname") != dns_config.HostName {
			return fmt.Errorf("The configured hostname %s does not match the hostname we expected: %s", dns_config.HostName, os.Getenv("hostname"))
		}

		return nil
	}
}
