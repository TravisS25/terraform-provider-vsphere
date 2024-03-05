---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_vcenter_dns"
sidebar_current: "docs-vsphere-data-source-vcenter-dns"
description: |-
  Gathers vcenter dns configurations
---

# vsphere_vapp_dns

`vsphere_vcenter_dns` Gathers vcenter dns configurations

## Example Usages

**Basic example:**

```hcl
data "vsphere_vcenter_dns" "dns" {}
```

## Attribute Reference

* `servers` - DNS servers from vcenter
