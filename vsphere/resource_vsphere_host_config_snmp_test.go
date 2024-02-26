// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
	esxissh "github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/ssh"
	"github.com/hashicorp/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
	"golang.org/x/crypto/ssh"
)

func TestAccResourceVSphereHostConfigSNMP_basic(t *testing.T) {
	testAccCheckEnvVariablesF(
		t,
		[]string{
			"TF_VAR_VSPHERE_DATACENTER",
			"TF_VAR_VSPHERE_CLUSTER",
			"TF_VAR_VSPHERE_ESXI1",
			"TF_VAR_VSPHERE_ESXI_SSH_USER",
			"TF_VAR_VSPHERE_ESXI_SSH_PASSWORD",
			"TF_VAR_VSPHERE_SSH_KNOWN_HOSTS_PATH",
		},
	)

	resourceName := "vsphere_host_config_snmp.h1"
	community := "public"
	newCommunity := "new_public"

	_, err := os.OpenFile(
		os.Getenv("TF_VAR_VSPHERE_SSH_KNOWN_HOSTS_PATH"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		os.ModePerm,
	)
	if err != nil {
		t.Fatalf("unable to create file: %s", err)
	}

	if _, err = esxissh.GetKnownHostsOutput(
		os.Getenv("TF_VAR_VSPHERE_SSH_KNOWN_HOSTS_PATH"),
		os.Getenv("TF_VAR_VSPHERE_ESXI1"),
	); err != nil && err == esxissh.ErrHostNotFound {
		runKeyScanCommand(t, os.Getenv("TF_VAR_VSPHERE_ESXI1"))
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostConfigSNMPDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostConfigSNMPConfig(community, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostConfigSNMPValidation(resourceName, community),
				),
			},
			{
				Config: testAccResourceVSphereHostConfigSNMPConfig(newCommunity, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostConfigSNMPValidation(resourceName, newCommunity),
				),
			},
			{
				ResourceName: resourceName,
				Config:       testAccResourceVSphereHostConfigSNMPConfig(resourceName, false),
				ImportState:  true,
			},
		},
	})
}

func TestAccResourceVSphereHostConfigSNMP_hostname(t *testing.T) {
	testAccCheckEnvVariablesF(
		t,
		[]string{
			"TF_VAR_VSPHERE_DATACENTER",
			"TF_VAR_VSPHERE_CLUSTER",
			"TF_VAR_VSPHERE_ESXI1",
			"TF_VAR_VSPHERE_ESXI_SSH_USER",
			"TF_VAR_VSPHERE_ESXI_SSH_PASSWORD",
			"TF_VAR_VSPHERE_SSH_KNOWN_HOSTS_PATH",
		},
	)

	resourceName := "vsphere_host_config_snmp.h1"
	community := "public"
	newCommunity := "new_public"

	_, err := os.OpenFile(
		os.Getenv("TF_VAR_VSPHERE_SSH_KNOWN_HOSTS_PATH"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		os.ModePerm,
	)
	if err != nil {
		t.Fatalf("unable to create file: %s", err)
	}

	if _, err = esxissh.GetKnownHostsOutput(
		os.Getenv("TF_VAR_VSPHERE_SSH_KNOWN_HOSTS_PATH"),
		os.Getenv("TF_VAR_VSPHERE_ESXI1"),
	); err != nil && err == esxissh.ErrHostNotFound {
		runKeyScanCommand(t, os.Getenv("TF_VAR_VSPHERE_ESXI1"))
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostConfigSNMPConfig(community, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostConfigSNMPValidation(resourceName, community),
				),
			},
			{
				Config: testAccResourceVSphereHostConfigSNMPConfig(newCommunity, false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostConfigSNMPValidation(resourceName, newCommunity),
				),
			},
			{
				ResourceName: resourceName,
				Config:       testAccResourceVSphereHostConfigSNMPConfig(resourceName, false),
				ImportState:  true,
			},
		},
	})
}

func testAccResourceVSphereHostConfigSNMPDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}

		outBuf, err := getTestCommandOutput(rs.Primary.ID)
		if err != nil {
			return err
		}

		for {
			line, err := outBuf.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					return fmt.Errorf("error reading configuration output: %s", err)
				}

				break
			}

			lineArr := strings.Split(line, ":")
			key := strings.TrimSpace(strings.ToLower(lineArr[0]))
			value := strings.TrimSpace(lineArr[1])

			switch key {
			case "authentication":
				if value != "none" {
					return fmt.Errorf("authentication_protocol should be 'none', got '%s'", value)
				}
			case "communities":
				if value != "" {
					return fmt.Errorf("communities should be empty, got '%s'", value)
				}
			case "loglevel":
				if value != "warning" {
					return fmt.Errorf("log_level should be 'warning', got '%s'", value)
				}
			case "port":
				if value != "161" {
					return fmt.Errorf("snmp_port should be '161', got '%s'", value)
				}
			case "privacy":
				if value != "none" {
					return fmt.Errorf("privacy_protocol should be 'none', got '%s'", value)
				}
			case "targets":
				if value != "" {
					return fmt.Errorf("trap_target should be empty, got '%s'", value)
				}
			case "remoteusers":
				if value != "" {
					return fmt.Errorf("remote_user should be empty, got '%s'", value)
				}
			}
		}

		return nil
	}
}

