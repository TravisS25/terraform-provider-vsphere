// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	//"errors"
	"fmt"
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/customattribute"
	"github.com/vmware/govmomi/ssoadmin"
	ssoadmin_types "github.com/vmware/govmomi/ssoadmin/types"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/sts"

	"net/url"
	"os"
	"github.com/vmware/govmomi/vim25/soap"

	"github.com/vmware/govmomi/ssoadmin/methods"
	//"log"

)

// var identitynotfound = errors.New("could not find identity source - this might be expected")

func resourceVSphereLDAPGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereLDAPGroupCreate,
		Read:   resourceVSphereLDAPGroupRead,
		Update: resourceVSphereLDAPGroupUpdate,
		Delete: resourceVSphereLDAPGroupDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVSphereLDAPGroupImport,
		},
		CustomizeDiff: resourceVSphereLDAPGroupCustomDiff,
		Schema: map[string]*schema.Schema{
			"ldap_group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vsphere_group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// these 2 params (username and password) will get removed when
			// travis makes the client for us in provider.go
			"vsphere_username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vsphere_password": {
				Type:     schema.TypeString,
				Required: true,
				Sensitive: true,
			},

			// Add tags schema
			vSphereTagAttributeKey: tagsSchema(),

			// Custom Attributes
			customattribute.ConfigKey: customattribute.ConfigSchema(),
		},
	}
}

// function for sanity checking the passed in vsphere_group actually exists in vsphere (this resource does NOT create vsphere_group(s))
func vsphereGroupExists(ssoclient *ssoadmin.Client, group_name string) error {

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	group, err := ssoclient.FindGroup(ctx, group_name)
	if err != nil {
		return fmt.Errorf("error fetching groups: %s\n", err)
	}

	// if the group was not found in vsphere
	if group == nil {
		return fmt.Errorf("unable to locate group: %s\n", group_name)
	}

	// At this point we can assume the group was found because we have not returned yet
	return nil
}

// function which sanity checks the vsphere_group exists AND checks if the given ldap_group is a member of the vsphere_group already
func ldapGroupInVsphereGroupCheck(ssoclient *ssoadmin.Client, vsphere_group string, ldap_group string) (*ssoadmin_types.AdminGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	err := vsphereGroupExists(ssoclient, vsphere_group)
	if err != nil {
		return nil, fmt.Errorf("error in ldapGroupInVsphereGroupCheck func - the vsphere group %s does not exist", vsphere_group)
	}

	// NOTE: This function accepts a 'search' string as the last param but it does not seem to do anything.
	// It seems to always return the entire array of groups within the group you are searching in.
	groups_in_group, err := ssoclient.FindGroupsInGroup(ctx, vsphere_group, ldap_group)
	if err != nil {
		return nil, fmt.Errorf("error locating groups in group. error: %s\n", err)
	}

	// if we returned any groups at all..
	if groups_in_group != nil {

		// iterate the array of returned groups
		for _, value := range groups_in_group {
			// If we find the group we are trying to add already within the group..
			if value.Id.Name == ldap_group {
				// if the ldap_group is already a member of the vsphere_group
				return &value, nil
			}
		}
	} 
	// if ldap_group is not currently a member of the vsphere_group...
	return nil, nil
}

func resourceVSphereLDAPGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	ssoclient, err := createSSOClientForLDAPGroups(client, d.Get("vsphere_username").(string), d.Get("vsphere_password").(string))
	if err != nil {
		return fmt.Errorf("error in create function creating ssoclient: %s", err)
	}

	err = vsphereGroupExists(ssoclient, d.Get("vsphere_group").(string))
	if err != nil {
		return fmt.Errorf("error in create func - the vsphere group %s does not exist", d.Get("vsphere_group").(string))
	}

	group_to_add := ssoadmin_types.PrincipalId {
		Name: d.Get("ldap_group").(string),
		Domain: d.Get("domain_name").(string),
	}

	// actually add the ldap_group to the vsphere_group
	err = ssoclient.AddGroupsToGroup(ctx, d.Get("vsphere_group").(string), group_to_add)
	if err != nil {
		return fmt.Errorf("error adding ldap_group to the vsphere_group: %s\n", err.Error())

	}

	// add the resource into the terraform state
	d.SetId(d.Get("ldap_group").(string))

	return resourceVSphereLDAPGroupRead(d, meta)
}

