package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var testAccDataSourceVSphereIscsiSoftwareAdapterExpectedRegexp = regexp.MustCompile("^host-")

func TestAccDataSourceVSphereIscsiSoftwareAdapter_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(
				t,
				[]string{"TF_VAR_VSPHERE_HOST_SYSTEM_ID"},
			)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereIscsiSoftwareAdapterConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_iscsi_software_adapter.h1",
						"id",
						testAccDataSourceVSphereIscsiSoftwareAdapterExpectedRegexp,
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereIscsiSoftwareAdapterConfig() string {
	return fmt.Sprintf(
		`
		data "vsphere_iscsi_software_adapter" "h1" {
			host_system_id = "%s"
		}
		`,
		os.Getenv("TF_VAR_VSPHERE_HOST_SYSTEM_ID"),
	)
}