func testAccVSphereHostConfigSNMPValidation(resourceName, community string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("%s key not found on the server", resourceName)
		}

		outBuf, err := getTestCommandOutput(rs.Primary.ID)
		if err != nil {
			return err
		}

		for {
			line, err := outBuf.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					return fmt.Errorf("error reading configuration output: %s", err)
				}

				break
			}

			lineArr := strings.Split(line, ":")
			key := strings.TrimSpace(strings.ToLower(lineArr[0]))
			value := strings.TrimSpace(lineArr[1])

			switch key {
			case "communities":
				if value != community {
					return fmt.Errorf("communities should be '%s', got '%s'", community, value)
				}
			}
		}

		return nil
	}
}

func testAccResourceVSphereHostConfigSNMPConfig(community string, useHostname bool) string {
	resourceStr :=
		`
	%s

	resource "vsphere_host_config_snmp" "h1" {
		%s
		user = "%s"
		password = "%s"
		known_hosts_path = "%s"
		read_only_communities = ["%s"]
		engine_id = "80001ADC0517464555781707920697"
		authentication_protocol = "SHA1"
		privacy_protocol = "AES128"
		remote_user {
			name = "user"
			authentication_password = "password"
			privacy_secret = "123456789abcdefg"
		}
		trap_target {
			hostname = "example.com"
			port = 161
			community = "public"
		}
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
			os.Getenv("TF_VAR_VSPHERE_ESXI_SSH_USER"),
			os.Getenv("TF_VAR_VSPHERE_ESXI_SSH_PASSWORD"),
			os.Getenv("TF_VAR_VSPHERE_SSH_KNOWN_HOSTS_PATH"),
			community,
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
		os.Getenv("TF_VAR_VSPHERE_ESXI_SSH_USER"),
		os.Getenv("TF_VAR_VSPHERE_ESXI_SSH_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_SSH_KNOWN_HOSTS_PATH"),
		community,
	)
}

func getTestCommandOutput(id string) (*bytes.Buffer, error) {
	client := testAccProvider.Meta().(*Client).vimClient
	host, _, err := hostsystem.CheckIfHostnameOrID(client, id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving host for 'vsphere_host_config_snmp' on delete test: %s", err)
	}

	sshPort := os.Getenv("TF_VAR_VSPHERE_ESXI_SSH_PORT")
	if sshPort == "" {
		sshPort = "22"
	}

	sshTimeout := os.Getenv("TF_VAR_VSPHERE_ESXI_SSH_TIMEOUT")
	if sshTimeout == "" {
		sshTimeout = "8"
	}

	port, err := strconv.Atoi(sshPort)
	if err != nil {
		return nil, fmt.Errorf("'TF_VAR_VSPHERE_ESXI_SSH_PORT' env variable should be integer")
	}

	timeout, err := strconv.Atoi(sshTimeout)
	if err != nil {
		return nil, fmt.Errorf("'TF_VAR_VSPHERE_ESXI_SSH_TIMEOUT' env variable should be integer")
	}

	result, err := esxissh.RunCommand(
		"/bin/esxcli system snmp get",
		host.Name(),
		port,
		esxissh.GetDefaultClientConfig(
			os.Getenv("TF_VAR_VSPHERE_ESXI_SSH_USER"),
			os.Getenv("TF_VAR_VSPHERE_ESXI_SSH_PASSWORD"),
			timeout,
			ssh.InsecureIgnoreHostKey(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing esxcli user command for host '%s': %s", host.Name(), err)
	}

	outBuf := bytes.Buffer{}
	if _, err = outBuf.ReadFrom(result); err != nil {
		return nil, fmt.Errorf("error reading command output: %s", err)
	}

	return &outBuf, nil
}

func runKeyScanCommand(t *testing.T, hostname string) {
	stdOut := &bytes.Buffer{}
	stdErr := &bytes.Buffer{}
	cmd := exec.Command(
		"sh",
		"-c",
		fmt.Sprintf("ssh-keyscan -H %s >> %s", hostname, os.Getenv("TF_VAR_VSPHERE_SSH_KNOWN_HOSTS_PATH")),
	)
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	err := cmd.Run()
	if err != nil {
		if stdErr.String() != "" {
			t.Fatalf("error running 'ssh-keyscan' command: %s", stdErr.String())
		}
		if stdOut.String() == "" {
			t.Fatalf(fmt.Sprintf("given hostname '%s' was not found in given known_hosts file", hostname))
		}
	}
}
