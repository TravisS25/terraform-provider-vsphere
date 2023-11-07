// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"errors"
	"fmt"
	"log"
	// "strings"
	// "time"

	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/customattribute"
	// "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/datacenter"
	// "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/folder"
	"github.com/vmware/govmomi/find"
	// "github.com/vmware/govmomi/object"
	// "github.com/vmware/govmomi/vim25/methods"
	// "github.com/vmware/govmomi/vim25/types"

	"github.com/vmware/govmomi/ssoadmin"
	ssoadmin_types "github.com/vmware/govmomi/ssoadmin/types"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/sts"

	"net/url"
	"os"
	"github.com/vmware/govmomi/vim25/soap"

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
	client := meta.(*Client).vimClient
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	ssoclient, err := createSSOClient(client)
	if err != nil {
		return fmt.Errorf("error in create function creating ssoclient: %s", err)
	}

	_, err = identitySourceExists(ssoclient, d.Get("domain_name").(string))
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

func resourceVSphereLDAPIdentitySourceStateRefreshFunc(d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Print("[DEBUG] Refreshing datacenter state")
		dc, err := datacenterExists(d, meta)
		if err != nil {
			switch err.(type) {
			case *find.NotFoundError:
				log.Printf("[DEBUG] Refreshing state. LDAPIdentitySource not found: %s", err)
				return nil, "InProgress", nil
			default:
				return nil, "Failed", err
			}
		}
		log.Print("[DEBUG] Refreshing state. LDAPIdentitySource found")
		return dc, "Created", nil
	}
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

//// NEW CODE BEGINS
func createSSOClient(client *govmomi.Client) (*ssoadmin.Client, error) {

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()

	// CODE FROM MY GO SCRIPT BEGINS

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
		Userinfo:    url.UserPassword( os.Getenv("vcenter_username"),  os.Getenv("vcenter_password")),
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


func resourceVSphereLDAPIdentitySourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	ssoclient, err := createSSOClient(client)
	if err != nil {
		return fmt.Errorf("error creating ssoclient: %s", err)
	}

	// if the user specifies a LDAP source to be created that already exists in vcenter this will fail to be created as there is a name conflict
	// You will not know this is going to error via a 'terraform plan' and it will only occur during a 'terraform apply'
	identitySource, err := identitySourceExists(ssoclient, d.Id())

	if err != nil {
		return fmt.Errorf("error checking if existing ldap source exists: %s", err)
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
	// // Load up the tags client, which will validate a proper vCenter before
	// // attempting to proceed if we have tags defined.
	// tagsClient, err := tagsManagerIfDefined(d, meta)
	// if err != nil {
	//	return err
	// }
	// // Verify a proper vCenter before proceeding if custom attributes are defined
	// client := meta.(*Client).vimClient
	// attrsProcessor, err := customattribute.GetDiffProcessorIfAttributesDefined(client, d)
	// if err != nil {
	//	return err
	// }

	// dc, err := datacenterExists(d, meta)
	// if err != nil {
	//	return fmt.Errorf("couldn't find the specified datacenter: %s", err)
	// }

	// // Apply any pending tags now
	// if tagsClient != nil {
	//	if err := processTagDiff(tagsClient, d, dc); err != nil {
	//		return err
	//	}
	// }

	// // Set custom attributes
	// if attrsProcessor != nil {
	//	if err := attrsProcessor.ProcessDiff(dc); err != nil {
	//		return err
	//	}
	// }

	return nil
}

func resourceVSphereLDAPIdentitySourceDelete(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*Client).vimClient
	// name := d.Get("name").(string)

	// path := name
	// if v, ok := d.GetOk("folder"); ok {
	//	path = v.(string) + "/" + name
	// }

	// finder := find.NewFinder(client.Client, true)
	// dc, err := finder.LDAPIdentitySource(context.TODO(), path)
	// if err != nil {
	//	log.Printf("couldn't find the specified datacenter: %s", err)
	//	d.SetId("")
	//	return nil
	// }

	// req := &types.Destroy_Task{
	//	This: dc.Common.Reference(),
	// }

	// _, err = methods.Destroy_Task(context.TODO(), client, req)
	// if err != nil {
	//	return fmt.Errorf("%s", err)
	// }

	// // Wait for the datacenter resource to be destroyed
	// stateConf := &resource.StateChangeConf{
	//	Pending:    []string{"Created"},
	//	Target:     []string{},
	//	Refresh:    resourceVSphereLDAPIdentitySourceStateRefreshFunc(d, meta),
	//	Timeout:    10 * time.Minute,
	//	MinTimeout: 3 * time.Second,
	//	Delay:      5 * time.Second,
	// }

	// _, err = stateConf.WaitForState()
	// if err != nil {
	//	return fmt.Errorf("error waiting for datacenter (%s) to become ready: %s", name, err)
	// }

	return nil
}

func resourceVSphereLDAPIdentitySourceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	client := meta.(*Client).vimClient

	ssoclient, err := createSSOClient(client)
	if err != nil {
		return nil, fmt.Errorf("error creating ssoclient: %s\n", err)
	}

	// sanity check that the identity source actually exists in vcenter
	_, err = identitySourceExists(ssoclient, d.Id())
	// throw error if it does NOT exist or issue getting data via API
	if err != nil {
		return nil, fmt.Errorf("error checking if identity source exists: %s\n", err)
	}

	// If no errors verifying that the identity source does exist
	d.SetId(d.Id())

	return []*schema.ResourceData{d}, nil

	// client := meta.(*Client).vimClient
	// p := d.Id()
	// if !strings.HasPrefix(p, "/") {
	//	return nil, errors.New("path must start with a trailing slash")
	// }

	// dc, err := datacenter.FromPath(client, p)
	// if err != nil {
	//	return nil, err
	// }

	// // determine a folder if one is present
	// f, err := folder.ParentFromPath(client, p, folder.VSphereFolderTypeLDAPIdentitySource, dc)
	// if err != nil {
	//	return nil, fmt.Errorf("cannot locate folder: %s", err)
	// }

	// path := strings.TrimPrefix(f.InventoryPath, "/")
	// if path != "" {
	//	if err := d.Set("folder", path); err != nil {
	//		return nil, err
	//	}
	// }

	// d.SetId(dc.Name())
	// return []*schema.ResourceData{d}, nil
}
