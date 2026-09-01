package github

import (
	"context"

	"github.com/google/go-github/v89/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGithubRepositoryRulesets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGithubRepositoryRulesetsRead,

		Schema: map[string]*schema.Schema{
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rulesets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"source_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"source": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enforcement": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"node_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubRepositoryRulesetsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*Owner).v3client
	owner := meta.(*Owner).name
	repoName := d.Get("repository").(string)

	// Only include repository-scoped rulesets (not those inherited from a
	// parent organization or enterprise) so that consumers can use the
	// returned ruleset IDs to import repo rulesets.
	opts := &github.RepositoryListRulesetsOptions{
		IncludesParents: github.Bool(false),
	}

	results := make([]map[string]any, 0)
	for {
		rulesets, resp, err := client.Repositories.GetAllRulesets(ctx, owner, repoName, opts)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, ruleset := range rulesets {
			target := ""
			if t := ruleset.GetTarget(); t != nil {
				target = string(*t)
			}
			sourceType := ""
			if st := ruleset.GetSourceType(); st != nil {
				sourceType = string(*st)
			}
			result := map[string]any{
				"id":          ruleset.GetID(),
				"name":        ruleset.GetName(),
				"target":      target,
				"source_type": sourceType,
				"source":      ruleset.GetSource(),
				"enforcement": string(ruleset.GetEnforcement()),
				"node_id":     ruleset.GetNodeID(),
			}
			results = append(results, result)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	d.SetId(repoName)
	if err := d.Set("rulesets", results); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
