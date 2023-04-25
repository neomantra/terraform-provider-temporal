package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	temporalEnums "go.temporal.io/api/enums/v1"
	temporalClient "go.temporal.io/sdk/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ScheduleResource{}
var _ resource.ResourceWithImportState = &ScheduleResource{}

func NewScheduleResource() resource.Resource {
	return &ScheduleResource{}
}

// ScheduleResource defines the resource implementation.
type ScheduleResource struct {
	tclient temporalClient.Client
}

// ScheduleResourceModel describes the resource data model.
type ScheduleResourceModel struct {
	ScheduleId types.String `tfsdk:"id"`
	DescJson   types.String `tfsdk:"desc"`
}

func (r *ScheduleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (r *ScheduleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Scheduled Workflow resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Schedule ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"desc": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Schedule description in JSON",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ScheduleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	tclient, ok := req.ProviderData.(temporalClient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected temporalClient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.tclient = tclient
}

func (r *ScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data *ScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleHandle, err := r.tclient.ScheduleClient().Create(ctx, temporalClient.ScheduleOptions{
		ID: data.ScheduleId.ValueString(),
		// TODO: we must express all this in Terraform!
		// Spec: &temporalClient.ScheduleSpec{
		// 	Intervals: []temporalClient.ScheduleIntervalSpec{
		// 		{Every: 10},
		// 	},
		// },
		Action: &temporalClient.ScheduleWorkflowAction{
			Workflow: "foo",
			//Args:      {"foo"},
			ID:        "some-id-workflow",
			TaskQueue: "queue",
		},
		Overlap: temporalEnums.SCHEDULE_OVERLAP_POLICY_SKIP,
	})
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("Create: Unable to create Schedule: %s", err))
		return
	}

	data.ScheduleId = types.StringValue(scheduleHandle.GetID())

	desc, err := scheduleHandle.Describe(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("Create: Unable to describe Schedule after create: %s", err))
		return
	}
	jsonBytes, err := json.Marshal(desc)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("Create: Unable to marshal Schedule description after create: %s", err))
		return
	}
	data.DescJson = basetypes.NewStringValue(string(jsonBytes))
	tflog.Trace(ctx, fmt.Sprintf("Created Schedule resource %s", data.ScheduleId.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var data *ScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the Schedule's description from the Server
	desc, err := r.tclient.ScheduleClient().GetHandle(ctx, data.ScheduleId.ValueString()).Describe(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("Read: Unable to describe Schedule %s : %s", data.ScheduleId.ValueString(), err))
		return
	}
	jsonBytes, err := json.Marshal(desc)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("Read: Unable to marshal ScheduledWorkflow description after Describe: %s", err))
		return
	}
	data.DescJson = basetypes.NewStringValue(string(jsonBytes))
	tflog.Trace(ctx, fmt.Sprintf("Read ScheduledWorkflow resource %s", data.ScheduleId.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data *ScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform prior state data into the model
	var data *ScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the Schedule from the Server
	err := r.tclient.ScheduleClient().GetHandle(ctx, data.ScheduleId.ValueString()).Delete(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("Delete: Unable to delete Schedule : %s", err))
	}
	tflog.Trace(ctx, fmt.Sprintf("Deleted ScheduledWorkflow resource %s", data.ScheduleId.ValueString()))
}

func (r *ScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
