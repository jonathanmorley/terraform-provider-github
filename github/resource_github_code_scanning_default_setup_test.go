package github

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccGithubCodeScanningDefaultSetup(t *testing.T) {
	t.Parallel()

	t.Run("creates_code_scanning_default_setup_without_error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandString(5)
		repoName := fmt.Sprintf("%scode-scanning-default-%s", testResourcePrefix, randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name       = "%s"
				visibility = "private"
				auto_init  = true
			}

			resource "github_code_scanning_default_setup" "test" {
				repository  = github_repository.test.name
				query_suite = "default"
			}
		`, repoName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("github_code_scanning_default_setup.test", tfjsonpath.New("query_suite"), knownvalue.StringExact("default")),
						statecheck.CompareValuePairs("github_code_scanning_default_setup.test", tfjsonpath.New("repository"), "github_repository.test", tfjsonpath.New("name"), compare.ValuesSame()),
						statecheck.ExpectKnownValue("github_code_scanning_default_setup.test", tfjsonpath.New("repository_id"), knownvalue.NotNull()),
					},
				},
			},
		})
	})

	t.Run("updates_code_scanning_default_setup_without_error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandString(5)
		repoName := fmt.Sprintf("%scode-scanning-default-%s", testResourcePrefix, randomID)

		config := `
			resource "github_repository" "test" {
				name       = "%s"
				visibility = "private"
				auto_init  = true
			}

			resource "github_code_scanning_default_setup" "test" {
				repository  = github_repository.test.name
				query_suite = "%s"
			}
		`

		compareValuesDiffer := statecheck.CompareValue(compare.ValuesDiffer())

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(config, repoName, "default"),
					ConfigStateChecks: []statecheck.StateCheck{
						compareValuesDiffer.AddStateValue("github_code_scanning_default_setup.test", tfjsonpath.New("query_suite")),
					},
				},
				{
					Config: fmt.Sprintf(config, repoName, "extended"),
					ConfigStateChecks: []statecheck.StateCheck{
						compareValuesDiffer.AddStateValue("github_code_scanning_default_setup.test", tfjsonpath.New("query_suite")),
					},
				},
			},
		})
	})

	t.Run("creates_code_scanning_default_setup_with_languages_without_error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandString(5)
		repoName := fmt.Sprintf("%scode-scanning-default-%s", testResourcePrefix, randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name       = "%s"
				visibility = "private"
				auto_init  = true
			}

			resource "github_code_scanning_default_setup" "test" {
				repository = github_repository.test.name
				languages  = ["javascript-typescript"]
			}
		`, repoName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("github_code_scanning_default_setup.test", tfjsonpath.New("languages"), knownvalue.SetSizeExact(1)),
					},
				},
			},
		})
	})

	t.Run("imports_code_scanning_default_setup_without_error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandString(5)
		repoName := fmt.Sprintf("%scode-scanning-default-%s", testResourcePrefix, randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name       = "%s"
				visibility = "private"
				auto_init  = true
			}

			resource "github_code_scanning_default_setup" "test" {
				repository  = github_repository.test.name
				query_suite = "default"
			}
		`, repoName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("github_code_scanning_default_setup.test", tfjsonpath.New("repository_id"), knownvalue.NotNull()),
					},
				},
				{
					ResourceName:            "github_code_scanning_default_setup.test",
					ImportState:             true,
					ImportStateId:           repoName,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"languages", "query_suite"},
				},
			},
		})
	})

	t.Run("creates_on_archived_repository_with_error)", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandString(5)
		repoName := fmt.Sprintf("%scode-scanning-default-%s", testResourcePrefix, randomID)

		repoConfig := `
			resource "github_repository" "test" {
				name       = "%s"
				visibility = "private"
				auto_init  = true
				archived   = %t
			}
			%s
		`

		defaultSetupBlock := `
			resource "github_code_scanning_default_setup" "test" {
				repository = github_repository.test.name
			}
		`

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(repoConfig, repoName, false, ""),
				},
				{
					Config: fmt.Sprintf(repoConfig, repoName, true, ""),
				},
				{
					Config:      fmt.Sprintf(repoConfig, repoName, true, defaultSetupBlock),
					ExpectError: regexp.MustCompile(`cannot configure code scanning default setup on archived repository`),
				},
			},
		})
	})
}
