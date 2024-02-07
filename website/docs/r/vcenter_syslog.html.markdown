---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_vcenter_syslog"
sidebar_current: "docs-vsphere-resource-vcenter-syslog"
description: |-
  Updates vcenter syslog configurations
---

# vsphere_vcenter_syslog

`vsphere_vcenter_syslog` Updates vcenter syslog configurations

## Example Usages

**Basic example:**

```hcl
resource "vsphere_vcenter_syslog" "syslog" {
  log_server {
    protocol = "UDP"
    hostname = "host.example.com"
    port = 514
  }
}
```

**Multiple log servers:**

```hcl
resource "vsphere_vcenter_syslog" "syslog" {
  log_server {
    protocol = "UDP"
    hostname = "host.example.com"
    port = 514
  }
  log_server {
    protocol = "UDP"
    hostname = "host2.example.com"
    port = 514
  }
}
```

## Argument Reference

The following arguments are supported:

* `log_server` - (Required) Configuration for forwarding logs
    * `hostname` - (Required) Hostname of server to forward logs to
    * `protcol` - (Required) Protocol to use when sending logs.  Options:
        * `TLS`
        * `TCP`
        * `RELP`
        * `UDP`
    * `port` - (Required) Port of server to forward requests to

~> **NOTE:** Only a total of 3 servers can be set

## Importing

Existing syslog servers can be imported via `tf-vcenter-syslog`.  An example is below:

```
terraform import vsphere_vcenter_syslog.syslog tf-vcenter-syslog
```

The above would import vcenter log servers to `vsphere_vcenter_syslog.syslog`
