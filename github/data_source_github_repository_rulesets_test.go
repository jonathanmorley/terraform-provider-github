package github

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGithubRepositoryRulesetsDataSource(t *testing.T) {
	t.Parallel()

	t.Run("queries repository rulesets without error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("%srepo-rulesets-%s", testResourcePrefix, randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name      = "%[1]s"
				auto_init = true
			}

			resource "github_repository_ruleset" "test" {
				name        = "test-ruleset"
				repository  = github_repository.test.name
				target      = "branch"
				enforcement = "active"

				conditions {
					ref_name {
						include = ["~DEFAULT_BRANCH"]
						exclude = []
					}
				}

				rules {
					required_linear_history = true
				}
			}

			data "github_repository_rulesets" "all" {
				repository = github_repository.test.name

				depends_on = [github_repository_ruleset.test]
			}
		`, repoName)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("data.github_repository_rulesets.all", "rulesets.0.id"),
			resource.TestCheckResourceAttr("data.github_repository_rulesets.all", "rulesets.0.name", "test-ruleset"),
			resource.TestCheckResourceAttr("data.github_repository_rulesets.all", "rulesets.0.target", "branch"),
			resource.TestCheckResourceAttr("data.github_repository_rulesets.all", "rulesets.0.enforcement", "active"),
			resource.TestCheckResourceAttrSet("data.github_repository_rulesets.all", "rulesets.0.node_id"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
			},
		})
	})
}
