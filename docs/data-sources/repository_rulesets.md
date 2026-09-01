---
page_title: "github_repository_rulesets (Data Source) - GitHub"
description: |-
  Get information about a repository's repository rulesets.
---

# github_repository_rulesets (Data Source)

Use this data source to retrieve a list of a repository's repository rulesets.

## Example Usage

```terraform
data "github_repository_rulesets" "example" {
  repository = "example"
}
```

## Argument Reference

The following arguments are supported:

- `repository` - (Required) The GitHub repository name.

## Attribute Reference

- `rulesets` - Collection of repository rulesets. Each of the results conforms to the following scheme:

  - `id` - The ruleset ID.
  - `name` - The ruleset name.
  - `target` - The target of the ruleset.
  - `source_type` - The type of the source of the ruleset (`Repository` for repository rulesets).
  - `source` - The name of the source (the repository or organization).
  - `enforcement` - The enforcement level of the ruleset (one of `disabled`, `active` or `evaluate`).
  - `node_id` - The Node ID of the ruleset.