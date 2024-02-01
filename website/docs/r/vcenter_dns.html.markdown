---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_vcenter_dns"
sidebar_current: "docs-vsphere-resource-vcenter-dns"
description: |-
  Updates vcenter dns servers
---

# vsphere_vcenter_dns

`vsphere_vcenter_dns` Updates vcenter dns servers

## Example Usages

**Basic example:**

```hcl
resource "vsphere_vcenter_dns" "dns" {
  servers = ["172.16.1.1"]
}
```

## Argument Reference

The following arguments are supported:

* `servers` - (Required) DNS servers to set for vcenter

## Importing

Existing vcenter dns servers can be imported via `tf-vcenter-dns`.  An example is below:

```
terraform import vsphere_vcenter_dns.dns tf-vcenter-dns
```

The above would import vcenter dns servers to `vsphere_vcenter_dns.dns`
