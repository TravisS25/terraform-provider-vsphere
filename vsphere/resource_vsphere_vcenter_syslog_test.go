// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/govmomi/vapi/appliance/logging"
)

func TestAccResourceVSphereVcenterSyslog_basic(t *testing.T) {
	resourceName := "vsphere_vcenter_syslog.syslog"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereVcenterSyslogDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereVcenterSyslogConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereVcenterSyslogValidation(resourceName, true),
				),
			},
			{
				Config: testAccResourceVSphereVcenterSyslogConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereVcenterSyslogValidation(resourceName, false),
				),
			},
			{
				ResourceName: resourceName,
				Config:       testAccResourceVSphereVcenterSyslogConfig(false),
				ImportState:  true,
			},
		},
	})
}

func testAccVSphereVcenterSyslogDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}

		client := testAccProvider.Meta().(*Client).restClient
		lm := logging.NewManager(client)
		logs, err := lm.Forwarding(context.Background())
		if err != nil {
			return err
		}

		if len(logs) != 0 {
			return fmt.Errorf("there should not be any log configuations set")
		}

		return nil
	}
}

func testAccVSphereVcenterSyslogValidation(resourceName string, isCreate bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("%s key not found on the server", resourceName)
		}

		client := testAccProvider.Meta().(*Client).restClient
		lm := logging.NewManager(client)
		logs, err := lm.Forwarding(context.Background())
		if err != nil {
			return err
		}

		if isCreate {
			if len(logs) != 1 {
				return fmt.Errorf("should have 1 log configuration; got %d", len(logs))
			}
		} else {
			if len(logs) != 2 {
				return fmt.Errorf("should have 2 log configuration; got %d", len(logs))
			}
		}

		return nil
	}
}

func testAccResourceVSphereVcenterSyslogConfig(isCreate bool) string {
	if isCreate {
		return `
		resource "vsphere_vcenter_syslog" "syslog" {
			log_server {
				hostname = "host.example.com"
				port = 514
				protocol = "UDP"
			}
		}
		`
	}

	return `
	resource "vsphere_vcenter_syslog" "syslog" {
		log_server {
			hostname = "host.example.com"
			port = 514
			protocol = "UDP"
		}
		log_server {
			hostname = "host2.example.com"
			port = 514
			protocol = "UDP"
		}
	}
	`
}
