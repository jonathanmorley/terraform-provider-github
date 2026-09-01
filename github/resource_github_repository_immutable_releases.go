package github

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/go-github/v89/github"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGithubRepositoryImmutableReleases() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGithubRepositoryImmutableReleasesCreate,
		ReadContext:   resourceGithubRepositoryImmutableReleasesRead,
		UpdateContext: resourceGithubRepositoryImmutableReleasesUpdate,
		DeleteContext: resourceGithubRepositoryImmutableReleasesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGithubRepositoryImmutableReleasesImport,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The repository name to configure immutable releases for.",
			},
			"repository_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the repository to configure immutable releases for.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether immutable releases are enabled for the repository.",
			},
		},

		CustomizeDiff: diffRepository,
	}
}

func resourceGithubRepositoryImmutableReleasesCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	tflog.Info(ctx, "Creating repository immutable releases", map[string]any{"id": d.Id()})
	meta, _ := m.(*Owner)
	client := meta.v3client

	owner := meta.name
	repoName := d.Get("repository").(string)

	immutableReleasesEnabled := d.Get("enabled").(bool)
	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return diag.FromErr(err)
	}
	if repo.GetArchived() {
		return diag.Errorf("cannot enable immutable releases on archived repository %s/%s", owner, repoName)
	}

	if err := setImmutableReleases(ctx, client, owner, repoName, immutableReleasesEnabled); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(int(repo.GetID())))

	if err = d.Set("repository_id", repo.GetID()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGithubRepositoryImmutableReleasesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	tflog.Info(ctx, "Reading repository immutable releases", map[string]any{"id": d.Id()})
	meta, _ := m.(*Owner)
	client := meta.v3client

	owner := meta.name
	repoName := d.Get("repository").(string)

	// Getting a 404 here means immutable releases are NOT enabled on the repo,
	// which is a valid (disabled) state. This is distinct from the repository
	// itself missing or archived, which we detect by fetching the repo.
	status, resp, err := client.Repositories.AreImmutableReleasesEnabled(ctx, owner, repoName)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			repo, _, getErr := client.Repositories.Get(ctx, owner, repoName)
			if getErr != nil {
				if ghErr, ok := errors.AsType[*github.ErrorResponse](getErr); ok && ghErr.Response.StatusCode == http.StatusNotFound {
					tflog.Warn(ctx, "repository no longer exists, removing immutable releases from state", map[string]any{"owner": owner, "repo_name": repoName})
					d.SetId("")
					return nil
				}
				return diag.FromErr(getErr)
			}
			if repo.GetArchived() {
				tflog.Warn(ctx, "repository is archived, removing immutable releases from state", map[string]any{"owner": owner, "repo_name": repoName})
				d.SetId("")
				return nil
			}
			// Repository exists and immutable releases are simply disabled.
			if err := d.Set("enabled", false); err != nil {
				return diag.FromErr(err)
			}
			return nil
		}
		return diag.Errorf("error reading repository immutable releases: %s", err.Error())
	}

	enabled := status != nil && status.GetEnabled()
	tflog.Debug(ctx, "Setting immutable releases enabled state", map[string]any{"owner": owner, "repo_name": repoName, "immutable_releases_enabled": enabled})
	if err := d.Set("enabled", enabled); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGithubRepositoryImmutableReleasesUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	tflog.Info(ctx, "Updating repository immutable releases", map[string]any{"id": d.Id()})
	meta, _ := m.(*Owner)
	client := meta.v3client

	owner := meta.name
	repoName := d.Get("repository").(string)

	immutableReleasesEnabled := d.Get("enabled").(bool)
	if err := setImmutableReleases(ctx, client, owner, repoName, immutableReleasesEnabled); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGithubRepositoryImmutableReleasesDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	tflog.Info(ctx, "Deleting repository immutable releases", map[string]any{"id": d.Id()})
	meta, _ := m.(*Owner)
	client := meta.v3client

	owner := meta.name
	repoName := d.Get("repository").(string)
	_, err := client.Repositories.DisableImmutableReleases(ctx, owner, repoName)
	if err != nil {
		return diag.FromErr(handleArchivedRepoDelete(err, "repository immutable releases", d.Id(), owner, repoName))
	}

	return nil
}

func resourceGithubRepositoryImmutableReleasesImport(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
	tflog.Debug(ctx, "Importing repository immutable releases", map[string]any{"id": d.Id()})
	repoName := d.Id()
	if err := d.Set("repository", repoName); err != nil {
		return nil, err
	}

	meta, _ := m.(*Owner)
	owner := meta.name
	client := meta.v3client

	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, err
	}

	d.SetId(strconv.Itoa(int(repo.GetID())))

	if err = d.Set("repository_id", repo.GetID()); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func setImmutableReleases(ctx context.Context, client *github.Client, owner, repo string, enabled bool) error {
	if enabled {
		_, err := client.Repositories.EnableImmutableReleases(ctx, owner, repo)
		return err
	}
	_, err := client.Repositories.DisableImmutableReleases(ctx, owner, repo)
	return err
}
