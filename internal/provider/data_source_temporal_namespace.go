package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/neomantra/terraform-provider-temporal/internal/docs"

	temporalClient "go.temporal.io/sdk/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &NamespaceDataSource{}

func NewNamespaceDataSource() datasource.DataSource {
	return &NamespaceDataSource{}
}

// NewNamespaceDataSource defines the data source for a Namespace.
type NamespaceDataSource struct {
	nsclient temporalClient.NamespaceClient
}

// NamespaceDataSourceModel describes the data source data model.
type NamespaceDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	RetentionHours types.Int64  `tfsdk:"retention_hours"`
	Description    types.String `tfsdk:"description"`
	OwnerEmail     types.String `tfsdk:"owner_email"`
}

func (d *NamespaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (d *NamespaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Temporal Namespace data source",
		Attributes: map[string]schema.Attribute{
			// Terraform-internal value, assigned to <name>
			// Seems to still be required by TF Framework testing
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: docs.NamespaceNameDocs,
			},
			"retention_hours": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: docs.NamespaceRetentionHoursDocs,
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: docs.NamespaceDescriptionDocs,
			},
			"owner_email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: docs.NamespaceOwnerEmailDocs,
			},
		},
	}
}

func (d *NamespaceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	tdata, ok := req.ProviderData.(TemporalProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("NamespaceDataSource Create: expected TemporalProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.nsclient = tdata.nsclient
}

func (d *NamespaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Read Terraform configuration data into the model
	var state NamespaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the Namespace's description from the Server
	desc, err := d.nsclient.Describe(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("NamespaceDataSource Read: unable to describe Namespace %s : %s", state.Name.ValueString(), err))
		return
	}

	state.Name = types.StringValue(desc.NamespaceInfo.Name)
	state.Id = state.Name
	state.Description = types.StringValue(desc.NamespaceInfo.Description)
	state.OwnerEmail = types.StringValue(desc.NamespaceInfo.OwnerEmail)
	state.RetentionHours = types.Int64Value(int64(desc.Config.WorkflowExecutionRetentionTtl.Hours()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
