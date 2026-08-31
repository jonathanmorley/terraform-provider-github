---
page_title: "github_repository_immutable_releases (Resource) - GitHub"
description: |-
  Manages immutable releases for a GitHub repository
---

# github_repository_immutable_releases (Resource)

This resource allows you to enforce immutable releases for a GitHub repository. Once enabled, the content of existing and future releases cannot be modified or deleted. See the [documentation](https://docs.github.com/en/code-security/how-tos/secure-your-supply-chain/establish-provenance-and-integrity/prevent-release-changes) for details of usage and how this will impact your repository.

## Example Usage

```terraform
resource "github_repository" "example" {
  name        = "my-repo"
  description = "GitHub repo managed by Terraform"
  visibility  = "private"
}

resource "github_repository_immutable_releases" "example" {
  repository = github_repository.example.name
  enabled    = true
}
```

## Argument Reference

The following arguments are supported:

- `repository` - (Required) The name of the repository to configure immutable releases for.

- `enabled` - (Optional) Whether immutable releases are enabled for the repository. Defaults to `true`.

## Attribute Reference

In addition to the above arguments, the following attributes are exported:

- `repository_id` - The ID of the repository.

## Notes

When the underlying repository no longer exists or is archived, this resource
is removed from Terraform state on the next `plan`/`apply` instead of
returning an error.

## Import

Repository immutable releases can be imported using the `repository_name`:

```sh
terraform import github_repository_immutable_releases.example my-repo
```
