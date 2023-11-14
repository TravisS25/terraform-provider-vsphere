// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"errors"
	"fmt"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/customattribute"
	"github.com/vmware/govmomi/ssoadmin"
	ssoadmin_types "github.com/vmware/govmomi/ssoadmin/types"

	"github.com/vmware/govmomi/ssoadmin/methods"

)

var identitynotfound = errors.New("could not find identity source - this might be expected")

func resourceVSphereLDAPIdentitySource() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereLDAPIdentitySourceCreate,
		Read:   resourceVSphereLDAPIdentitySourceRead,
		Update: resourceVSphereLDAPIdentitySourceUpdate,
		Delete: resourceVSphereLDAPIdentitySourceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVSphereLDAPIdentitySourceImport,
		},
		CustomizeDiff: resourceVSphereLDAPIdentitySourceCustomDiff,
		Schema: map[string]*schema.Schema{
			"ldap_username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ldap_password": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server_type": {
				Type:     schema.TypeString,
				ForceNew: true,
				Default: "ActiveDirectory",
				Optional: true,
			},
			"friendly_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_base_dn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"group_base_dn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"primary_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"failover_url": {
				Type:     schema.TypeString,
				Default: "",
				Optional: true,
			},

			// Add tags schema
			vSphereTagAttributeKey: tagsSchema(),

			// Custom Attributes
			customattribute.ConfigKey: customattribute.ConfigSchema(),
		},
	}
}

func resourceVSphereLDAPIdentitySourceCreate(d *schema.ResourceData, meta interface{}) error {
	ssoclient := meta.(*Client).ssoClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()


	_, err := identitySourceExists(ssoclient, d.Get("domain_name").(string))
	// check if the domain we are about to create already exists (we don't want it to)
	if err == nil {
		return fmt.Errorf("the domain %s already exists", d.Get("domain_name").(string))
	}
	if !errors.Is(err, identitynotfound) {
		return fmt.Errorf("error getting currently configured ldap identity source: %s", err)
	}

	details := ssoadmin_types.LdapIdentitySourceDetails {
		FriendlyName: d.Get("friendly_name").(string),
		UserBaseDn: d.Get("user_base_dn").(string),
		GroupBaseDn: d.Get("group_base_dn").(string),
		PrimaryURL: d.Get("primary_url").(string),
		FailoverURL: d.Get("failover_url").(string),
	}
	auth := ssoadmin_types.SsoAdminIdentitySourceManagementServiceAuthenticationCredentails {
		Username: d.Get("ldap_username").(string),
		Password: d.Get("ldap_password").(string),
	}

	// actually add the LDAP identity source to vcenter
	err = ssoclient.RegisterLdap(ctx, d.Get("server_type").(string), d.Get("domain_name").(string), d.Get("domain_alias").(string), details, auth)
	if err != nil {
		return fmt.Errorf("error registering ldap: %s\n", err.Error())

	}

	// add the resource into the terraform state
	d.SetId(d.Get("domain_name").(string))

	return resourceVSphereLDAPIdentitySourceRead(d, meta)
}

func identitySourceExists(ssoclient *ssoadmin.Client, id string) (*ssoadmin_types.LdapIdentitySource, error) {

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	Myidentitysources, err := ssoclient.IdentitySources(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching identity sources: %s\n", err)
	}

	for _, value := range Myidentitysources.LDAPS {
		// we found it existing
		if value.Name == id {
			return &value, nil
		}
	}

	return nil, identitynotfound
}

// Hopefully this will be replaced by travis building the ssoclient within the provider first and we can remove all the env var hacky stuff here
// func createSSOClient(client *govmomi.Client, vcenter_username string, vcenter_password string) (*ssoadmin.Client, error) {

// 	// if we are doing an import, we do not have access TF variables.tf vars e.g. cannot do a d.Get("var")
// 	// this means the params for vcenter_username and vcenter_password will be blank here (for imports ONLY)
// 	// - for creates, reads, and updates, we let TF check these for us.
// 	if vcenter_username == "" {
// 		if os.Getenv("TF_VAR_vsphere_username") == "" {
// 			return nil, fmt.Errorf("please set your TF_VAR_vsphere_username to a username with administrative access to vcenter and retry")
// 		}
// 		vcenter_username = os.Getenv("TF_VAR_vsphere_username")

// 	}

// 	if vcenter_password == "" {
// 		if os.Getenv("TF_VAR_vsphere_password") == "" {
// 			return nil, fmt.Errorf("please set your TF_VAR_vsphere_password for the TF_VAR_vsphere_username with admin access and retry")
// 		}
// 		vcenter_password = os.Getenv("TF_VAR_vsphere_password")
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
// 	defer cancel()

// 	ssoclient, err := ssoadmin.NewClient(ctx, client.Client)
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating sso client: %s\n", err.Error())
// 	}

// 	tokens, cerr := sts.NewClient(ctx, client.Client)
// 	if cerr != nil {
// 		return nil, fmt.Errorf("error trying to get token: %s", cerr)
// 	}

// 	req := sts.TokenRequest{
// 		Certificate: client.Certificate(),
// 		Userinfo:    url.UserPassword(vcenter_username, vcenter_password),
// 	}

// 	header := soap.Header{
// 		Security: &sts.Signer{
// 			Certificate: client.Certificate(),
// 			//Token:       token,
// 		},
// 	}

// 	header.Security, cerr = tokens.Issue(ctx, req)
// 	if cerr != nil {
// 		return nil, fmt.Errorf("error trying to set security header: %s", cerr)
// 	}

// 	if err = ssoclient.Login(client.WithHeader(ctx, header)); err != nil {
// 		return nil, fmt.Errorf("error trying to login: %s", cerr)
// 	}

