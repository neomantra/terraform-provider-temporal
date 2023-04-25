package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	temporalClient "go.temporal.io/sdk/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ScheduleDataSource{}

func NewScheduleDataSource() datasource.DataSource {
	return &ScheduleDataSource{}
}

// NewScheduleDataSource defines the data source for a Scheduled Workflow.
type ScheduleDataSource struct {
	tclient temporalClient.Client
}

// ScheduleDataSourceModel describes the data source data model.
type ScheduleDataSourceModel struct {
	ScheduleId types.String `tfsdk:"id"`
	DescJson   types.String `tfsdk:"desc"`
}

func (d *ScheduleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (d *ScheduleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Scheduled Workflow data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Schedule ID",
				Required:            true,
			},
			"desc": schema.StringAttribute{
				MarkdownDescription: "Schedule description in JSON",
				Computed:            true,
			},
		},
	}
}

func (d *ScheduleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	tclient, ok := req.ProviderData.(temporalClient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected temporalClient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.tclient = tclient
}

func (d *ScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Read Terraform configuration data into the model
	var state ScheduleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the Schedule's description from the Server
	desc, err := d.tclient.ScheduleClient().GetHandle(ctx, state.ScheduleId.ValueString()).Describe(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("Read: Unable to describe Schedule %s : %s", state.ScheduleId.ValueString(), err))
		return
	}
	jsonBytes, err := json.Marshal(desc)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("Read: Unable to marshal Schedule description after Describe: %s", err))
		return
	}
	state.DescJson = basetypes.NewStringValue(string(jsonBytes))
	tflog.Trace(ctx, fmt.Sprintf("Read Schedule data source %s", state.ScheduleId.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
