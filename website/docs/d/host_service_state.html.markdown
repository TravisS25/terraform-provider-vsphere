---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_host_service_state"
sidebar_current: "docs-vsphere-data-source-host-service-state"
description: |-
  A data source that can be used to return an esxi hosts services and their states
---

# vsphere_host_service_state

The `vsphere_host_service_state` data source can be used to gather all the services for a given host
and find out the state and/or policy of each service

~> **NOTE:** This data source will get ALL of the services for given host, whether the service is running or not

## Example Usage

```hcl
data "vsphere_host_service_state" "host" {
  host_system_id = "host-01"
}
```

## Argument Reference

The following arguments are supported:

* `host_system_id` - (Required) The id of the host we want to gather service info from.


## Attribute Reference

* `id` - Same as `host_system_id`
* `host_system_id` - The id of the host we want to gather service info
* `service` - List of all of the host services from given host
    * `key`     - The key of current service
    * `running` - Boolean that indicates whether the current service is running
    * `policy`  - Policy of current service