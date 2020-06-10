package tfe

import (
	"fmt"
	"log"
	"strings"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/ryboe/q"
)

func resourceTFETeamAccess() *schema.Resource {
	return &schema.Resource{
		Create: resourceTFETeamAccessCreate,
		Read:   resourceTFETeamAccessRead,
		Update: resourceTFETeamAccessUpdate,
		Delete: resourceTFETeamAccessDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTFETeamAccessImporter,
		},

		CustomizeDiff: updateComputedAttributes,
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceTfeTeamAccessResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceTfeTeamAccessStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"access": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.AccessAdmin),
						string(tfe.AccessRead),
						string(tfe.AccessPlan),
						string(tfe.AccessWrite),
						string(tfe.AccessCustom),
					},
					false,
				),
			},

			"runs": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.RunsPermissionRead),
						string(tfe.RunsPermissionPlan),
						string(tfe.RunsPermissionApply),
					},
					false,
				),
			},

			"variables": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.VariablesPermissionNone),
						string(tfe.VariablesPermissionRead),
						string(tfe.VariablesPermissionWrite),
					},
					false,
				),
			},

			"state_versions": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.StateVersionsPermissionNone),
						string(tfe.StateVersionsPermissionReadOutputs),
						string(tfe.StateVersionsPermissionRead),
						string(tfe.StateVersionsPermissionWrite),
					},
					false,
				),
			},

			"sentinel_mocks": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						string(tfe.SentinelMocksPermissionNone),
						string(tfe.SentinelMocksPermissionRead),
					},
					false,
				),
			},

			"workspace_locking": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"team_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					workspaceIdRegexp,
					"must be the workspace's external_id",
				),
			},
		},
	}
}

func resourceTFETeamAccessCreate(d *schema.ResourceData, meta interface{}) error {
	q.Q("Running CREATE action")
	tfeClient := meta.(*tfe.Client)

	// Get the access level
	access := d.Get("access").(string)

	// Get the workspace
	workspaceID := d.Get("workspace_id").(string)
	ws, err := tfeClient.Workspaces.ReadByID(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf(
			"Error retrieving workspace %s: %v", workspaceID, err)
	}

	// Get the team.
	teamID := d.Get("team_id").(string)
	tm, err := tfeClient.Teams.Read(ctx, teamID)
	if err != nil {
		return fmt.Errorf("Error retrieving team %s: %v", teamID, err)
	}

	// Create a new options struct.
	options := tfe.TeamAccessAddOptions{
		Access:    tfe.Access(tfe.AccessType(access)),
		Team:      tm,
		Workspace: ws,
	}

	if access == "custom" {
		if v, ok := d.GetOk("runs"); ok {
			options.Runs = tfe.RunsPermission(tfe.RunsPermissionType(v.(string)))
		}
		if v, ok := d.GetOk("variables"); ok {
			options.Variables = tfe.VariablesPermission(tfe.VariablesPermissionType(v.(string)))
		}
		if v, ok := d.GetOk("state_versions"); ok {
			options.StateVersions = tfe.StateVersionsPermission(tfe.StateVersionsPermissionType(v.(string)))
		}
		if v, ok := d.GetOk("sentinel_mocks"); ok {
			options.SentinelMocks = tfe.SentinelMocksPermission(tfe.SentinelMocksPermissionType(v.(string)))
		}
		if v, ok := d.GetOk("workspace_locking"); ok {
			options.WorkspaceLocking = tfe.Bool(v.(bool))
		}
	}

	log.Printf("[DEBUG] Give team %s %s access to workspace: %s", tm.Name, access, ws.Name)
	tmAccess, err := tfeClient.TeamAccess.Add(ctx, options)
	if err != nil {
		return fmt.Errorf(
			"Error giving team %s %s access to workspace %s: %v", tm.Name, access, ws.Name, err)
	}

	d.SetId(tmAccess.ID)

	return resourceTFETeamAccessRead(d, meta)
}

