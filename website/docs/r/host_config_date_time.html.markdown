---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_host_config_date_time"
sidebar_current: "docs-vsphere-resource-host-config-date-time"
description: |-
  Sets date time configuration for esxi host
---

# vsphere_host_config_date_time

`vsphere_host_config_date_time` Sets date time configuration for esxi host

## Example Usages

**Basic example:**

```hcl
resource "vsphere_host_config_date_time" "host" {
  host_system_id = "host-01"
  ntp_servers = ["0.north-america.pool.ntp.org"]
}
```

**Using hostname:**

```hcl
resource "vsphere_host_config_date_time" "host" {
  hostname = "host.example.com"
  ntp_servers = ["0.north-america.pool.ntp.org"]
}
```

## Argument Reference

The following arguments are supported:

* `host_system_id` - (Required/Optional) The host id to set date time configuration
* `hostname` - (Required/Optional) The hostname to set date time configuration
* `ntp_servers` - (Required/Optional) The ntp server list to use for syncing time via ip/fqdn
* `protocol` - (Optional) Specify which network time configuration to discipline vmkernel clock.  Options are (case sensitive):
  * `ntp`
* `disable_events` - (Optional) Disables detected failures being sent to VCenter if set
* `disable_fallback` - (Optional) Disables falling back to ntp if ptp fails when set

~> **NOTE:** Must choose either `host_system_id` or `hostname` but not both

## Attribute Reference

* `id` - Same as `host_system_id` or `hostname`

## Importing

Importing the current date time configuration for host can be done via `host_system_id` or `hostname`.  Using `host_system_id`:

```
terraform import vsphere_host_config_date_time.host host-01
```

The above would import date time configuration for host with id `host-01`

Using hostname:

```
terraform import vsphere_host_config_date_time.host host.example.com
```

The above would import date time configuration for host with hostname `host.example.com`