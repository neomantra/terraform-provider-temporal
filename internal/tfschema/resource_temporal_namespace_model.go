package tfschema

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NamespaceResourceModel describes the resource data model.
type NamespaceResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	RetentionHours types.Int64  `tfsdk:"retention_hours"`
	Description    types.String `tfsdk:"description"`
	OwnerEmail     types.String `tfsdk:"owner_email"`
}
