---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_vnic_list
sidebar_current: "docs-vsphere-data-source-vnic-list"
description: |-
  Gathers vnic configuration for given host
---

# vsphere_vnic_list

`vsphere_vnic_list` Gathers vnic configuration for given host

## Example Usages

**Basic example:**

```hcl
data "vsphere_vnic_list" "h1" {
    hostname = "host.example.com"
}
```

## Argument Reference

* `host_system_id` - (Required/Optional) The id of the host we want to gather vnic configuration
* `hostname` - (Required/Optional) The hostname of the host we want to gather vnic configuration

~> **NOTE:** Must choose either `host_system_id` or `hostname` but not both

## Attribute Reference

* `vnics` - List of configurations for vnics
    * `device` - Device name of vnic
    * `port` - Port of vnic
    * `spec`:
        * `mac` - Mac of vnic
        * `mtu` - MTU of vnic
        * `ip`:
            * `dhcp` - Determines if vnic gets address based on dhcp
            * `ip_address` - IP of vnic
            * `subnet_mask` - Subnet mask of ip for vnic
