// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/iscsi"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccResourceVSphereIscsiSoftwareAdapter_basic(t *testing.T) {
	testIscsiName := "iqn.1998-01.com.testacc"
	newTestIscsiName := testIscsiName + ".new"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariablesF(
				t,
				[]string{"TF_VAR_VSPHERE_DATACENTER", "TF_VAR_VSPHERE_CLUSTER", "TF_VAR_VSPHERE_ESXI1"},
			)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereIscsiSoftwareAdapterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereIscsiSoftwareAdapterConfig(testIscsiName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereIscsiSoftwareAdapterValidation("vsphere_iscsi_software_adapter.h1", testIscsiName),
				),
			},
			{
				Config: testAccResourceVSphereIscsiSoftwareAdapterConfig(newTestIscsiName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereIscsiSoftwareAdapterValidation("vsphere_iscsi_software_adapter.h1", newTestIscsiName),
				),
			},
			{
				ResourceName: "vsphere_iscsi_software_adapter.h1",
				Config:       testAccResourceVSphereIscsiSoftwareAdapterConfig(newTestIscsiName, false),
				ImportState:  true,
			},
		},
	})
}

func TestAccResourceVSphereIscsiSoftwareAdapter_hostname(t *testing.T) {
	testIscsiName := "iqn.1998-01.com.testacc"
	newTestIscsiName := testIscsiName + ".new"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariablesF(
				t,
				[]string{"TF_VAR_VSPHERE_DATACENTER", "TF_VAR_VSPHERE_CLUSTER", "TF_VAR_VSPHERE_ESXI1"},
			)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereIscsiSoftwareAdapterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereIscsiSoftwareAdapterConfig(testIscsiName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereIscsiSoftwareAdapterValidation("vsphere_iscsi_software_adapter.h1", testIscsiName),
				),
			},
			{
				Config: testAccResourceVSphereIscsiSoftwareAdapterConfig(newTestIscsiName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereIscsiSoftwareAdapterValidation("vsphere_iscsi_software_adapter.h1", newTestIscsiName),
				),
			},
			{
				ResourceName: "vsphere_iscsi_software_adapter.h1",
				Config:       testAccResourceVSphereIscsiSoftwareAdapterConfig(newTestIscsiName, true),
				ImportState:  true,
			},
		},
	})
}

func testAccVSphereIscsiSoftwareAdapterDestroy(s *terraform.State) error {
	message := ""
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_host" {
			continue
		}
		client := testAccProvider.Meta().(*Client).vimClient
		idSplit := strings.Split(rs.Primary.ID, ":")
		host, _, err := hostsystem.CheckIfHostnameOrID(client, idSplit[0])
		if err != nil {
			return fmt.Errorf("error retrieving host for iscsi: %s", err)
		}
		hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, host)
		if err != nil {
			return err
		}

		if _, err = iscsi.GetIscsiSoftwareAdater(hssProps, host.Name()); err == nil {
			message = "iscsi software adapter still exists/enabled"
		}
	}
	if message != "" {
		return errors.New(message)
	}
	return nil
}

func testAccVSphereIscsiSoftwareAdapterValidation(resourceName, iscsiName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%s key not found on the server", resourceName)
		}
		idSplit := strings.Split(rs.Primary.ID, ":")
		client := testAccProvider.Meta().(*Client).vimClient
		host, _, err := hostsystem.CheckIfHostnameOrID(client, idSplit[0])
		if err != nil {
			return fmt.Errorf("error retrieving host for iscsi: %s", err)
		}
		hssProps, err := hostsystem.GetHostStorageSystemPropertiesFromHost(client, host)
		if err != nil {
			return err
		}

		adapter, err := iscsi.GetIscsiSoftwareAdater(hssProps, host.Name())
		if err != nil {
			return err
		}

		if adapter.IScsiName != iscsiName {
			return fmt.Errorf(
				"iscsi adapter name invalid.  current value: %s; expected value: %s",
				adapter.IScsiName,
				host.Name(),
			)
		}

		return nil
	}
}

func testAccResourceVSphereIscsiSoftwareAdapterConfig(iscsiName string, useHostname bool) string {
	resourceStr :=
		`
	%s

	resource "vsphere_iscsi_software_adapter" "h1" {
		%s
		iscsi_name = "%s"
	}
	`

	if useHostname {
		return fmt.Sprintf(
			resourceStr,
			testhelper.CombineConfigs(
				testhelper.ConfigDataRootDC1(),
				testhelper.ConfigDataRootComputeCluster1(),
				testhelper.ConfigDataRootHost1(),
			),
			"hostname = data.vsphere_host.roothost1.name",
			iscsiName,
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
		iscsiName,
	)
}
