---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_vcenter_syslog"
sidebar_current: "docs-vsphere-data-source-vcenter-syslog"
description: |-
  Gathers vcenter syslog configurations
---

# vsphere_vcenter_syslog

`vsphere_vcenter_syslog` Gathers vcenter syslog configurations

## Example Usages

**Basic example:**

```hcl
data "vsphere_vcenter_syslog" "syslog" {}
```

## Attribute Reference

* `log_server` - Configuration for log servers
    * `hostname` - Hostname of log server
    * `protcol` - Protocol of log server.  Values:
        * `TLS`
        * `TCP`
        * `RELP`
        * `UDP`
    * `port` - Port of log server
