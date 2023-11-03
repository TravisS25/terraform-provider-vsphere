---
subcategory: "Storage"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_iscsi_software_adapter"
sidebar_current: "docs-vsphere-resource-storage-iscsi-software-adapter"
description: |-
  Provides a data source to return the iscsi software adapter information for given host
---

# vsphere_iscsi_software_adapter

The `vsphere_iscsi_software_adapter` data source can be used to discover the iscsi software adapter information of a given host

## Example Usage

```hcl
data "vsphere_iscsi_software_adapter" "host" {
  host_system_id = "my_host_id"
}
```

## Argument Reference

The following arguments are supported:

* `host_system_id` - (Required) The host id we want to obtain iscsi software adapter information.

~> **NOTE:** The iscsi software adapter for given host must already be enabled to grab information

## Attribute Reference

* `id` - The same as the `host_system_id` parameter
* `host_system_id` - The host id the iscsi software adapter is attached to
* `iscsi_name` - The iscsi software adapter name from either being user defined or vmware generated
