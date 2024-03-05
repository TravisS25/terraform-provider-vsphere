---
subcategory: "Storage"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_iscsi_target"
sidebar_current: "docs-vsphere-resource-storage-iscsi-target"
description: |-
  Provides a data source to return all the targets for given adapter on given host
---

# vsphere_iscsi_iscsi

`vsphere_iscsi_target` data source returns all the targets for given adapter on given host

## Example Usage

```hcl
data "vsphere_iscsi_target" "host" {
  host_system_id = "host-1"
  adapter_id = "vmhba65"
}
```

## Argument Reference

The following arguments are supported:

* `host_system_id` - (Required/Optional) The host id to gather target information
* `hostname` - (Required/Optional) The hostname to gather target information
* `adapter_id` - (Required) The adapter on given host to gather target information

~> **NOTE:** Must choose either `host_system_id` or `hostname` but not both

## Attribute Reference

* `id` - Represents the host and adpater id of targets in the form of: `<host_system_id | hostname>:<adapter_id>`
* `host_system_id` - The host id the iscsi adapter is attached to
* `hostname` - The hostname the iscsi adapter is attached to
* `adapter_id` - The iscsi adapter id the targets will be attached to
* `static_target` - The set of resource static targets for given host and adapter id
  * `ip` - The ip of the static target
  * `port` - The port of the static target
  * `name` - The iqn name of the static target
* `send_target` - The set of resource send targets for given host and adapter id
  * `ip` - The ip of the send target
  * `port` - The port of the send target