// 	return ssoclient, nil
// }


func resourceVSphereLDAPIdentitySourceRead(d *schema.ResourceData, meta interface{}) error {
	ssoclient := meta.(*Client).ssoClient

	// if the user specifies a LDAP source to be created that already exists in vcenter this will fail to be created as there is a name conflict
	identitySource, err := identitySourceExists(ssoclient, d.Id())
	if err != nil {
		return fmt.Errorf("Read func - error checking if existing ldap source exists: %s", err)
	}

	d.Set("domain_name", identitySource.Name)
	d.Set("domain_alias", identitySource.Name)
	d.Set("server_type", identitySource.Type)
	d.Set("friendly_name", identitySource.Details.FriendlyName)
	d.Set("user_base_dn", identitySource.Details.UserBaseDn)
	d.Set("group_base_dn", identitySource.Details.GroupBaseDn)
	d.Set("primary_url", identitySource.Details.PrimaryURL)
	d.Set("failover_url", identitySource.Details.FailoverURL)
	d.Set("ldap_username", identitySource.AuthenticationDetails.Username)
	// we are unable to get the password via the API for this
	d.Set("ldap_password", d.Get("ldap_password"))

	return nil
}

func resourceVSphereLDAPIdentitySourceUpdate(d *schema.ResourceData, meta interface{}) error {
	ssoclient := meta.(*Client).ssoClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()



	_, err := identitySourceExists(ssoclient, d.Get("domain_name").(string))
	if err != nil {
		// check if the domain we are about to create already exists (it should...) and we get no other errors
		if errors.Is(err, identitynotfound) {
			return fmt.Errorf("unable to locate the domain which should exist: %s", err)
		} else {
			return fmt.Errorf("error getting currently configured ldap identity sources for update: %s", err)
		}
	}


	if d.HasChanges("ldap_username", "ldap_password") {
		auth := ssoadmin_types.SsoAdminIdentitySourceManagementServiceAuthenticationCredentails {
			Username: d.Get("ldap_username").(string),
			Password: d.Get("ldap_password").(string),
		}

		err = ssoclient.UpdateLdapAuthnType(ctx, d.Get("domain_name").(string), auth)
		if err != nil {
			return fmt.Errorf("error updating ldap username or password: %s", err)
		}
	}

	if d.HasChanges("friendly_name", "user_base_dn", "group_base_dn", "primary_url", "secondary_url") {
		details := ssoadmin_types.LdapIdentitySourceDetails {
			FriendlyName: d.Get("friendly_name").(string),
			UserBaseDn: d.Get("user_base_dn").(string),
			GroupBaseDn: d.Get("group_base_dn").(string),
			PrimaryURL: d.Get("primary_url").(string),
			FailoverURL: d.Get("failover_url").(string),
		}

		err = ssoclient.UpdateLdap(ctx, d.Get("domain_name").(string), details)

		if err != nil {
			return fmt.Errorf("error updating ldap details such as friendly name: %s", err)
		}
	}

	return nil
}

func resourceVSphereLDAPIdentitySourceDelete(d *schema.ResourceData, meta interface{}) error {
	ssoclient := meta.(*Client).ssoClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()


	_, err := identitySourceExists(ssoclient, d.Get("domain_name").(string))
	if err != nil {
		// check if the domain we are about to create already exists (it should...) and we get no other errors
		if errors.Is(err, identitynotfound) {
			return fmt.Errorf("unable to locate the domain which should exist: %s", err)
		} else {
			return fmt.Errorf("error getting currently configured ldap identity sources for update: %s", err)
		}
	}

	a := ssoadmin_types.DeleteDomain {
		This: ssoclient.ServiceContent.DomainManagementService,
		Name: d.Get("domain_name").(string),
	}

	_, err = methods.DeleteDomain(ctx, ssoclient, &a)
	if err != nil {
		return fmt.Errorf("error deleting ldap identity source: %s", err)
	}
	return nil
}

// NOTE: This import will create the resource within state successfully but the next 'terraform apply' WILL note some changes for it, even if there is nothing actually changing
// this is due to our inability to fetch the currently configured passwords that LDAP is using and TF will enforce the ones defined in it.
func resourceVSphereLDAPIdentitySourceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	ssoclient := meta.(*Client).ssoClient

	// sanity check that the identity source actually exists in vcenter
	_, err := identitySourceExists(ssoclient, d.Id())
	// throw error if it does NOT exist or issue getting data via API
	if err != nil {
		return nil, fmt.Errorf("Import func - error checking if identity source exists: %s\n", err)
	}

	// If no errors verifying that the identity source does exist
	d.SetId(d.Id())

	return []*schema.ResourceData{d}, nil

}


// this function sanity checks that the domain you are trying to create / update with terraform is not going to create an error when you run a 'terraform apply'
// - e.g. this function attempts to catch errors in a 'terraform plan'
func resourceVSphereLDAPIdentitySourceCustomDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// If the LDAP identity source does NOT exist in state yet...
	if d.Id() == "" {
		ssoclient := meta.(*Client).ssoClient

		// check to see if the identitysource exists - this is what alerts you to a possible issue via 'terraform plan' instead of the 'plan' saying all is good and the 'apply' actually failing
		_, err := identitySourceExists(ssoclient, d.Get("domain_name").(string))

		if err == nil {
			return fmt.Errorf("the input domain: %s already exists - considering running a 'terraform import'!", d.Get("domain_name").(string))
		}
		if !errors.Is(err, identitynotfound) {
			return fmt.Errorf("error getting currently configured ldap identity sources: %s", err)
		}
	}

	return nil
}
