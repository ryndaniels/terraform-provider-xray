---
layout: "xray"
page_title: "Provider: Xray"
sidebar_current: "docs-xray-index"
description: |-
  The Xray provider is used to deploy jfrog xray resources
---

# Xray Provider

The [Xray](https://jfrog.com/xray/) provider is used to interact with the
resources supported by Xray. The provider needs to be configured
with the proper credentials before it can be used.

- Available Resources
    * [Policy](./r/xray_policy.html.markdown)
    * [Watch](./r/xray_watch.html.markdown)

## Example Usage
```hcl
# Configure the Xray provider
provider "xray" {
  url = "${var.xray_url}"
  username = "${var.xray_username}"
  password = "${var.xray_password}"
}

# Create a new policy
resource "xray_policy" "example" {
  name  = "License Policy"
  description = "example of a license policy"
  type = "license"

  rules [
    name = "license rule"
    priority = 1
    criteria [
      allowed_licenses = ["0BSD", "AAL"]
    ]
  ]
}
```

## Authentication
The Xray provider supports multiple means of authentication. The following methods are supported:
    * Basic Auth
    * Access Token

### Basic Auth
Basic auth may be used by adding a `username` and `password` field to the provider block.
Getting this value from the environment is supported with the `XRAY_USERNAME` and `XRAY_PASSWORD` variables. It is not recommended to store
sensitive values such as passwords in plaintext in your Terraform code.

Usage:
```hcl
# Configure the Xray provider
provider "xray" {
  url = "xray.site.com"
  username = "myusername"
  password = "mypassword"
}
```

### Bearer Token
Xray access tokens may be used via the Authorization header by providing the `access_token` field to the provider
block. Getting this value from the environment is supported with the `XRAY_ACCESS_TOKEN` variable. It is not recommended to store
sensitive values such as access tokens in plaintext in your Terraform code.

Usage:
```hcl
# Configure the Xray provider
provider "xray" {
  url = "xray.site.com"
  access_token = "ASDFGHJKL1234567890"
}
```

## Argument Reference

The following arguments are supported:

* `url` - (Required) URL of Xray. This can also be sourced from the `XRAY_URL` environment variable.
* `username` - (Optional) Username for basic auth. Requires `password` to be set. 
    Conflicts with `api_key`, and `access_token`. This can also be sourced from the `XRAY_USERNAME` environment variable.
* `password` - (Optional) Password for basic auth. Requires `username` to be set. 
    Conflicts with `api_key`, and `access_token`. This can also be sourced from the `XRAY_PASSWORD` environment variable.
* `access_token` - (Optional) API key for token auth. Uses `Authorization: Bearer` header. 
    Conflicts with `username` and `password`, and `api_key`. This can also be sourced from the `XRAY_ACCESS_TOKEN` environment variable.
