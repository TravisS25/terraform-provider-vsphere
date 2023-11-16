// NOTE: This test relies on the LDAP identity source (different TF resource) being created and present within vsphere in order to run. Without it this test will fail.
package vsphere

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
	"fmt"
	"strings"
	"os"
)

func TestAccResourceVSphereLdapGroup_basic(t *testing.T) {
	resource_name := "vsphere_ldap_group.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ldap_group", "vsphere_group",  "domain_name"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereLdapGroupDestroy,
		Steps: []resource.TestStep{
			{
				// create the original testing resource
				Config: testAccResourceVSphereLdapGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereLdapGroupExists(resource_name),
				),
			},
			{
				ResourceName: resource_name,
				Config:       testAccResourceVSphereLdapGroupConfig(),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVSphereLdapGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_ldap_group" {
			continue
		}

		ssoclient := testAccProvider.Meta().(*Client).ssoClient

		id_split := strings.Split(rs.Primary.ID, ":")
		
		group, err := ldapGroupInVsphereGroupCheck(ssoclient, id_split[0], id_split[1])
		if err != nil {
			return fmt.Errorf("there was an error in the test destroy func: %s", err)
		}
		if group != nil {
			return fmt.Errorf("the ldap_group is still a member of the vsphere_group but should have been removed during destroy action")
		}
	}

	return nil
}

func testAccResourceVSphereLdapGroupConfig() string {
	return fmt.Sprintf(
		`
		resource "vsphere_ldap_group" "test" {
		  ldap_group    = "%s"
		  vsphere_group = "%s"
		  domain_name   = "%s"
		}
	`,
		os.Getenv("ldap_group"),
		os.Getenv("vsphere_group"),
		os.Getenv("domain_name"),
	)
}


func testAccVSphereLdapGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		ssoclient := testAccProvider.Meta().(*Client).ssoClient
		id_split := strings.Split(rs.Primary.ID, ":")
		
		group, err := ldapGroupInVsphereGroupCheck(ssoclient, id_split[0], id_split[1])
		if err != nil {
			return fmt.Errorf("Error checking the group membership of %s: %s", id_split[0], err)
		}
		if group == nil {
			return fmt.Errorf("The vsphere group: %s was supposed to contain ldap group: %s and did not", id_split[0], id_split[1])
		}
		return nil
	}
}

