---
page_title: "github_code_scanning_default_setup (Resource) - GitHub"
description: |-
  Manages code scanning default setup for a GitHub repository
---

# github_code_scanning_default_setup (Resource)

This resource allows you to manage code scanning default setup for a GitHub repository. See the [documentation](https://docs.github.com/en/code-security/code-scanning/enabling-code-scanning-for-a-repository) for details of usage and how this will impact your repository.

## Example Usage

```terraform
resource "github_repository" "example" {
  name        = "my-repo"
  description = "GitHub repo managed by Terraform"
  visibility  = "private"
}

resource "github_code_scanning_default_setup" "example" {
  repository  = github_repository.example.name
  query_suite = "default"
}
```

## Argument Reference

The following arguments are supported:

- `repository` - (Required) The name of the repository to configure code scanning default setup for.

- `query_suite` - (Optional) The CodeQL query suite to run. Can be one of `default` or `extended`. Defaults to `default`.

- `languages` - (Optional) The set of languages to analyze with code scanning default setup. Can be one of `actions`, `c-cpp`, `csharp`, `go`, `java-kotlin`, `javascript-typescript`, `python`, `ruby` or `swift`. Omit this argument to let GitHub auto-detect the languages of the repository.

## Attribute Reference

In addition to the above arguments, the following attributes are exported:

- `repository_id` - The ID of the repository.

## Notes

Configuring code scanning default setup is asynchronous: GitHub schedules the
configuration change and returns before the analysis has started.

GitHub may report detected languages outside of the list documented above
(for example `kotlin`). Such values are recorded in state without error.

When code scanning default setup is no longer configured on GitHub's side,
or the underlying repository no longer exists or is archived, this resource
is removed from Terraform state on the next `plan`/`apply` instead of
returning an error.

## Import

Code scanning default setup can be imported using the `repository_name`:

```sh
terraform import github_code_scanning_default_setup.example my-repo
```
