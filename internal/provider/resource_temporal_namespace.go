package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neomantra/terraform-provider-temporal/internal/tfschema"

	temporalNamespace "go.temporal.io/api/namespace/v1"
	temporalWorkflowService "go.temporal.io/api/workflowservice/v1"
	temporalClient "go.temporal.io/sdk/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NamespaceResource{}
var _ resource.ResourceWithImportState = &NamespaceResource{}

func NewNamespaceResource() resource.Resource {
	return &NamespaceResource{}
}

// NamespaceResource defines the resource implementation.
type NamespaceResource struct {
	nsclient temporalClient.NamespaceClient
}

//////////////////////////////////////////////////////////////////////////////

func (r *NamespaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (r *NamespaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tfschema.GetResourceNamespaceSchema()
}

func (r *NamespaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	tdata, ok := req.ProviderData.(TemporalProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("NamespaceResource expected temporalClient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.nsclient = tdata.nsclient
}

//////////////////////////////////////////////////////////////////////////////

func (r *NamespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform PLAN data into the models
	var plan *tfschema.NamespaceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the Temporal Namespaces
	retentionDuration := time.Duration(plan.RetentionHours.ValueInt64()) * time.Hour
	nsRequest := temporalWorkflowService.RegisterNamespaceRequest{
		Namespace:                        plan.Name.ValueString(),
		Description:                      plan.Description.ValueString(),
		OwnerEmail:                       plan.OwnerEmail.ValueString(),
		WorkflowExecutionRetentionPeriod: &retentionDuration,
		// Clusters:                         []*replication.ClusterReplicationConfig{},
		// ActiveClusterName:                "",
		// Data:                             map[string]string{},
		// SecurityToken:                    "",
		// IsGlobalNamespace:                false,
		// HistoryArchivalState:             0,
		// HistoryArchivalUri:               "",
		// VisibilityArchivalState:          0,
		// VisibilityArchivalUri:            "",
	}
	err := r.nsclient.Register(ctx, &nsRequest)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("NamespaceResource Create: Unable to register Namespace %s : %s", plan.Name.ValueString(), err))
		return
	}

	// Fetch the Namespaces's description from the Server
	desc, err := r.nsclient.Describe(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("NamespaceResource Create: Unable to describe Namespace %s : %s", plan.Name.ValueString(), err))
		return
	}

	plan.Id = plan.Name
	plan.Description = types.StringValue(desc.NamespaceInfo.Description)
	plan.OwnerEmail = types.StringValue(desc.NamespaceInfo.OwnerEmail)
	plan.RetentionHours = types.Int64Value(int64(desc.Config.WorkflowExecutionRetentionTtl.Hours()))

	// Write state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//////////////////////////////////////////////////////////////////////////////

func (r *NamespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var state *tfschema.NamespaceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the Namespaces's description from the Server
	// TODO: check for empty string or Terraform does these sanity checks?
	desc, err := r.nsclient.Describe(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("NamespaceResource Read: unable to describe Namespace %s : %s", state.Name.ValueString(), err))
		return
	}

	// Overwrite with refreshed state
	state.Name = types.StringValue(desc.NamespaceInfo.Name)
	state.Id = state.Name
	state.Description = types.StringValue(desc.NamespaceInfo.Description)
	state.OwnerEmail = types.StringValue(desc.NamespaceInfo.OwnerEmail)
	state.RetentionHours = types.Int64Value(int64(desc.Config.WorkflowExecutionRetentionTtl.Hours()))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//////////////////////////////////////////////////////////////////////////////

func (r *NamespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform PLAN data into the model
	var state *tfschema.NamespaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("NamespaceResource update resource %s", state.Name.ValueString()))

	retentionDuration := time.Duration(state.RetentionHours.ValueInt64()) * time.Hour

	err := r.nsclient.Update(ctx, &temporalWorkflowService.UpdateNamespaceRequest{
		Namespace: state.Name.ValueString(),
		UpdateInfo: &temporalNamespace.UpdateNamespaceInfo{
			Description: state.Description.ValueString(),
			OwnerEmail:  state.OwnerEmail.ValueString(),
		},
		Config: &temporalNamespace.NamespaceConfig{
			WorkflowExecutionRetentionTtl: &retentionDuration,
			// BadBinaries:                   &temporalNamespace.BadBinaries{},
			// HistoryArchivalState:          0,
			// HistoryArchivalUri:            "",
			// VisibilityArchivalState:       0,
			// VisibilityArchivalUri:         "",
			// CustomSearchAttributeAliases:  map[string]string{},
		},
		// ReplicationConfig: &replication.NamespaceReplicationConfig{},
		// SecurityToken:     "",
		// DeleteBadBinary:   "",
		// PromoteNamespace:  false,
	})
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("NamespaceResource Update: unable to update Namespace %s : %s", state.Name.ValueString(), err))
		return
	}

	// Fetch the Namespaces's description from the Server, to overwrite state
	// TODO: check for empty string or Terraform does these sanity checks?
	desc, err := r.nsclient.Describe(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("NamespaceResource Read: unable to describe Namespace %s : %s", state.Name.ValueString(), err))
		return
	}

	// Overwrite with refreshed state
	state.Name = types.StringValue(desc.NamespaceInfo.Name)
	state.Id = state.Name
	state.Description = types.StringValue(desc.NamespaceInfo.Description)
	state.OwnerEmail = types.StringValue(desc.NamespaceInfo.OwnerEmail)
	state.RetentionHours = types.Int64Value(int64(desc.Config.WorkflowExecutionRetentionTtl.Hours()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//////////////////////////////////////////////////////////////////////////////

func (r *NamespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform prior state data into the model
	var state *tfschema.NamespaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	tflog.Info(ctx, fmt.Sprintf("NamespaceResource Delete DOES NOTHING RIGHT NOW: %s", name))
}

//////////////////////////////////////////////////////////////////////////////

func (r *NamespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
