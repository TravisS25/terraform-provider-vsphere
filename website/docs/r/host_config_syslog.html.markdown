---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_host_config_syslog"
sidebar_current: "docs-vsphere-resource-host-config-syslog"
description: |-
  Allows user to update syslog settings for given esxi host
---

# vsphere_host_config_syslog

Allows user to update syslog settings for given esxi host

## Example Usages

**Basic Configuration:**

```hcl
resource "vsphere_host_config_syslog" "host" {
  host_system_id = "host-01"
  log_host = "udp://host.example.com:514"
  log_level = "debug"
}
```

## Argument Reference

The following arguments are supported:

* `host_system_id` - (Required) ID of esxi host
* `log_host` - (Optional) Sets the remote host the logs will be forwarded to
* `log_level` - (Optional) Sets the log level the esxi host will output.  Options:
    * `info`
    * `debug`
    * `warning`
    * `error`

## Attribute Reference

* `id` - Same as `host_system_id`

## Importing

Existing syslog settings can be imported from host into this resource by supplying
the host's ID.  An example is below:

```
terraform import vsphere_host_config_syslog.host host-01
```

The above would import the syslog settings for host with ID `host-01`.

## Note when deleting syslog settings

When deleting `vsphere_host_config_syslog` resource, all attributes will simply be set to sane defaults.

`log_host` will be set to empty string / null

`log_level` will be set to `info` (the default)
