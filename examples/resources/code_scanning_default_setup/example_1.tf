resource "github_repository" "example" {
  name        = "my-repo"
  description = "GitHub repo managed by Terraform"
  visibility  = "private"
}

resource "github_code_scanning_default_setup" "example" {
  repository  = github_repository.example.name
  query_suite = "default"
}
