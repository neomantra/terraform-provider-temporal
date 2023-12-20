package tfschema

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/neomantra/terraform-provider-temporal/internal/docs"
)

const temporalNamespaceMinRetentionHours = 24 // Temporal minimum namespace retention is 24 hours
const defaultRetentionHours = 24 * 3          // 3 days is the default

func GetResourceNamespaceDefaultValue() basetypes.ObjectValue {
	return types.ObjectValueMust(
		map[string]attr.Type{
			"retention_hours": types.Int64Type,
		},
		map[string]attr.Value{
			"retention_hours": types.Int64Value(defaultRetentionHours),
		},
	)
}

func GetResourceNamespaceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Namespace resource",
		Attributes: map[string]schema.Attribute{
			// Terraform-internal value, assigned to <name>
			// Seems to still be required by TF Framework testing
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: docs.NamespaceNameDocs,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"retention_hours": schema.Int64Attribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: docs.NamespaceRetentionHoursDocs,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(temporalNamespaceMinRetentionHours),
				},
				Default: int64default.StaticInt64(defaultRetentionHours),
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: docs.NamespaceDescriptionDocs,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner_email": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: docs.NamespaceOwnerEmailDocs,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Clusters                         []*v1.ClusterReplicationConfig `protobuf:"bytes,5,rep,name=clusters,proto3" json:"clusters,omitempty"`
// ActiveClusterName                string                         `protobuf:"bytes,6,opt,name=active_cluster_name,json=activeClusterName,proto3" json:"active_cluster_name,omitempty"`
// A key-value map for any customized purpose.
// Data              map[string]string `protobuf:"bytes,7,rep,name=data,proto3" json:"data,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
// SecurityToken     string            `protobuf:"bytes,8,opt,name=security_token,json=securityToken,proto3" json:"security_token,omitempty"`
// IsGlobalNamespace bool              `protobuf:"varint,9,opt,name=is_global_namespace,json=isGlobalNamespace,proto3" json:"is_global_namespace,omitempty"`

// // If unspecified (ARCHIVAL_STATE_UNSPECIFIED) then default server configuration is used.
// HistoryArchivalState v11.ArchivalState `protobuf:"varint,10,opt,name=history_archival_state,json=historyArchivalState,proto3,enum=temporal.api.enums.v1.ArchivalState" json:"history_archival_state,omitempty"`
// HistoryArchivalUri   string            `protobuf:"bytes,11,opt,name=history_archival_uri,json=historyArchivalUri,proto3" json:"history_archival_uri,omitempty"`
// // If unspecified (ARCHIVAL_STATE_UNSPECIFIED) then default server configuration is used.
// VisibilityArchivalState v11.ArchivalState `protobuf:"varint,12,opt,name=visibility_archival_state,json=visibilityArchivalState,proto3,enum=temporal.api.enums.v1.ArchivalState" json:"visibility_archival_state,omitempty"`
// VisibilityArchivalUri   string            `protobuf:"bytes,13,opt,name=visibility_archival_uri,json=visibilityArchivalUri,proto3" json:"visibility_archival_uri,omitempty"`
