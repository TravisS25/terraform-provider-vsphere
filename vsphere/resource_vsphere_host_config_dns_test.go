package vsphere

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"

	//"strings"
	"context"
	"os"

	"github.com/vmware/govmomi/vim25/mo"
)

func TestAccResourceVSphereHostConfigDNS_basic(t *testing.T) {
	resourceName := "vsphere_host_config_dns.h1"
	dnsHostname := "testdomain"
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
					"TF_VAR_VSPHERE_ESXI_DNS_HOSTNAME",
					"TF_VAR_VSPHERE_ESXI_DOMAIN_NAME",
					"TF_VAR_VSPHERE_ESXI_DNS_SERVERS",
					"TF_VAR_VSPHERE_ESXI_SEARCH_DOMAINS",
				},
			)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				// create the original testing resource
				Config: testAccResourceVSphereHostConfigDNSConfig(dnsHostname, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostConfigDNSValidate(resourceName, dnsHostname),
				),
			},
			{
				// change the originally created resources hostname (create a diff and apply an update)
				Config: testAccResourceVSphereHostConfigDNSConfig(os.Getenv("TF_VAR_VSPHERE_ESXI_DNS_HOSTNAME"), false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostConfigDNSValidate(resourceName, os.Getenv("TF_VAR_VSPHERE_ESXI_DNS_HOSTNAME")),
				),
			},
			{
				ResourceName:      resourceName,
				Config:            testAccResourceVSphereHostConfigDNSConfig(os.Getenv("TF_VAR_VSPHERE_ESXI_DNS_HOSTNAME"), false),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceVSphereHostConfigDNSConfig(dnsHostname string, useHostname bool) string {
	resourceStr :=
		`
	%s

	resource "vsphere_host_config_dns" "h1" {
		%s
		dns_hostname = "%s"
		domain_name = "%s"
		dns_servers = [%s]
		search_domains = [%s]
	  }
	`

	envDNSServers := strings.Split(os.Getenv("TF_VAR_VSPHERE_ESXI_DNS_SERVERS"), ",")
	dnsStr := ""

	for i, s := range envDNSServers {
		dnsStr += `"` + strings.TrimSpace(s) + `"`

		if i != len(envDNSServers)-1 {
			dnsStr += ", "
		}
	}

	envSearchDomains := strings.Split(os.Getenv("TF_VAR_VSPHERE_ESXI_SEARCH_DOMAINS"), ",")
	searchDomainStr := ""

	for i, s := range envSearchDomains {
		searchDomainStr += `"` + strings.TrimSpace(s) + `"`

		if i != len(envSearchDomains)-1 {
			searchDomainStr += ", "
		}
	}

	if useHostname {
		return fmt.Sprintf(
			resourceStr,
			testhelper.CombineConfigs(
				testhelper.ConfigDataRootDC1(),
				testhelper.ConfigDataRootComputeCluster1(),
				testhelper.ConfigDataRootHost1(),
			),
			"hostname = data.vsphere_host.roothost1.name",
			dnsHostname,
			os.Getenv("TF_VAR_VSPHERE_ESXI_DOMAIN_NAME"),
			dnsStr,
			searchDomainStr,
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
		dnsHostname,
		os.Getenv("TF_VAR_VSPHERE_ESXI_DOMAIN_NAME"),
		dnsStr,
		searchDomainStr,
	)
}

func testAccVSphereHostConfigDNSValidate(name, dnsHostname string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}

		client := testAccProvider.Meta().(*Client).vimClient
		ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer cancel()

		host, _, err := hostsystem.CheckIfHostnameOrID(client, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error retrieving host: %s", err)
		}

		hns, err := hostNetworkSystemFromHostSystemID(client, host.Name())
		if err != nil {
			return fmt.Errorf("error getting host network system: %s", err)
		}

		var hostNetworkProps mo.HostNetworkSystem
		err = hns.Properties(ctx, hns.Reference(), nil, &hostNetworkProps)
		if err != nil {
			fmt.Printf("had an error getting the network system properties: %s", err)
		}

		dnsCfg := hostNetworkProps.DnsConfig.GetHostDnsConfig()

		if dnsHostname != dnsCfg.HostName {
			return fmt.Errorf("The configured hostname %s does not match the hostname we expected: %s", dnsCfg.HostName, os.Getenv("hostname"))
		}

		return nil
	}
}
