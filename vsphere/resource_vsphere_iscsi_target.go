package vsphere

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/customattribute"
)

func resourceVSphereIscsiTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereIscsiTargetCreate,
		Read:   resourceVSphereIscsiTargetRead,
		Update: resourceVSphereIscsiTargetUpdate,
		Delete: resourceVSphereIscsiTargetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVSphereIscsiTargetImport,
		},

		Schema: map[string]*schema.Schema{
			"host_system_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the host system to attach iscsi adapter to",
			},
			"discovery_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "dynamic",
				Description:  "Determines what type of iscsi to create.  Valid options are 'dynamic' and 'static'",
				ValidateFunc: validation.StringInSlice([]string{"dynamic", "static"}, true),
			},
			"adapter_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "software",
				Description: "Iscsi adapter the iscsi targets will be added to",
			},
			"target": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "IP of the iscsi target",
							ValidateFunc: validation.IsCIDR,
						},
						"port": {
							Type:         schema.TypeInt,
							Default:      3260,
							Description:  "Port of the iscsi target",
							ValidateFunc: validation.IsPortNumber,
						},
						"target_name": {
							Type:        schema.TypeString,
							Description: "The iqn of the storage device if iscsi type is 'static'",
						},
						"chap": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Description: "The chap credentials for iscis devices",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"method": {
										Type:         schema.TypeString,
										Default:      "unidirectional",
										Description:  "Chap auth method.  Valid options are 'unidirectional' and 'bidirectional'",
										ValidateFunc: validation.StringInSlice([]string{"unidirectional", "bidirectional"}, true),
									},
									"outgoing_creds": {
										Type:        schema.TypeList,
										Required:    true,
										MaxItems:    1,
										Description: "Outgoing creds for iscsi device",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"username": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Username to auth against iscsi device",
												},
												"password": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Password to auth against iscsi device",
													Sensitive:   true,
												},
											},
										},
									},
									"incoming_creds": {
										Type: schema.TypeList,
										//Required:    true,
										MaxItems:    1,
										Description: "Incoming creds for iscsi device to auth host",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"username": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Username to auth against host",
												},
												"password": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Password to auth against host",
													Sensitive:   true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},

			// Add tags schema
			vSphereTagAttributeKey: tagsSchema(),

			// Custom Attributes
			customattribute.ConfigKey: customattribute.ConfigSchema(),
		},
	}
}

func resourceVSphereIscsiTargetCreate(d *schema.ResourceData, meta interface{}) error {
	//client := meta.(*Client).vimClient

	return nil
}

func resourceVSphereIscsiTargetRead(d *schema.ResourceData, meta interface{}) error {
	//client := meta.(*Client).vimClient

	return nil
}

func resourceVSphereIscsiTargetUpdate(d *schema.ResourceData, meta interface{}) error {
	//client := meta.(*Client).vimClient

	return nil
}

func resourceVSphereIscsiTargetDelete(d *schema.ResourceData, meta interface{}) error {

	return nil
}

func resourceVSphereIscsiTargetImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	//client := meta.(*Client).vimClient

	return []*schema.ResourceData{d}, nil
}
