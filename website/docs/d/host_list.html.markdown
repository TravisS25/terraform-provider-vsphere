---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_host_list"
sidebar_current: "docs-vsphere-data-source-host-list"
description: |-
  A data source that can be used to return all esxi hosts connected to datacenter
---

# vsphere_host_list

The `vsphere_host_list` data source can be used to return all esxi hosts connected to datacenter

## Example Usage

```hcl
data "vsphere_host_list" "current" {
  datacenter_id = "datacenter-01"
}
```

## Argument Reference

The following arguments are supported:

* `datacenter_id` - (Required) The id of the datacenter to get esxi hosts info from.


## Attribute Reference

* `id` - Same as `datacenter_id`
* `hosts` - List of all esxi hosts connected to datacenter
    * `host_system_id` - The id of current host
    * `hostname` - The hostname of current host