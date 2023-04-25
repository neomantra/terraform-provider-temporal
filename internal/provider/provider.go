package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	temporalClient "go.temporal.io/sdk/client"

	"github.com/neomantra/terraform-provider-temporal/internal/zapadapter"
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
		Logger:    zapadapter.NewZapAdapter(buildProviderZapLogger()),
		Identity:  getProviderTemporalIdentity(),
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

////////////////////////////////////////////////////////////////////////

func getProviderTemporalIdentity() string {
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "unknown"
	}
	return fmt.Sprintf("terraform@%s", hostName)
}

func buildProviderZapLogger() *zap.Logger {
	encodeConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      zapcore.OmitKey, // we use our own caller
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   nil,
	}

	logLevel := zap.InfoLevel
	if os.Getenv("TF_DEBUG") != "" {
		logLevel = zap.DebugLevel
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(logLevel),
		Development:       false,
		DisableStacktrace: os.Getenv("TEMPORAL_CLI_SHOW_STACKS") == "",
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     encodeConfig,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableCaller:     true,
	}
	logger, _ := config.Build()
	return logger
}
