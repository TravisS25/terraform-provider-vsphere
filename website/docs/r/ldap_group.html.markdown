---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_ldap_group"
sidebar_current: "docs-vsphere-resource-ldap-group"
description: |-
  Allows user to add LDAP groups into vCenter groups.
---

# vsphere_ldap_group

Allows user to add LDAP groups into vCenter groups.

~> **NOTE:** This resource depends on an LDAP identity source being configured before it is ran. Terraform can add these identity sources via the `vsphere_ldap_identity_source` resource.

## Example Usages

**Basic Configuration:**

```hcl
resource "vsphere_ldap_group" "admins" {
  ldap_group = "vmware-admins"
  vsphere_group = "administrators"
  domain_name = "domain.com"
}

```

## Argument Reference

The following arguments are supported:
* `ldap_group` - (Required) Name of the LDAP group
* `vsphere_group` - (Required) Name of the vSphere group the `ldap_group` will be added to
* `domain_name` - (Required) Name LDAP domain which contains the `ldap_group`

## Importing

Existing LDAP groups can be imported into terraform state.

The first step of the import is to define the resource in your TF file so you can reference the name you gave the resource in the .TF file in the `terraform import` command

An example is below:

```
# terraform import vsphere_ldap_group.name-of-resource-in-tf-file vsphere_group:ldap_group
terraform import vsphere_ldap_group.admins administrators:vmware-admins
```

The above would import the currently `ldap_group` 'vmware-admins' which is a member of the `vsphere_group` 'administrators'

