package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/go-github/v89/github"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceGithubCodeScanningDefaultSetup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGithubCodeScanningDefaultSetupCreate,
		ReadContext:   resourceGithubCodeScanningDefaultSetupRead,
		UpdateContext: resourceGithubCodeScanningDefaultSetupUpdate,
		DeleteContext: resourceGithubCodeScanningDefaultSetupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGithubCodeScanningDefaultSetupImport,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The repository name to configure code scanning default setup for.",
			},
			"repository_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the repository to configure code scanning default setup for.",
			},
			"query_suite": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "default",
				Description:      "The query suite to use for code scanning default setup. Can be one of 'default' or 'extended'.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"default", "extended"}, false)),
			},
			"languages": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "The languages to configure code scanning default setup for. Omit to let GitHub auto-detect them.",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(codeScanningDefaultSetupLanguages, false)),
				},
				Set: schema.HashString,
			},
		},

		CustomizeDiff: diffRepository,
	}
}

var codeScanningDefaultSetupLanguages = []string{
	"actions",
	"c-cpp",
	"csharp",
	"go",
	"java-kotlin",
	"javascript-typescript",
	"python",
	"ruby",
	"swift",
}

func resourceGithubCodeScanningDefaultSetupCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	tflog.Info(ctx, "Creating code scanning default setup", map[string]any{"id": d.Id()})
	meta, _ := m.(*Owner)
	client := meta.v3client

	owner := meta.name
	repoName := d.Get("repository").(string)

	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return diag.FromErr(err)
	}
	if repo.GetArchived() {
		return diag.Errorf("cannot configure code scanning default setup on archived repository %s/%s", owner, repoName)
	}

	if err := configureCodeScanningDefaultSetup(ctx, d, meta, "configured"); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(int(repo.GetID())))

	if err = d.Set("repository_id", repo.GetID()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGithubCodeScanningDefaultSetupRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	tflog.Info(ctx, "Reading code scanning default setup", map[string]any{"id": d.Id()})
	meta, _ := m.(*Owner)
	client := meta.v3client

	owner := meta.name
	repoName := d.Get("repository").(string)

	cfg, resp, err := client.CodeScanning.GetDefaultSetupConfiguration(ctx, owner, repoName)
	if err != nil {
		// The default setup endpoint itself always returns the current
		// configuration, even when it is not configured, so a 404 here means
		// the repository is gone or archived. Look the repository up to decide
		// whether to drop the resource from state.
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			repo, _, err := client.Repositories.Get(ctx, owner, repoName)
			if err != nil {
				if ghErr, ok := errors.AsType[*github.ErrorResponse](err); ok && ghErr.Response.StatusCode == http.StatusNotFound {
					tflog.Warn(ctx, "repository no longer exists, removing code scanning default setup from state", map[string]any{"owner": owner, "repo_name": repoName})
					d.SetId("")
					return nil
				}
				return diag.FromErr(err)
			}
			if repo.GetArchived() {
				tflog.Warn(ctx, "repository is archived, removing code scanning default setup from state", map[string]any{"owner": owner, "repo_name": repoName})
				d.SetId("")
				return nil
			}
			tflog.Warn(ctx, "code scanning default setup not found, removing it from state", map[string]any{"owner": owner, "repo_name": repoName})
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading code scanning default setup: %s", err.Error())
	}

	if cfg.GetState() != "configured" {
		tflog.Warn(ctx, "code scanning default setup is not configured, removing it from state", map[string]any{"owner": owner, "repo_name": repoName})
		d.SetId("")
		return nil
	}
	tflog.Debug(ctx, "Setting code scanning default setup state", map[string]any{"owner": owner, "repo_name": repoName})

	if err = d.Set("query_suite", cfg.GetQuerySuite()); err != nil {
		return diag.FromErr(err)
	}

	languages := make([]any, 0, len(cfg.Languages))
	for _, language := range cfg.Languages {
		languages = append(languages, language)
	}
	// The API may report detected languages outside of the documented enum
	// (e.g. "kotlin"), so record whatever is returned.
	if err = d.Set("languages", languages); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGithubCodeScanningDefaultSetupUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	tflog.Info(ctx, "Updating code scanning default setup", map[string]any{"id": d.Id()})
	meta, _ := m.(*Owner)

	if err := configureCodeScanningDefaultSetup(ctx, d, meta, "configured"); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGithubCodeScanningDefaultSetupDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	tflog.Info(ctx, "Deleting code scanning default setup", map[string]any{"id": d.Id()})
	meta, _ := m.(*Owner)

	owner := meta.name
	repoName := d.Get("repository").(string)

	err := configureCodeScanningDefaultSetup(ctx, d, meta, "not-configured")
	if err != nil {
		return diag.FromErr(handleArchivedRepoDelete(err, "repository code scanning default setup", d.Id(), owner, repoName))
	}

	return nil
}

func configureCodeScanningDefaultSetup(ctx context.Context, d *schema.ResourceData, meta *Owner, state string) error {
	client := meta.v3client
	owner := meta.name
	repoName := d.Get("repository").(string)

	querySuite, ok := d.Get("query_suite").(string)
	if !ok {
		return fmt.Errorf("unexpected type for attribute \"query_suite\": %T", d.Get("query_suite"))
	}
	opts := &github.UpdateDefaultSetupConfigurationOptions{
		State:      state,
		QuerySuite: &querySuite,
	}
	if v, ok := d.GetOk("languages"); ok {
		set := v.(*schema.Set)
		opts.Languages = make([]string, 0, set.Len())
		for _, language := range set.List() {
			opts.Languages = append(opts.Languages, language.(string))
		}
	}

	_, _, err := client.CodeScanning.UpdateDefaultSetupConfiguration(ctx, owner, repoName, opts)
	if err != nil {
		// Updating default setup is asynchronous: GitHub answers a 202 Accepted
		// once it has scheduled the configuration change.
		if _, ok := errors.AsType[*github.AcceptedError](err); !ok {
			return err
		}
	}

	return nil
}

func resourceGithubCodeScanningDefaultSetupImport(ctx context.Context, d *schema.ResourceData, m any) ([]*schema.ResourceData, error) {
	tflog.Debug(ctx, "Importing code scanning default setup", map[string]any{"id": d.Id()})
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
