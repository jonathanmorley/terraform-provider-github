package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-github/v89/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGithubRepositoryDependabotSecurityUpdates(t *testing.T) {
	t.Parallel()

	t.Run("enables automated security fixes without error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("%srepo-depbot-updates-%s", testResourcePrefix, randomID)
		enabled := "enabled = false"
		updatedEnabled := "enabled = true"
		config := fmt.Sprintf(`

			resource "github_repository" "test" {
				name = "%s"
				visibility = "private"
			  	auto_init = true
				vulnerability_alerts   = true
			}


			resource "github_repository_dependabot_security_updates" "test" {
			  repository  = github_repository.test.id
			  %s
			}
		`, repoName, enabled)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_dependabot_security_updates.test", "enabled",
					"false",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_dependabot_security_updates.test", "enabled",
					"true",
				),
			),
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks["before"],
				},
				{
					Config: strings.Replace(config,
						enabled,
						updatedEnabled, 1),
					Check: checks["after"],
				},
			},
		})
	})

	t.Run("disables automated security fixes without error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("%srepo-depbot-updates-%s", testResourcePrefix, randomID)
		enabled := "enabled = true"
		updatedEnabled := "enabled = false"

		config := fmt.Sprintf(`

			resource "github_repository" "test" {
				name = "%s"
				visibility = "private"
			  	auto_init = true
				vulnerability_alerts   = true
			}


			resource "github_repository_dependabot_security_updates" "test" {
			  repository  = github_repository.test.id
			  %s
			}
		`, repoName, enabled)

		checks := map[string]resource.TestCheckFunc{
			"before": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_dependabot_security_updates.test", "enabled",
					"true",
				),
			),
			"after": resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(
					"github_repository_dependabot_security_updates.test", "enabled",
					"false",
				),
			),
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  checks["before"],
				},
				{
					Config: strings.Replace(config,
						enabled,
						updatedEnabled, 1),
					Check: checks["after"],
				},
			},
		})
	})

	t.Run("imports automated security fixes without error", func(t *testing.T) {
		t.Parallel()

		randomID := acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)
		repoName := fmt.Sprintf("%srepo-depbot-updates-%s", testResourcePrefix, randomID)
		config := fmt.Sprintf(`
			resource "github_repository" "test" {
			  name = "%s"
			  vulnerability_alerts   = true
			}

			resource "github_repository_dependabot_security_updates" "test" {
			  repository  = github_repository.test.id
			  enabled = false
			}
    `, repoName)

		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet("github_repository_dependabot_security_updates.test", "repository"),
			resource.TestCheckResourceAttrSet("github_repository_dependabot_security_updates.test", "enabled"),
			resource.TestCheckResourceAttr("github_repository_dependabot_security_updates.test", "enabled", "false"),
		)

		resource.Test(t, resource.TestCase{
			PreCheck:          func() { skipUnauthenticated(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: config,
					Check:  check,
				},
				{
					ResourceName:      "github_repository_dependabot_security_updates.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func Test_resourceGithubRepositoryDependabotSecurityUpdatesRead(t *testing.T) {
	t.Parallel()

	newTestResourceData := func(t *testing.T, securityFixesStatus int, repositoryJSON string) (*schema.ResourceData, *Owner) {
		t.Helper()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/repos/owner/repo/automated-security-fixes":
				w.WriteHeader(securityFixesStatus)
				if securityFixesStatus == http.StatusOK {
					fmt.Fprint(w, `{"enabled":true}`)
					return
				}
				fmt.Fprint(w, `{}`)
			case "/repos/owner/repo":
				if repositoryJSON == "" {
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, `{"message":"Not Found"}`)
					return
				}
				fmt.Fprint(w, repositoryJSON)
			default:
				http.NotFound(w, r)
			}
		}))
		t.Cleanup(srv.Close)

		baseURL := srv.URL + "/"
		client, err := github.NewClient(github.WithHTTPClient(srv.Client()), github.WithURLs(&baseURL, nil))
		if err != nil {
			t.Fatalf("failed to create test client: %s", err)
		}

		d := schema.TestResourceDataRaw(t, resourceGithubRepositoryDependabotSecurityUpdates().Schema, map[string]any{"repository": "repo"})
		d.SetId("repo")

		return d, &Owner{name: "owner", v3client: client}
	}

	t.Run("reads_enabled_state", func(t *testing.T) {
		t.Parallel()

		d, meta := newTestResourceData(t, http.StatusOK, `{"enabled":true}`)

		if err := resourceGithubRepositoryDependabotSecurityUpdatesRead(d, meta); err != nil {
			t.Fatalf("unexpected error reading automated security fixes: %s", err)
		}

		if got := d.Get("enabled").(bool); !got {
			t.Errorf("expected enabled to be true, got %v", got)
		}
	})

	t.Run("treats_404_as_disabled_when_repository_exists", func(t *testing.T) {
		t.Parallel()

		d, meta := newTestResourceData(t, http.StatusNotFound, `{"id":1,"name":"repo","archived":false}`)

		if err := resourceGithubRepositoryDependabotSecurityUpdatesRead(d, meta); err != nil {
			t.Fatalf("unexpected error reading disabled automated security fixes: %s", err)
		}

		if d.Id() != "repo" {
			t.Errorf("expected resource to stay in state, got id %q", d.Id())
		}

		if got := d.Get("enabled").(bool); got {
			t.Errorf("expected enabled to be false, got %v", got)
		}
	})

	t.Run("removes_state_when_repository_deleted", func(t *testing.T) {
		t.Parallel()

		d, meta := newTestResourceData(t, http.StatusNotFound, "")

		if err := resourceGithubRepositoryDependabotSecurityUpdatesRead(d, meta); err != nil {
			t.Fatalf("unexpected error reading deleted repository: %s", err)
		}

		if d.Id() != "" {
			t.Errorf("expected resource to be removed from state, got id %q", d.Id())
		}
	})

	t.Run("removes_state_when_repository_archived", func(t *testing.T) {
		t.Parallel()

		d, meta := newTestResourceData(t, http.StatusNotFound, `{"id":1,"name":"repo","archived":true}`)

		if err := resourceGithubRepositoryDependabotSecurityUpdatesRead(d, meta); err != nil {
			t.Fatalf("unexpected error reading archived repository: %s", err)
		}

		if d.Id() != "" {
			t.Errorf("expected resource to be removed from state, got id %q", d.Id())
		}
	})
}
