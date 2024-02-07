---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_host_config_snmp"
sidebar_current: "docs-vsphere-data-source-host-config-snmp"
description: |-
  A data source used to return the current snmp configuration
---

# vsphere_host_config_snmp

`vsphere_host_config_snmp` data source is used to return the current snmp configuration

## Example Usage

```hcl
data "vsphere_host_config_date_time" "host" {
  user     = "root"
  password = var.host_password
  host_system_id = "host-01"
}
```

**Using hostname**

```hcl
data "vsphere_host_config_date_time" "host" {
  user     = "root"
  password = var.host_password
  hostname = "nor1devhvmw98.dev.encore.internal"
}
```

## Argument Reference

The following arguments are supported:

* `host_system_id` - (Required/Optional) The id of the host we want to gather snmp configuration
* `hostname` - (Required/Optional) The hostname of the host we want to gather snmp configuration
* `user` - (Required) The user of esxi host to login as through ssh
* `password` - (Optional) The user of esxi host to login as through ssh
* `ssh_port` - (Optional) The port of esxi host to connect to through ssh
* `ssh_timeout` - (Optional) Number in seconds it should take to establish connection before timing out

## Attribute Reference

* `id` - Same as `host_system_id` or `hostname`
* `host_system_id` - The id of the host we want to gather snmp info
* `hostname` - The hostname of the host we want to gather snmp info
* `user` - The user of esxi host to login as through ssh
* `password` - The user of esxi host to login as through ssh
* `ssh_port` - The port of esxi host to connect to through ssh
* `ssh_timeout` - Number in seconds it should take to establish connection before timing out
* `engine_id` - SNMPv3 engine id
* `authentication_protocol` - Protocol used ensure the identity of users of SNMP v3
* `privacy_protocol` - Protocol used to allow encryption of SNMP v3 messages
* `log_level` - Log level the host snmp agent will output
* `remote_user`:
    * `name`: Name of user
* `snmp_port` - Port for the agent listen on
* `read_only_communities` - Communities that are read only.  Only valid for version 1 and 2
* `trap_target`:
    * `hostname` - Hostname of receiver for notifications from host
    * `port` - Port of receiver for notifications from host
    * `community` - Community of receiver for notifications from host