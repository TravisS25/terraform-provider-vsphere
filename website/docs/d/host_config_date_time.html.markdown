---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_host_config_date_time"
sidebar_current: "docs-vsphere-data-source-host-config-date-time"
description: |-
  A data source used to return the current date time configuration
---

# vsphere_host_config_date_time

`vsphere_host_config_date_time` data source is used to return the current date time configuration

## Example Usage

```hcl
data "vsphere_host_config_date_time" "host" {
  host_system_id = "host-01"
}
```

## Argument Reference

The following arguments are supported:

* `host_system_id` - (Required) The id of the host we want to gather date time configuration


## Attribute Reference

* `id` - Same as `host_system_id`
* `host_system_id` - The id of the host we want to gather date time info
* `ntp_servers` - Gathers list of ntp servers set for given host
* `protocol` - Gathers network time configuration for clock
* `events_disabled` - Gathers whether events are disabled
* `fallback_disabled` - Gathers whether fallback to ntp is disabled