---
subcategory: "Host and Cluster Management"
layout: "vsphere"
page_title: "VMware vSphere: vsphere_ldap_identity_source"
sidebar_current: "docs-vsphere-resource-ldap-identity-source"
description: |-
  Allows user to add LDAP identity authentication sources for vCenter logins.
---

# vsphere_ldap_identity_source

Allows user to add LDAP identity authentication sources for vCenter logins.

## Example Usages

**Basic Configuration:**

```hcl
resource "vsphere_ldap_identity_source" "domain" {
  ldap_username = "my_ldap_username@domain.com"
  ldap_password = "my-secure-password"
  domain_name = "domain.com"
  domain_alias = "domain.com"
  server_type   = "ActiveDirectory"
  friendly_name = "domain.com"
  user_base_dn  = "dc=domain,dc=com"
  group_base_dn = "dc=domain,dc=com"
  primary_url   = "ldap://domain-controller01.domain.com"
  failover_url  = "ldap://domain-controller02.domain.com"
}

```

## Argument Reference

The following arguments are supported:


* `ldap_username` - (Required) Username of account used to authenticate with LDAP
* `ldap_password` - (Required) Password of account used to authenticate with LDAP
* `domain_name` - (Required) The name of the LDAP domain
* `domain_alias` - (Required) The alias of the LDAP domain
* `server_type` - The type of LDAP to bind with. Defaults to "ActiveDirectory"
* `friendly_name` - (Required) Friendly name used to identity the authentication source
* `user_base_dn` - (Required) Base distinguished name (dn) to look for LDAP user accounts.
* `group_base_dn` - (Required) Base distinguished name (dn) to look for LDAP user group membership.
* `primary_url` - (Required) The primary URL vCenter will use to reach a domain controller. Can be a load balancer or aimed directly at a AD-DC.
* `failover_url` - (Required) The failover URL vCenter will use to reach a domain controller. Can be a load balancer or aimed directly at a AD-DC. Can be a blank string. Cannot be the same as `primary_url`

## Importing

Existing LDAP identity sources can be imported into terraform state.

The first step of the import is to define the resource in your TF file so you can reference the name you gave the resource in the .TF file in the `terraform import` command

~> **NOTE:** This import will create the resource within state successfully but the next 'terraform apply' WILL note some changes for it, even if there is nothing actually changing. This occurs due to our inability to fetch the currently configured password that LDAP is using and TF will enforce the ones defined in it on the next `terraform apply`

An example is below:

```
# terraform import vsphere_ldap_identity_source.name-of-resource-in-tf-file domain-name-goes-here
terraform import vsphere_ldap_identity_source.domain domain.com
```

The above would import the currently configured identity source for `domain.com`

As previously mentioned, the next `terraform apply` *WILL* have a change for this identity source to enforce the `ldap_password` - you can see this is the only item changing via a `terraform plan`

