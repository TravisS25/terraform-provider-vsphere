---
subcategory: "Storage"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_iscsi_software_adapter"
sidebar_current: "docs-vsphere-resource-storage-iscsi-software-adapter"
description: |-
  Enables VMware vSphere iscsi software adapter for a given host
---

# vsphere_iscsi_software_adapter

Enables VMware vSphere iscsi software adapter for a given host

## Example Usages

**Enable iscsi adapter for host:**

```hcl
resource "vsphere_iscsi_software_adapter" "host" {
  host_system_id = "my_host_id"
}
```

**Override iscsi name for iscsi software adapter:**

```hcl
resource "vsphere_iscsi_software_adapter" "host" {
  host_system_id   = "my_host_id"
  iscsi_name = "custom_iscsi_name"
}
```

## Argument Reference

The following arguments are supported:

* `host_system_id` - (Required) The host id to enable iscsi software adapter
* `iscsi_name` - (Optional) The unique iqn name for the iscsi software adapter.  If left blank, vmware will generate the iqn name


## Attribute Reference

* `id` - Represents the host and software adapter id in the form of: `<host_system_id>:<adapter_id>`
* `host_system_id` - The host id the iscsi software adapter is attached to
* `iscsi_name` - The iscsi software adapter name from either being user defined or vmware generated

## Importing

An existing iscsi software adapter can be imported into this resource
via `<host_system_id>:<adapter_id>`.  An example is below:

```
terraform import vsphere_iscsi_software_adapter.host host-1:vmhba65
```

The above would import the iscsi software adapter from host `host-1` and software adapter id of `vmhba65`

~> **NOTE:** The iscsi software adapter for given host must already be enabled for import to work or an error will occur
