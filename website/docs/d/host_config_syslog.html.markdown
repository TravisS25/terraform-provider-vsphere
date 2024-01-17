---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_host_config_syslog"
sidebar_current: "docs-vsphere-data-source-host-config-syslog"
description: |-
  A data source that can be used to return esxi syslog information
---

# vsphere_host_config_syslog

The `vsphere_host_config_syslog` data source can be used to gather syslog information for given esxi host

## Example Usage

```hcl
data "vsphere_host_config_syslog" "host" {
  host_system_id = "host-01"
}
```

## Example Using Hostname

```hcl
data "vsphere_host_config_syslog" "host" {
  hostname = "host.example.com"
}
```

## Argument Reference

The following arguments are supported:

* `host_system_id` - (Required/Optional) The id of the host we want to gather syslog info from.
* `hostname` - (Required/Optional) The hotname of the host we want to gather syslog info from.

~> **Note:** Must either use `host_system_id` or `hostname` but not both

## Attribute Reference

* `id` - Same as `host_system_id` or `hostname`
* `host_system_id` - The id of the host we want to gather syslog info
* `hostname` - The hostname of the host we want to gather syslog info
* `log_host` - Gets the current log host(s) for current esxi host
* `log_level` - Gets the log level that the esxi hosts is currently outputing