---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_host_service_state"
sidebar_current: "docs-vsphere-resource-host-service-state"
description: |-
  Allows user to update state and/or policy of services on a given esxi host
---

# vsphere_host_service_state

Allows user to update state and/or policy of services on a given esxi host

## Example Usages

**Basic Configuration:**

```hcl
resource "vsphere_host_service_state" "host" {
  host_system_id = "host-01"

  service {
    key = "TSM-SSH"
    policy = "on"
  }
}
```

**Create with multiple services:**

```hcl
resource "vsphere_host_service_state" "host" {
  host_system_id = "host-01"

  service {
    key = "TSM-SSH"
    policy = "on"
  }

  service {
    key = "TSM"
    policy = "automatic"
  }
}
```

**Apply to multiple hosts:**

```hcl
locals {
    host_ids = ["host-01", "host-02"]
}

resource "vsphere_host_service_state" "hosts" {
  for_each = local.host_ids
  host_system_id = each.value

  service {
    key = "TSM-SSH"
    policy = "on"
  }

  service {
    key = "TSM"
    policy = "automatic"
  }
}
```

## Argument Reference

The following arguments are supported:

* `host_system_id` - (Required) ID of esxi host
* `service` - (Required) List of host services to enable
    * `key` - (Required) The key to service to enable (case sensitive)
        * `DCUI`           - Direct Console UI
        * `TSM`            - ESXi Shell
        * `TSM-SSH`        - SSH
        * `attestd`        - attestd
        * `dpd`            - dpd
        * `kmxd`           - kmxd
        * `lbtd`           - Load-Based Teaming Daemon
        * `lwsmd`          - Active Directory Service
        * `ntpd`           - NTP Daemon
        * `pcscd`          - PC/SC Smart Card Daemon
        * `ptpd`           - PTP Daemon
        * `sfcbd-watchdog` - CIM Server
        * `slpd`           - slpd
        * `snmpd`          - SNMP Server
        * `vltd`           - vltd
        * `xorg`           - X.Org Server
    * `policy` - (Required) The policy to service to enable (case sensitive)
        * `on`
        * `off`
        * `automatic`

## Attribute Reference

* `id` - The ID of the host that services are updated.

## Importing

Existing services can be imported from host into this resource by supplying
the host's ID.  An example is below:

~> **NOTE:** Only services that are actively running on host will be imported

```
terraform import vsphere_host_service_state.host host-01
```

The above would import the active running services for host with ID `host-01`.

## Note when deleting service/resource

When removing a service from the `vsphere_host_service_state` resource or removing the entire resource itself,
the services are simply turned off
