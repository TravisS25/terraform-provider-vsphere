package vsphere

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
	"os"
	"fmt"
	"errors"
	"context"
)

func TestAccResourceVSphereLdapIdentitySource_basic(t *testing.T) {
	friendly_name := "test_friendly_name"
	new_friendly_name := "foo_friendly_bar"
	resource_name := "vsphere_ldap_identity_source.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ldap_username", "ldap_password", "domain_name", "domain_alias", "user_base_dn", "group_base_dn", "primary_url"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereLdapIdentitySourceDestroy,
		Steps: []resource.TestStep{
			{
				// create the original testing resource
				Config: testAccResourceVSphereLdapIdentitySourceConfig(friendly_name),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereLdapIdentitySourceExists(resource_name),
				),
			},
			{
				// change the originally created resources friendly name (create a diff and apply an update)
				Config: testAccResourceVSphereLdapIdentitySourceConfig(new_friendly_name),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereLdapIdentitySourceWithFriendlyName(resource_name, new_friendly_name),
				),
			},
			{
				ResourceName: resource_name,
				Config:       testAccResourceVSphereLdapIdentitySourceConfig(new_friendly_name),
				ImportState:       true,
				// because we can't get the passwords from the vcenter API we cannot do ImportStateVerify: true here -- it will always be blank for a 'tf import'
			},
		},
	})
}

func testAccVSphereLdapIdentitySourceDestroy(s *terraform.State) error {
	found := false
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_ldap_identity_source" {
			continue
		}
		found = true

		ssoclient := testAccProvider.Meta().(*Client).ssoClient

		_, err := identitySourceExists(ssoclient, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("the ldap identity source still exists and it should have been destroyed")
		}
		if !errors.Is(err, identitynotfound) {
			return fmt.Errorf("unable to locate the retrieve the identity sources: %s", err)
		}
	}

	if !found {
		return fmt.Errorf("did not find the resource to be destroyed")
	}

	return nil
}

func testAccResourceVSphereLdapIdentitySourceConfig(friendly_name string) string {
	return fmt.Sprintf(
		`
		resource "vsphere_ldap_identity_source" "test" {
		  ldap_username = "%s"
		  ldap_password = "%s"
		  domain_name = "%s"
		  domain_alias = "%s"
		  server_type      = "ActiveDirectory"
		  friendly_name    = "%s"
		  user_base_dn     = "%s"
		  group_base_dn    = "%s"
		  primary_url      = "%s"
		  failover_url     = ""
		}
	`,
		os.Getenv("ldap_username"),
		os.Getenv("ldap_password"),
		os.Getenv("domain_name"),
		os.Getenv("domain_alias"),
                friendly_name,
		os.Getenv("user_base_dn"),
		os.Getenv("group_base_dn"),
		os.Getenv("primary_url"),
 	)
}

func testAccVSphereLdapIdentitySourceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		ssoclient := testAccProvider.Meta().(*Client).ssoClient

		_, err := identitySourceExists(ssoclient, rs.Primary.ID)
		if err != nil {
			if errors.Is(err, identitynotfound) {
				return fmt.Errorf("The identity source that was supposed to be created could not be found")
			} else {
				return fmt.Errorf("There was a problem checking the configured identity sources: %s", err)
			}
		}
		return nil
	}
}

func testAccVSphereLdapIdentitySourceWithFriendlyName(resource_name string, friendly_name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer cancel()
		
		_, ok := s.RootModule().Resources[resource_name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", resource_name)
		}
		ssoclient := testAccProvider.Meta().(*Client).ssoClient

		identitysources, err := ssoclient.IdentitySources(ctx)
		if err != nil {
			return fmt.Errorf("error fetching identity sources: %s\n", err)
		}
		
                for _, value := range identitysources.LDAPS {
			if value.Details.FriendlyName == friendly_name {
				return nil
			}
		}

		return fmt.Errorf("unable to locate a matching identity source with friendly_name: %s", friendly_name)
	}
}
