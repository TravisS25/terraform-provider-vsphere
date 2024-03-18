---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_vsphere_snmp"
sidebar_current: "docs-vsphere-resource-vsphere-snmp"
description: |-
  Sets snmp configuration for vcenter host
---

# vsphere_vcenter_snmp

`vsphere_vcenter_snmp` Sets snmp configuration for vcenter host

## Example Usages

**Basic example:**

```hcl
resource "vsphere_vcenter_snmp" "host" {
  user     = "root"
  password = var.vsphere_password
  read_only_communities   = ["public"]
  engine_id               = "80001ADC0510151278081707752953"
  authentication_protocol = "SHA1"
  privacy_protocol        = "AES128"
  log_level               = "debug"
  remote_user {
    name                    = "terraform_user_new"
    authentication_password = "password"
    privacy_secret          = "5jx3CeCm3H5D$Bzu"
  }
  trap_target {
    hostname  = "target.example.com"
    port      = 161
    community = "public"
  }
}
```

## Argument Reference

The following arguments are supported:

* `user` - (Required) The user of vcenter host to login as through ssh
* `password` - (Optional) The password for user
* `known_hosts_path` - (Optional) File path to 'known_hosts' file that must contain the hostname of vcenter host.  This is used to verify a host against their current public ssh key.  Must be full path
* `ssh_port` - (Optional) The port of vcenter host to connect to through ssh
* `ssh_timeout` - (Optional) Number in seconds it should take to establish connection before timing out
* `engine_id` - (Required) Sets SNMPv3 engine id
* `authentication_protocol` - (Optional) Protocol used ensure the identity of users of SNMP v3
* `privacy_protocol` - (Optional) Protocol used to allow encryption of SNMP v3 messages
* `log_level` - (Optional) Log level the host snmp agent will output
* `remote_user` (Optional):
    * `name` - (Required) Name of user
    * `authentication_password` - (Optional) Password of remote user
    * `privacy_secret` - (Optional) Secret to use for encryption of messages
* `snmp_port` - (Optional) Port for the agent listen on
* `read_only_communities` - (Optional) Communities that are read only.  Only valid for version 1 and 2
* `trap_target` (Optional):
    * `hostname` - Hostname of receiver for notifications from host
    * `port` - Port of receiver for notifications from host
    * `community` - Community of receiver for notifications from host

~> **NOTE:** Must choose either `host_system_id` or `hostname` but not both

## Attribute Reference

* `id` - Always returns as `tf-vcenter-snmp`

## Importing

Importing the current snmp configuration for vcenter host can be done by simply using string `tf-vcenter-snmp`

```
terraform import vsphere_vcenter_snmp.vcenter tf-vcenter-snmp
```

The above would import snmp configuration for vcenter host

~> **NOTE:** Must set `TF_VAR_vsphere_vcenter_ssh_user` and `TF_VAR_vsphere_vcenter_ssh_password` env variables to import.  Optionally can set `TF_VAR_vsphere_center_ssh_port`, `TF_VAR_vsphere_vcenter_ssh_timeout` and `TF_VAR_vsphere_ssh_known_hosts_path`