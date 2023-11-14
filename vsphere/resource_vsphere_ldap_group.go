// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"context"
	"strings"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/ssoadmin"
	ssoadmin_types "github.com/vmware/govmomi/ssoadmin/types"
	"github.com/vmware/govmomi/ssoadmin/methods"
)

// var identitynotfound = errors.New("could not find identity source - this might be expected")

func resourceVSphereLDAPGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereLDAPGroupCreate,
		Read:   resourceVSphereLDAPGroupRead,
		// Note: Due to all params on this resource being marked as ForceNew there is no need for an update function. If params change it simply runs a 'Delete' then a 'Create'
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
	ssoclient := meta.(*Client).ssoClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	err := vsphereGroupExists(ssoclient, d.Get("vsphere_group").(string))
	if err != nil {
		return fmt.Errorf("error in create func - the vsphere group %s does not exist", d.Get("vsphere_group").(string))
	}

	group_to_add := ssoadmin_types.PrincipalId {
		Name: d.Get("ldap_group").(string),
		Domain: d.Get("domain_name").(string),
	}

	// add the ldap_group to the vsphere_group
	err = ssoclient.AddGroupsToGroup(ctx, d.Get("vsphere_group").(string), group_to_add)
	if err != nil {
		return fmt.Errorf("error adding ldap_group to the vsphere_group: %s\n", err.Error())

	}

	// add the resource into the terraform state
	id := d.Get("vsphere_group").(string) + ":" + d.Get("ldap_group").(string)
	d.SetId(id)

	return resourceVSphereLDAPGroupRead(d, meta)
}

func resourceVSphereLDAPGroupRead(d *schema.ResourceData, meta interface{}) error {
	ssoclient := meta.(*Client).ssoClient
	
	group, err := ldapGroupInVsphereGroupCheck(ssoclient, d.Get("vsphere_group").(string), d.Get("ldap_group").(string))
	if err != nil {
		return fmt.Errorf("Read func - error checking if ldap_group '%s' is a member of vsphere_group '%s': %s", d.Get("ldap_group").(string), d.Get("vsphere_group").(string), err)
	}
	// If the ldap_group is already a member of the vsphere_group
	if group != nil {
		d.Set("ldap_group", group.Id.Name)
		d.Set("domain_name", group.Id.Domain)

	// this 'else' block applies only to TF imports - it sanity checks the ldap_group passed in IS actually a member of the given vsphere_group 
	} else {
		return fmt.Errorf("Read func - ldap_group '%s' is a member of vsphere_group '%s'", d.Get("ldap_group").(string), d.Get("vsphere_group").(string))
	}

	return nil
}

func resourceVSphereLDAPGroupDelete(d *schema.ResourceData, meta interface{}) error {
	ssoclient := meta.(*Client).ssoClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	err := vsphereGroupExists(ssoclient, d.Get("vsphere_group").(string))
	if err != nil {
		return fmt.Errorf("Error in delete func - unable to locate vsphere group: %s", err)
	}

	group_to_remove := []ssoadmin_types.PrincipalId {
		{
			Name: d.Get("ldap_group").(string),
			Domain: d.Get("domain_name").(string),
		},
	}

	a := ssoadmin_types.RemovePrincipalsFromLocalGroup {
		This: ssoclient.ServiceContent.PrincipalManagementService,
		PrincipalsIds: group_to_remove,
		GroupName: d.Get("vsphere_group").(string),	
	}

	// remove the ldap_group from the vsphere_group
	_, err = methods.RemovePrincipalsFromLocalGroup(ctx, ssoclient, &a)
	if err != nil {
		return fmt.Errorf("error removing ldap_group %s from vsphere_group %s. Error: %s", d.Get("ldap_group").(string), d.Get("vsphere_group").(string), err)
	}
	return nil
}

// NOTE: This import will create the resource within state successfully but the next 'terraform apply' WILL note some changes for it, even if there is nothing actually changing
// this is due to our inability to fetch the currently configured passwords that LDAP is using and TF will enforce the ones defined in it.
func resourceVSphereLDAPGroupImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	d.SetId(d.Id())

	error_msg := "Invalid Id format given. Only 2 strings seperated by a ':' is allowed. Proper format is: 'vsphere_group_name:ldap_group_name'"

	// ID format we want should be 'vsphere_group:ldap_group'
	id_split := strings.Split(d.Id(), ":")
	if len(id_split) != 2 {
		return nil, fmt.Errorf("%s", error_msg)
	}
	if id_split[0] == "" || id_split[1] == "" {
		return nil, fmt.Errorf("%s", error_msg)
	}

	d.Set("vsphere_group", id_split[0])
	d.Set("ldap_group", id_split[1])

	return []*schema.ResourceData{d}, nil
}


// this function sanity checks that the ldap_group you are trying to add to the vsphere_group with terraform is not going to create an error when you run a 'terraform apply'
// - e.g. this function attempts to catch errors in a 'terraform plan'
func resourceVSphereLDAPGroupCustomDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// If the LDAP group does NOT exist in state yet...
	if d.Id() == "" {
		ssoclient := meta.(*Client).ssoClient

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
