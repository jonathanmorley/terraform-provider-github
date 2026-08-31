package github

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccGithubRepositoryImmutableReleases(t *testing.T) {
	t.Parallel()

	t.Run("creates_immutable_releases_as_enabled_without_error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandString(5)
		repoName := fmt.Sprintf("%simmutable-releases-%s", testResourcePrefix, randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name       = "%s"
				visibility = "private"
				auto_init  = true
			}

			resource "github_repository_immutable_releases" "test" {
				repository  = github_repository.test.name
				enabled     = true
			}
		`, repoName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("github_repository_immutable_releases.test", tfjsonpath.New("enabled"), knownvalue.Bool(true)),
						statecheck.CompareValuePairs("github_repository_immutable_releases.test", tfjsonpath.New("repository"), "github_repository.test", tfjsonpath.New("name"), compare.ValuesSame()),
						statecheck.ExpectKnownValue("github_repository_immutable_releases.test", tfjsonpath.New("repository_id"), knownvalue.NotNull()),
					},
				},
			},
		})
	})

	t.Run("creates_immutable_releases_as_disabled_without_error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandString(5)
		repoName := fmt.Sprintf("%simmutable-releases-%s", testResourcePrefix, randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name       = "%s"
				visibility = "private"
				auto_init  = true
			}

			resource "github_repository_immutable_releases" "test" {
				repository  = github_repository.test.name
				enabled     = false
			}
		`, repoName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("github_repository_immutable_releases.test", tfjsonpath.New("enabled"), knownvalue.Bool(false)),
					},
				},
			},
		})
	})

	t.Run("updates_immutable_releases_without_error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandString(5)
		repoName := fmt.Sprintf("%simmutable-releases-%s", testResourcePrefix, randomID)

		config := `
			resource "github_repository" "test" {
				name       = "%s"
				visibility = "private"
				auto_init  = true
			}

			resource "github_repository_immutable_releases" "test" {
				repository  = github_repository.test.name
				enabled     = %t
			}
		`

		compareValuesDiffer := statecheck.CompareValue(compare.ValuesDiffer())

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(config, repoName, false),
					ConfigStateChecks: []statecheck.StateCheck{
						compareValuesDiffer.AddStateValue("github_repository_immutable_releases.test", tfjsonpath.New("enabled")),
					},
				},
				{
					Config: fmt.Sprintf(config, repoName, true),
					ConfigStateChecks: []statecheck.StateCheck{
						compareValuesDiffer.AddStateValue("github_repository_immutable_releases.test", tfjsonpath.New("enabled")),
					},
				},
				{
					Config: fmt.Sprintf(config, repoName, false),
					ConfigStateChecks: []statecheck.StateCheck{
						compareValuesDiffer.AddStateValue("github_repository_immutable_releases.test", tfjsonpath.New("enabled")),
					},
				},
			},
		})
	})

	t.Run("imports_immutable_releases_without_error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandString(5)
		repoName := fmt.Sprintf("%simmutable-releases-%s", testResourcePrefix, randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name       = "%s"
				visibility = "private"
				auto_init  = true
			}

			resource "github_repository_immutable_releases" "test" {
				repository  = github_repository.test.name
				enabled     = true
			}
		`, repoName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("github_repository_immutable_releases.test", tfjsonpath.New("repository_id"), knownvalue.NotNull()),
					},
				},
				{
					ResourceName:      "github_repository_immutable_releases.test",
					ImportState:       true,
					ImportStateId:     repoName,
					ImportStateVerify: true,
				},
			},
		})
	})

	t.Run("creates_with_defaults_without_error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandString(5)
		repoName := fmt.Sprintf("%simmutable-releases-%s", testResourcePrefix, randomID)

		config := fmt.Sprintf(`
			resource "github_repository" "test" {
				name       = "%s"
				visibility = "private"
				auto_init  = true
			}

			resource "github_repository_immutable_releases" "test" {
				repository = github_repository.test.name
			}
		`, repoName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("github_repository_immutable_releases.test", tfjsonpath.New("enabled"), knownvalue.Bool(true)),
					},
				},
			},
		})
	})

	t.Run("removes_from_state_when_repository_deleted", func(t *testing.T) {
		t.Parallel()

		repo := mustCreateTestRepository(t)
		repoName := repo.GetName()

		config := fmt.Sprintf(`
			resource "github_repository_immutable_releases" "test" {
				repository = "%s"
				enabled    = true
			}
		`, repoName)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
				},
				{
					// Delete the repository out of band. On refresh, the immutable
					// releases Read hits a deleted repository and must remove the
					// resource from state instead of erroring, leaving a non-empty
					// plan (Terraform wants to recreate it).
					PreConfig: func() {
						if _, err := testAccConf.meta.v3client.Repositories.Delete(t.Context(), testAccConf.meta.name, repoName); err != nil {
							t.Errorf("failed to delete repository out-of-band: %s", err)
						}
					},
					RefreshState:       true,
					ExpectNonEmptyPlan: true,
					RefreshPlanChecks: resource.RefreshPlanChecks{
						PostRefresh: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("github_repository_immutable_releases.test", plancheck.ResourceActionCreate),
						},
					},
				},
			},
		})
	})

	t.Run("creates_on_archived_repository_with_error)", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandString(5)
		repoName := fmt.Sprintf("%simmutable-releases-%s", testResourcePrefix, randomID)

		repoConfig := `
			resource "github_repository" "test" {
				name       = "%s"
				visibility = "private"
				auto_init  = true
				archived   = %t
			}
			%s
		`

		immutableReleasesBlock := `
			resource "github_repository_immutable_releases" "test" {
				repository = github_repository.test.name
				enabled    = true
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
					Config:      fmt.Sprintf(repoConfig, repoName, true, immutableReleasesBlock),
					ExpectError: regexp.MustCompile(`cannot enable immutable releases on archived repository`),
				},
			},
		})
	})
}
