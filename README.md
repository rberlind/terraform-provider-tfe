<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

# Terraform Cloud/Enterprise Provider

The official Terraform provider for [Terraform Cloud/Enterprise](https://www.hashicorp.com/products/terraform).

As Terraform Enterprise is a self-hosted distribution of Terraform Cloud, this
provider supports both Cloud and Enterprise use cases. In all/most
documentation, the platform will always be stated as 'Terraform Enterprise' -
but a feature will be explicitly noted as only supported in one or the other, if
applicable (rare).

Note this provider is in beta and is subject to change (though it is generally
quite stable). We will indicate any breaking changes by releasing new versions.
Until the release of v1.0, any minor version changes will indicate possible
breaking changes. Patch version changes will be used for both bugfixes and
non-breaking changes.

- **Documentation:** https://registry.terraform.io/providers/hashicorp/tfe/latest/docs
- Website: https://registry.terraform.io/providers/hashicorp/tfe / https://www.terraform.io
- Discuss forum: https://discuss.hashicorp.com/c/terraform-providers


## Installation

Declare the provider in your configuration and `terraform init` will automatically fetch and install the provider for you from the [Terraform Registry](https://registry.terraform.io/) (Terraform version 0.12.0+):

```
terraform {
  required_providers {
    tfe = "~> 0.22.0"
  }
}
```

For production use, you should constrain the acceptable provider versions via configuration,
to ensure that new versions with breaking changes will not be automatically installed by
`terraform init` in future. As this provider is still at version zero, you should constrain
the acceptable provider versions on the minor version.

If you are using Terraform CLI version 0.11.x, you can constrain this provider to 0.15.x versions
by adding the version constraint to the `tfe` provider block.

```
provider "tfe" {
  version = "~> 0.15.0"
  ...
}
```

For more information on provider installation and constraining provider versions, see the [Provider Requirements documentation](https://www.terraform.io/docs/configuration/provider-requirements.html).

### Manually building the provider

You might prefer to manually build the provider yourself - perhaps access to the Terraform Registry or the official
release binaries on [releases.hashicorp.com](https://releases.hashicorp.com/terraform-provider-tfe/) are not available
in your operating environment, or you're looking to contribute to the provider and are testing out a custom build.

#### Requirements
-	[Terraform](https://www.terraform.io/downloads.html) >= 0.11.x
-	[Go](https://golang.org/doc/install) >= 1.15

Clone the repository, enter the directory, and build the provider:

```sh
$ git clone git@github.com:hashicorp/terraform-provider-tfe
$ cd terraform-provider-tfe
$ make build
```

This will build the provider and put the binary in the `$GOPATH/bin` directory. To use the compiled binary, you have several different options (this list is not exhaustive!):
* You can copy the provider binary to your `~/.terraform.d/plugins` directory by running the following command:
   ```sh
   $ mv $GOPATH/bin/terraform-provider-tfe ~/.terraform.d/plugins
   ```
* You can create your test Terraform configurations in the same directory as your provider binary or you can copy the provider binary into the same directory as your test configurations.
* You can copy the provider binary into the same locations as your `terraform` binary.

To learn more about using a local build of a provider, you can look at the [documentation on writing custom providers](https://www.terraform.io/docs/extend/writing-custom-providers.html#invoking-the-provider) and the [documentation on how Terraform plugin discovery works](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery)


## Usage

[Create a user or team API token in Terraform Cloud/Enterprise](https://www.terraform.io/docs/cloud/users-teams-organizations/api-tokens.html), and use the token in the provider configuration block:

```hcl
provider "tfe" {
  hostname = "${var.hostname}" # Optional, for use with Terraform Enterprise. Defaults to app.terraform.io.
  token    = "${var.token}"
}

# Create an organization
resource "tfe_organization" "org" {
  # ...
}
```

There are several other ways to configure the authentication token, depending on
your use case. For other methods, see the [Authentication documentation](https://registry.terraform.io/providers/hashicorp/tfe/latest/docs#authentication)

For more information on configuring providers in general, see the [Provider Configuration documentation](https://www.terraform.io/docs/configuration/providers.html).

## Contributing

Thanks for your interest in contributing; we appreciate your help! If you're unsure or afraid of anything, you can
submit a work in progress (WIP) pull request, or file an issue with the parts you know. We'll do our best to guide you
in the right direction, and let you know if there are guidelines we will need to follow. We want people to be able to
participate without fear of doing the wrong thing.

👉 _See [Manually building the provider](#manually-building-the-provider) above._

### Referencing a local version of `go-tfe`

You may want to create configs or run tests against a local version of `go-tfe`. Add the following line to `go.mod` above the require statement, using your local path to `go-tfe`:

```
replace github.com/hashicorp/go-tfe => /path-to-local-repo/go-tfe
```

### Running the tests

See [TESTS.md](https://github.com/hashicorp/terraform-provider-tfe/tree/master/TESTS.md).

## Updating the Changelog

Only update the `Unreleased` section. Make sure you change the unreleased tag to an appropriate version, using [Semantic Versioning](https://semver.org/) as a guideline.

Please use the template below when updating the changelog:
```
<change category>:
* **New Resource:** `name_of_new_resource` ([#123](link-to-PR))
* r/tfe_resource: description of change or bug fix ([#124](link-to-PR))
```

### Change categories

- BREAKING CHANGES: Use this for any changes that aren't backwards compatible. Include details on how to handle these changes.
- FEATURES: Use this for any large new features added.
- ENHANCEMENTS: Use this for smaller new features added.
- BUG FIXES: Use this for any bugs that were fixed.
- NOTES: Use this section if you need to include any additional notes on things like upgrading, upcoming deprecations, or any other information you might want to highlight.
