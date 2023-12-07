---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_host_config_dns"
sidebar_current: "docs-vsphere-resource-host-config-dns"
description: |-
  Allows user to manage the DNS settings on ESXi hosts.
---

# vsphere_host_config_dns

Allows user to manage the DNS settings on ESXi hosts. This includes:
* Hostname of the system
* Domain name of the system (to complete the fqdn of the system using `Hostname`+`Domain name`
* DNS servers used for name resolution
* Search domains

## Example Usages

**Basic Configuration:**

```hcl
resource "vsphere_host_config_dns" "dns" {
  host_system_id = "host-99"
  hostname = "my-vmware-host01"
  dns_servers = ["192.168.1.10", "192.168.1.11"]
  domain_name = "my.domain.com"
  search_domains = ["my.domain.com"]
}
```

**Example of managing multiple hosts**

~> **NOTE:** This example assumes you have a `vsphere_host` resource defined above it named `host` which actually adds the hosts into vCenter. It also assumes that you have defined hosts by FQDN and not by IP address. We will use data from this `vsphere_host` resource to make writing the `vsphere_host_config_dns` resource easier to set the DNS settings across all of the ESXi hosts.

```hcl
resource "vsphere_host_config_dns" "dns" {
  for_each       = vsphere_host.host
  host_system_id = vsphere_host.host[each.key].id
  # This selects just the hostname from the FQDN of the host
  hostname = split(".", vsphere_host.host[each.key].hostname)[0]
  dns_servers = ["192.168.1.10", "192.168.1.11"]
  domain_name = "my.domain.com"
  search_domains = ["my.domain.com"]
}
```
## Argument Reference

The following arguments are supported:
* `host_system_id` - (Required) ESXi host ID of the host you want to configure DNS on.
* `hostname` - (Required) Hostname of the system. NOT a fqdn.
* `domain_name` - (Required) Domain name of the system (to complete the fqdn of the system using `hostname`+`domain_name`.
* `dns_servers` - (Required) The DNS servers used for name resolution.
* `search_domains` - (Required) Search domains used for hostname resolution.


## Importing

Existing DNS configurations can be imported into terraform but is not required. There is no danger in simply defining DNS configurations in terraform and applying them, even if the defined configuration in TF matches the current configuration on the host. The import simply allows you to see no diff when running a `terraform plan` or changes occuring via  `terraform apply`.

The first step of the import is to define the resource in your TF file so you can reference the name you gave the resource in the .TF file in the `terraform import` command

~> **NOTE:** If you are using `for_each` on the `vsphere_host_config_dns` resource you need to import using the "for_each" import example below. The basic import command below will import for only a single host.

### Basic import example

This will import DNS for a single host and no more hosts can use this same resource name.

```
# terraform import vsphere_host_config_dns.name-of-resource-in-tf-file host-ID
terraform import vsphere_host_config_dns.dns host-99
```

### for_each import example

This is used when you use `for_each` on the `vsphere_host_config_dns` resource and will allow you to import multiple hosts into the same `vsphere_host_config_dns` resource name. The example below shows importing two hosts into the TF state. `host01.my.domain.com` has a `host_system_id` of `host-99` and `host02.my.domain.com` has a `host_system_id` of `host-98`

```
# terraform import vsphere_host_config_dns.name-of-resource-in-tf-file['host.fqdn.tld'] host-ID
terraform import 'vsphere_host_config_dns.dns["host01.my.domain.com"]' host-99
terraform import 'vsphere_host_config_dns.dns["host02.my.domain.com"]' host-98
```
