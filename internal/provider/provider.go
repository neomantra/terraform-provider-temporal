package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	temporalClient "go.temporal.io/sdk/client"
)

// Ensure TemporalProvider satisfies various provider interfaces.
var _ provider.Provider = &TemporalProvider{}

// TemporalProvider defines the provider implementation.
type TemporalProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TemporalProviderModel describes the provider data model.
type TemporalProviderModel struct {
	HostPort  types.String `tfsdk:"hostport"`
	Namespace types.String `tfsdk:"namespace"`
}

func (p *TemporalProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "temporal"
	resp.Version = p.version
}

func (p *TemporalProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		// TODO: environment varibales TEMPORAL_CLI_ADDRESS and TEMPORAL_NAMESPACE
		Attributes: map[string]schema.Attribute{
			"hostport": schema.StringAttribute{
				MarkdownDescription: "`host:port` of the Temporal Server. Overrides TEMPORAL_CLI_ADDRESS",
				Optional:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Temporal namespace. Overrides TEMPORAL_NAMESPACE or 'default'",
				Optional:            true,
			},
		},
	}
}

func (p *TemporalProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var providerConfig TemporalProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &providerConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for hostport/namespace from environment
	hostPort := providerConfig.HostPort.ValueString()
	if hostPort == "" {
		hostPort = os.Getenv("TEMPORAL_CLI_ADDRESS")
		if hostPort == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("hostport"),
				"Temporal attribute 'hostport' or environment 'TEMPORAL_CLI_ADDRESS' must be set",
				"Temporal attribute 'hostport' or environment 'TEMPORAL_CLI_ADDRESS' must be set",
			)
		}
	}
	namespace := providerConfig.Namespace.ValueString()
	if namespace == "" {
		namespace = os.Getenv("TEMPORAL_NAMESPACE")
		if namespace == "" {
			namespace = "default"
		}
	}

	// Example client configuration for data sources and resources
	tclient, _ := temporalClient.NewLazyClient(temporalClient.Options{
		HostPort:  hostPort,
		Namespace: namespace,
	})

	resp.DataSourceData = tclient
	resp.ResourceData = tclient
}

func (p *TemporalProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewScheduleResource,
	}
}

func (p *TemporalProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewScheduleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TemporalProvider{
			version: version,
		}
	}
}