func resourceTFETeamAccessRead(d *schema.ResourceData, meta interface{}) error {
	q.Q("Running READ action")
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Read configuration of team access: %s", d.Id())
	tmAccess, err := tfeClient.TeamAccess.Read(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			log.Printf("[DEBUG] Team access %s does no longer exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading configuration of team access %s: %v", d.Id(), err)
	}

	// Update config.
	d.Set("access", string(tmAccess.Access))
	d.Set("runs", string(tmAccess.Runs))
	d.Set("variables", string(tmAccess.Variables))
	d.Set("state_versions", string(tmAccess.StateVersions))
	d.Set("sentinel_mocks", string(tmAccess.SentinelMocks))
	d.Set("workspace_locking", bool(tmAccess.WorkspaceLocking))

	if tmAccess.Team != nil {
		d.Set("team_id", tmAccess.Team.ID)
	} else {
		d.Set("team_id", "")
	}

	return nil
}

func resourceTFETeamAccessUpdate(d *schema.ResourceData, meta interface{}) error {
	q.Q("Running UPDATE action")
	tfeClient := meta.(*tfe.Client)

	// create an options struct
	options := tfe.TeamAccessUpdateOptions{}

	if d.HasChange("access") {
		if v, ok := d.GetOk("access"); ok {
			access := tfe.AccessType(v.(string))
			options.Access = tfe.Access(access)
		}
	}

	if d.HasChange("runs") {
		if v, ok := d.GetOk("runs"); ok {
			options.Runs = tfe.RunsPermission(tfe.RunsPermissionType(v.(string)))
		}
	}

	if d.HasChange("variables") {
		if v, ok := d.GetOk("variables"); ok {
			options.Variables = tfe.VariablesPermission(tfe.VariablesPermissionType(v.(string)))
		}
	}

	if d.HasChange("state_versions") {
		if v, ok := d.GetOk("state_versions"); ok {
			options.StateVersions = tfe.StateVersionsPermission(tfe.StateVersionsPermissionType(v.(string)))
		}
	}

	if d.HasChange("sentinel_mocks") {
		if v, ok := d.GetOk("sentinel_mocks"); ok {
			options.SentinelMocks = tfe.SentinelMocksPermission(tfe.SentinelMocksPermissionType(v.(string)))
		}
	}

	if d.HasChange("workspace_locking") {
		if v, ok := d.GetOk("workspace_locking"); ok {
			options.WorkspaceLocking = tfe.Bool(v.(bool))
		}
	}

	log.Printf("[DEBUG] Update team access: %s", d.Id())
	tmAccess, err := tfeClient.TeamAccess.Update(ctx, d.Id(), options)
	if err != nil {
		return fmt.Errorf(
			"Error updating team access %s: %v", d.Id(), err)
	}

	d.Set("runs", string(tmAccess.Runs))
	d.Set("variables", string(tmAccess.Variables))
	d.Set("state_versions", string(tmAccess.StateVersions))
	d.Set("sentinel_mocks", string(tmAccess.SentinelMocks))
	d.Set("workspace_locking", bool(tmAccess.WorkspaceLocking))

	return nil
}

func resourceTFETeamAccessDelete(d *schema.ResourceData, meta interface{}) error {
	q.Q("Running DELETE action")
	tfeClient := meta.(*tfe.Client)

	log.Printf("[DEBUG] Delete team access: %s", d.Id())
	err := tfeClient.TeamAccess.Remove(ctx, d.Id())
	if err != nil {
		if err == tfe.ErrResourceNotFound {
			return nil
		}
		return fmt.Errorf("Error deleting team access %s: %v", d.Id(), err)
	}

	return nil
}

func resourceTFETeamAccessImporter(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tfeClient := meta.(*tfe.Client)

	s := strings.SplitN(d.Id(), "/", 3)
	if len(s) != 3 {
		return nil, fmt.Errorf(
			"invalid team access import format: %s (expected <ORGANIZATION>/<WORKSPACE>/<TEAM ACCESS ID>)",
			d.Id(),
		)
	}

	// Set the fields that are part of the import ID.
	workspace_id, err := fetchWorkspaceExternalID(s[0]+"/"+s[1], tfeClient)
	if err != nil {
		return nil, fmt.Errorf(
			"error retrieving workspace %s from organization %s: %v", s[0], s[1], err)
	}
	d.Set("workspace_id", workspace_id)
	d.SetId(s[2])

	return []*schema.ResourceData{d}, nil
}

func updateComputedAttributes(d *schema.ResourceDiff, meta interface{}) error {
	if d.HasChange("access") {
		old, new := d.GetChange("access")

		toCustomAccess := new.(string) == "custom"
		fromCustomAccess := old.(string) == "custom"

		if toCustomAccess || fromCustomAccess {
			d.ForceNew("access")
		}
	}

	return nil
}
