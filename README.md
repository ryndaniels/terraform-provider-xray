# Terraform Provider Xray

This provider provides (limited) support for JFrog Xray. It is modeled after the [Atlassian Artifactory Provider](https://github.com/atlassian/terraform-provider-artifactory).

## Build the Provider
If you're building the provider, follow the instructions to [install it as a plugin](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin).
After placing it into your plugins directory,  run `terraform init` to initialize it.

Requirements:
- [Terraform](https://www.terraform.io/downloads.html) 0.11
- [Go](https://golang.org/doc/install) 1.13+ (to build the provider plugin)

Clone repository to: `$GOPATH/src/github.com/ryndaniels/terraform-provider-xray

Enter the provider directory and build the provider

```sh
cd $GOPATH/src/github.com/ryndaniels/terraform-provider-xray
go build
```

To install the provider
```sh
cd $GOPATH/src/github.com/ryndaniels/terraform-provider-xray
go install
```

## Contributors
This is a best effort provider at the moment. Pull requests, issues and comments are welcomed.