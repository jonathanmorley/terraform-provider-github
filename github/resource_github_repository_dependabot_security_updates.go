package github

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/go-github/v89/github"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubRepositoryDependabotSecurityUpdates() *schema.Resource {
	return &schema.Resource{
		Create: resourceGithubRepositoryDependabotSecurityUpdatesCreateOrUpdate,
		Read:   resourceGithubRepositoryDependabotSecurityUpdatesRead,
		Update: resourceGithubRepositoryDependabotSecurityUpdatesCreateOrUpdate,
		Delete: resourceGithubRepositoryDependabotSecurityUpdatesDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGithubRepositoryDependabotSecurityUpdatesImport,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The GitHub repository.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "The state of the automated security fixes.",
			},
		},
	}
}

func resourceGithubRepositoryDependabotSecurityUpdatesCreateOrUpdate(d *schema.ResourceData, meta any) error {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)
	enabled := d.Get("enabled").(bool)

	ctx := context.Background()
	var err error
	if enabled {
		_, err = client.Repositories.EnableAutomatedSecurityFixes(ctx, owner, repoName)
	} else {
		_, err = client.Repositories.DisableAutomatedSecurityFixes(ctx, owner, repoName)
	}

	if err != nil {
		return err
	}
	d.SetId(repoName)
	return resourceGithubRepositoryDependabotSecurityUpdatesRead(d, meta)
}

func resourceGithubRepositoryDependabotSecurityUpdatesRead(d *schema.ResourceData, meta any) error {
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	repoName := d.Get("repository").(string)

	ctx := context.Background()

	p, resp, err := client.Repositories.GetAutomatedSecurityFixes(ctx, orgName, repoName)
	if err != nil {
		// GetAutomatedSecurityFixes maps a 404 to an error, so a 404 here is
		// ambiguous: automated security fixes may simply be disabled, or the
		// repository may be gone or archived. Look the repository up to decide
		// whether to drop the resource from state.
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			repo, _, err := client.Repositories.Get(ctx, orgName, repoName)
			if err != nil {
				if ghErr, ok := errors.AsType[*github.ErrorResponse](err); ok && ghErr.Response.StatusCode == http.StatusNotFound {
					tflog.Warn(ctx, "repository no longer exists, removing dependabot security updates from state", map[string]any{"owner": orgName, "repo_name": repoName})
					d.SetId("")
					return nil
				}
				return err
			}
			if repo.GetArchived() {
				tflog.Warn(ctx, "repository is archived, removing dependabot security updates from state", map[string]any{"owner": orgName, "repo_name": repoName})
				d.SetId("")
				return nil
			}
		} else {
			return err
		}
	}

	if p != nil {
		_ = d.Set("enabled", p.GetEnabled())
	} else {
		_ = d.Set("enabled", false)
	}

	return nil
}

func resourceGithubRepositoryDependabotSecurityUpdatesDelete(d *schema.ResourceData, meta any) error {
	client := meta.(*Owner).v3client
	orgName := meta.(*Owner).name
	repoName := d.Get("repository").(string)

	ctx := context.Background()

	_, err := client.Repositories.DisableAutomatedSecurityFixes(ctx, orgName, repoName)
	if err != nil {
		return err
	}

	return resourceGithubRepositoryDependabotSecurityUpdatesRead(d, meta)
}

func resourceGithubRepositoryDependabotSecurityUpdatesImport(d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	repoName := d.Id()

	_ = d.Set("repository", repoName)

	err := resourceGithubRepositoryDependabotSecurityUpdatesRead(d, meta)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
