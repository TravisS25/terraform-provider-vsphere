---
subcategory: "Storage"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_iscsi_target"
sidebar_current: "docs-vsphere-resource-storage-iscsi-target"
description: |-
  Adds static and send targets for given adapter on given host
---

# vsphere_iscsi_target

`vsphere_iscsi_target` adds static and send targets for given adapter on given host

## Example Usages

**Basic example:**

```hcl
resource "vsphere_iscsi_target" "host" {
  host_system_id = "host-1"
  adapter_id = "vmhba65"
  static_target{
    ip = "172.16.0.1"
    port = 3260
    name = "iqn.test_name_1"
  }
  static_target{
    ip = "172.16.0.2"
    port = 3260
    name = "iqn.test_name_2"
  }
  send_target{
    ip = "172.17.0.1"
    port = 3260
  }
  send_target{
    ip = "172.17.0.2"
    port = 3260
  }
}
```

**Using hostname:**

```hcl
resource "vsphere_iscsi_target" "host" {
  hostname = "host.example.com"
  adapter_id = "vmhba65"
  static_target{
    ip = "172.16.0.1"
    port = 3260
    name = "iqn.test_name_1"
  }
  static_target{
    ip = "172.16.0.2"
    port = 3260
    name = "iqn.test_name_2"
  }
  send_target{
    ip = "172.17.0.1"
    port = 3260
  }
  send_target{
    ip = "172.17.0.2"
    port = 3260
  }
}
```

## Argument Reference

The following arguments are supported:

* `host_system_id` - (Required/Optional) The host id to add static and send targets
* `hostname` - (Required/Optional) The hostname to add static and send targets
* `adapter_id` - (Required) The adapter to attach static or send targets
* `static_target` - (Required/Optional) The set of resource static targets for given host and adapter id
  * `ip` - The ip to set for static target
  * `port` - (Default: 3260) The port to set static target
  * `name` - The iqn name to set static target
* `send_target` - (Required/Optional) The set of resource send targets for given host and adapter id
  * `ip` - The ip to set for send target
  * `port` - The port to set for send target

~> **NOTE:** At least one `static_target` or `send_target` must be set

~> **NOTE:** Must choose either `host_system_id` or `hostname` but not both

## Attribute Reference

* `id` - Represents the host and adapter id of targets in the form of: `<host_system_id | hostname>:<adapter_id>`

## Importing

Existing iscsi adapter targets can be imported via `<host_system_id | hostname>:<adapter_id>`.  An example is below:

```
terraform import vsphere_iscsi_target.host host-1:vmhba65
```

The above would import iscsi targets for host with id `host-1` on adapter `vmhba65`

Using hostname:

```
terraform import vsphere_iscsi_target.host host.example.com:vmhba65
```

The above would import iscsi targets for host with hostname `host.example.com` on adapter `vmhba65`