// Hopefully this will be replaced by travis building the ssoclient within the provider first and we can remove all the env var hacky stuff here
func createSSOClientForLDAPGroups(client *govmomi.Client, vcenter_username string, vcenter_password string) (*ssoadmin.Client, error) {

	// if we are doing an import, we do not have access TF variables.tf vars e.g. cannot do a d.Get("var")
	// this means the params for vcenter_username and vcenter_password will be blank here (for imports ONLY)
	// - for creates, reads, and updates, we let TF check these for us.
	if vcenter_username == "" {
		if os.Getenv("TF_VAR_vsphere_username") == "" {
			return nil, fmt.Errorf("please set your TF_VAR_vsphere_username to a username with administrative access to vcenter and retry")
		}
		vcenter_username = os.Getenv("TF_VAR_vsphere_username")

	}

	if vcenter_password == "" {
		if os.Getenv("TF_VAR_vsphere_password") == "" {
			return nil, fmt.Errorf("please set your TF_VAR_vsphere_password for the TF_VAR_vsphere_username with admin access and retry")
		}
		vcenter_password = os.Getenv("TF_VAR_vsphere_password")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	ssoclient, err := ssoadmin.NewClient(ctx, client.Client)
	if err != nil {
		return nil, fmt.Errorf("error creating sso client: %s\n", err.Error())
	}

	tokens, cerr := sts.NewClient(ctx, client.Client)
	if cerr != nil {
		return nil, fmt.Errorf("error trying to get token: %s", cerr)
	}

	req := sts.TokenRequest{
		Certificate: client.Certificate(),
		Userinfo:    url.UserPassword(vcenter_username, vcenter_password),
	}

	header := soap.Header{
		Security: &sts.Signer{
			Certificate: client.Certificate(),
			//Token:       token,
		},
	}

	header.Security, cerr = tokens.Issue(ctx, req)
	if cerr != nil {
		return nil, fmt.Errorf("error trying to set security header: %s", cerr)
	}

	if err = ssoclient.Login(client.WithHeader(ctx, header)); err != nil {
		return nil, fmt.Errorf("error trying to login: %s", cerr)
	}

	return ssoclient, nil
}


func resourceVSphereLDAPGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	ssoclient, err := createSSOClientForLDAPGroups(client, d.Get("vsphere_username").(string), d.Get("vsphere_password").(string))
	if err != nil {
		return fmt.Errorf("error creating ssoclient: %s", err)
	}

	group, err := ldapGroupInVsphereGroupCheck(ssoclient, d.Get("vsphere_group").(string), d.Get("ldap_group").(string))
	if err != nil {
		return fmt.Errorf("Read func - error checking if ldap_group '%s' is a member of vsphere_group '%s': %s", d.Get("ldap_group").(string), d.Get("vsphere_group").(string), err)
	}
	// If the ldap_group is already a member of the vsphere_group
	if group != nil {
		d.Set("ldap_group", group.Id.Name)
		d.Set("domain_name", group.Id.Domain)
	}
	// TRAVIS Q - Do i need an 'else' above? we only run d.Set if the ldap_group is already a member of the vsphere_group - that's what we want in the Read func, right?

	// TRAVIS Q
	// is there any value in us fetching this from the API?? - the ldapGroupInVsphereGroupCheck is already validating this vsphere_group exists
	// it is also validating that our ldap_group is a member of that group
	d.Set("vsphere_group", d.Get("vsphere_group"))

	return nil
}

func resourceVSphereLDAPGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	ssoclient, err := createSSOClientForLDAPGroups(client, d.Get("vsphere_username").(string), d.Get("vsphere_password").(string))
	if err != nil {
		return fmt.Errorf("error in create function creating ssoclient: %s", err)
	}

	_, err = ldapGroupInVsphereGroupCheck(ssoclient, d.Get("vsphere_group").(string), d.Get("ldap_group").(string))
	if err != nil {
		return fmt.Errorf("update func - error checking if ldap_group '%s' is a member of vsphere_group '%s': %s", d.Get("ldap_group").(string), d.Get("vsphere_group").(string), err)
	}

	// TRAVIS Q: if all my params are 'force_new' does the update function do anything??
	// if d.HasChanges("ldap_group", "ldap_password") {
	// 	auth := ssoadmin_types.SsoAdminIdentitySourceManagementServiceAuthenticationCredentails {
	// 		Username: d.Get("ldap_username").(string),
	// 		Password: d.Get("ldap_password").(string),
	// 	}

	// 	err = ssoclient.UpdateLdapAuthnType(ctx, d.Get("domain_name").(string), auth)
	// 	if err != nil {
	// 		return fmt.Errorf("error updating ldap username or password: %s", err)
	// 	}
	// }

	// if d.HasChanges("friendly_name", "user_base_dn", "group_base_dn", "primary_url", "secondary_url") {
	// 	details := ssoadmin_types.LdapIdentitySourceDetails {
	// 		FriendlyName: d.Get("friendly_name").(string),
	// 		UserBaseDn: d.Get("user_base_dn").(string),
	// 		GroupBaseDn: d.Get("group_base_dn").(string),
	// 		PrimaryURL: d.Get("primary_url").(string),
	// 		FailoverURL: d.Get("failover_url").(string),
	// 	}

	// 	err = ssoclient.UpdateLdap(ctx, d.Get("domain_name").(string), details)

	// 	if err != nil {
	// 		return fmt.Errorf("error updating ldap details such as friendly name: %s", err)
	// 	}
	// }

	return nil
}

func resourceVSphereLDAPGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	ssoclient, err := createSSOClientForLDAPGroups(client, d.Get("vsphere_username").(string), d.Get("vsphere_password").(string))
	if err != nil {
		return fmt.Errorf("error in delete function creating ssoclient: %s", err)
	}

	err = vsphereGroupExists(ssoclient, d.Get("vsphere_group").(string))
	if err != nil {
		return fmt.Errorf("Error in delete func - unable to locate vsphere group: %s", err)
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
func resourceVSphereLDAPGroupImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	client := meta.(*Client).vimClient

	// normally we would use d.Get("vsphere_username").(string), d.Get("vsphere_password").(string) here but the values are not accessible via an import
	// so they will ALWAYS be blank strings
	ssoclient, err := createSSOClientForLDAPGroups(client, "", "")
	if err != nil {
		return nil, fmt.Errorf("error creating ssoclient: %s\n", err)
	}

        // need to think this over a bit -- the Id is gonna be the ldap_group not the vsphere_group..
	err = vsphereGroupExists(ssoclient, d.Id())
	// throw error if it does NOT exist or issue getting data via API
	if err != nil {
		return nil, fmt.Errorf("Import func - error checking if vsphere group exists: %s\n", err)
	}

	// If no errors verifying that the identity source does exist
	d.SetId(d.Id())

	return []*schema.ResourceData{d}, nil

}


// this function sanity checks that the ldap_group you are trying to add to the vsphere_group with terraform is not going to create an error when you run a 'terraform apply'
// - e.g. this function attempts to catch errors in a 'terraform plan'
func resourceVSphereLDAPGroupCustomDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// If the LDAP group does NOT exist in state yet...
	if d.Id() == "" {
		client := meta.(*Client).vimClient
		_, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer cancel()

		ssoclient, err := createSSOClientForLDAPGroups(client, d.Get("vsphere_username").(string), d.Get("vsphere_password").(string))
		if err != nil {
			return fmt.Errorf("error in custom diff function creating ssoclient: %s", err)
		}

		// check to see if the vsphere_group exists and if the ldap_group is already a member of the vsphere_group 
		// - this is what alerts you to a possible issue via 'terraform plan' instead of the 'plan' saying all is good and the 'apply' actually failing
		group, err := ldapGroupInVsphereGroupCheck(ssoclient, d.Get("vsphere_group").(string), d.Get("ldap_group").(string))
		if err != nil {
			return fmt.Errorf("error in custom diff func checking if ldap_group already in vsphere_group: %s", err)
		}
		if group != nil {
			return fmt.Errorf("error: ldap_group: %s is already a member of vsphere_group %s - consider a 'terraform import'!", d.Get("ldap_group").(string), d.Get("vsphere_group").(string))
		}
	}

	return nil
}
