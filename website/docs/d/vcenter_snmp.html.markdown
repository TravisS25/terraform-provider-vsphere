---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_vcenter_snmp"
sidebar_current: "docs-vsphere-data-source-vcenter-snmp"
description: |-
  A data source used to return the current snmp configuration for vcenter
---

# vsphere_vcenter_snmp

`vsphere_vcenter_snmp` is data source is used to return the current snmp configuration from vcenter host

## Example Usage

```hcl
data "vsphere_host_config_snmp" "host" {
  user     = "root"
  password = var.host_password
  known_hosts_path = /path/to/known_hosts/file
}
```

## Argument Reference

The following arguments are supported:

* `user` - (Required) The user of vcenter host to login as through ssh
* `password` - (Optional) The password of user
* `ssh_port` - (Optional) The port of vcenter host to connect to through ssh
* `ssh_timeout` - (Optional) Number in seconds it should take to establish connection before timing out
* `known_hosts_path` - (Optional) File path to 'known_hosts' file that must contain the hostname of vcenter host.  This is used to verify a host against their current public ssh key.  Must be full path

## Attribute Reference

* `id` - Same as `host_system_id` or `hostname`
* `user` - The user of vcenter host to login as through ssh
* `password` - The password of user
* `known_hosts_path` - File path to 'known_hosts' file that must contain the hostname of vcenter host.  This is used to verify a host against their current public ssh key.  Must be full path
* `ssh_port` - The port of vcenter host to connect to through ssh
* `ssh_timeout` - Number in seconds it should take to establish connection before timing out
* `engine_id` - SNMPv3 engine id / "mac address" of device
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