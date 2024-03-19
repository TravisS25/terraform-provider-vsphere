---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_host_config_snmp"
sidebar_current: "docs-vsphere-resource-host-config-snmp"
description: |-
  Sets snmp configuration for esxi host
---

# vsphere_host_config_snmp

`vsphere_host_config_snmp` Sets snmp configuration for esxi host

## Example Usages

**Basic example:**

```hcl
resource "vsphere_host_config_snmp" "host" {
  host_system_id = "host-01"
  user     = "root"
  password = var.vsphere_password
  hostname = "nor1devhvmw98.dev.encore.internal"
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

**Using hostname:**

```hcl
resource "vsphere_host_config_snmp" "host" {
  hostname = "host.example.com"
  user     = "root"
  password = var.vsphere_password
  hostname = "nor1devhvmw98.dev.encore.internal"
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

* `host_system_id` - (Required/Optional) The id of the host we want to gather snmp info
* `hostname` - (Required/Optional) The hostname of the host we want to gather snmp info
* `user` - (Required) The user of esxi host to login as through ssh
* `password` - (Optional) The user of esxi host to login as through ssh
* `known_hosts_path` - (Optional) File path to 'known_hosts' file that must contain the hostname of esxi host.  This is used to verify a host against their current public ssh key.  Must be full path
* `ssh_port` - (Optional) The port of esxi host to connect to through ssh
* `ssh_timeout` - (Optional) Number in seconds it should take to establish connection before timing out
* `engine_id` - (Required) A unique identifier used for SNMP communication within vmware environments.  We can think of this as like a mac address for snmp that we can set.  Must be at least 10 to 32 hexadecimal characters
* `authentication_protocol` - (Optional) Protocol used ensure the identity of users of SNMP v3
* `privacy_protocol` - (Optional) Protocol used to allow encryption of SNMP v3 messages
* `log_level` - (Optional) Log level the host snmp agent will output.  Options are:
  * `info`
  * `warning`
  * `debug`
  * `error`
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

* `id` - Same as `host_system_id` or `hostname`

## Importing

Importing the current snmp configuration for host can be done via `host_system_id` or `hostname`.  Using `host_system_id`:

```
terraform import vsphere_host_config_snmp.host host-01
```

The above would import snmp configuration for host with id `host-01`

Using hostname:

```
terraform import vsphere_host_config_snmp.host host.example.com
```

The above would import snmp configuration for host with hostname `host.example.com`

~> **NOTE:** Must set `TF_VAR_vsphere_esxi_ssh_user` and `TF_VAR_vsphere_esxi_ssh_password` env variables to import. Optionally can set `TF_VAR_vsphere_esxi_ssh_port`, `TF_VAR_vsphere_esxi_ssh_timeout` and `TF_VAR_vsphere_ssh_known_hosts_path`