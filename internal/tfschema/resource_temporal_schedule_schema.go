package tfschema

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	temporalEnums "go.temporal.io/api/enums/v1"

	"github.com/neomantra/terraform-provider-temporal/internal/docs"
)

//////////////////////////////////////////////////////////////////////////////

func GetResourceScheduleDefaultValue() basetypes.ObjectValue {
	return types.ObjectValueMust(
		map[string]attr.Type{
			"catchup_window":      types.StringType,
			"pause_on_failure":    types.BoolType,
			"remaining_actions":   types.Int64Type,
			"trigger_immediately": types.BoolType,
			"overlap":             types.StringType,
		},
		map[string]attr.Value{
			"catchup_window":      types.StringValue("1m0s"),
			"pause_on_failure":    types.BoolValue(false),
			"remaining_actions":   types.Int64Value(0),
			"trigger_immediately": types.BoolValue(false),
			"overlap":             types.StringValue(temporalEnums.SCHEDULE_OVERLAP_POLICY_SKIP.String()),
		},
	)
}

func GetResourceScheduleSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Schedule resource",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: docs.ScheduleIDDocs,
			},
			"schedule":       GetScheduleSingleNestedAttribute(),
			"start_workflow": GetStartWorkflowActionSingleNestedAttribute(),
			"overlap": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(temporalEnums.SCHEDULE_OVERLAP_POLICY_SKIP.String()),
				MarkdownDescription: docs.ScheduleOverlapDocs,
			},
			"catchup_window": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("1m0s"),
				MarkdownDescription: docs.ScheduleCatchupWindowDocs,
			},
			"pause_on_failure": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: docs.SchedulePauseOnFailureDocs,
			},
			"note": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: docs.ScheduleNoteDocs,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"paused": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: docs.SchedulePausedDocs,
			},
			"remaining_actions": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: docs.ScheduleRemainingActionsDocs,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"trigger_immediately": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: docs.ScheduleTriggerImmediatelyDocs,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			// TODO
			// "schedule_backfill": schema.BoolAttribute{
			// 	Optional:            true,
			// 	MarkdownDescription: docs.scheduleScheduleBackfillDocs,
			// 	// PlanModifiers: []planmodifier.String{
			// 	// 	stringplanmodifier.UseStateForUnknown(),
			// 	// },
			// },
			// "memo_json": schema.StringAttribute{
			// 	Optional:            true,
			// 	MarkdownDescription: docs.scheduleMemoDocs,
			// 	// PlanModifiers: []planmodifier.String{
			// 	// 	stringplanmodifier.UseStateForUnknown(),
			// 	// },
			// },
			// "search_attributes_json": schema.StringAttribute{
			// 	Optional:            true,
			// 	MarkdownDescription: docs.scheduleSearchAttributesDocs,
			// 	// PlanModifiers: []planmodifier.String{
			// 	// 	stringplanmodifier.UseStateForUnknown(),
			// 	// },
			// },
		},
	}
}

//////////////////////////////////////////////////////////////////////////////

func GetScheduleDefaultValue() basetypes.ObjectValue {
	return types.ObjectValueMust(
		map[string]attr.Type{
			//"crons": types.ListType{ElemType: types.StringType},
			"jitter":    types.StringType,
			"time_zone": types.StringType,
			// "start_at":  types.StringType,
			// "end_at":    types.StringType,
		},
		map[string]attr.Value{
			//"crons": types.ListValueMust(types.StringType, []attr.Value{}),
			"jitter":    types.StringValue("0s"),
			"time_zone": types.StringValue(""),
			// "start_at":  types.StringValue("0001-01-01T00:00:00Z"),
			// "end_at":    types.StringValue("0001-01-01T00:00:00Z"),
		},
	)
}

func GetScheduleSingleNestedAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed:            true,
		Optional:            true,
		MarkdownDescription: docs.ScheduleScheduleDocs,
		Default:             objectdefault.StaticValue(GetScheduleDefaultValue()),
		Attributes: map[string]schema.Attribute{
			// "calendars": schema.ListNestedAttribute{
			// 	// read-only, do it via "crons" for now
			// 	Computed:            true,
			// 	MarkdownDescription: docs.ScheduleCalendarSpecDocs,
			// 	NestedObject:        getScheduleCalendarNestedAttributeObject(),
			// 	Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			// },
			// "crons": schema.ListAttribute{
			// 	Optional:            true,
			// 	Computed:            true,
			// 	MarkdownDescription: docs.ScheduleSpecCronArgDocs,
			// 	Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			// 	ElementType:         types.StringType,
			// },
			// "start_at": schema.StringAttribute{
			// 	CustomType:          timetypes.RFC3339Type{},
			// 	Optional:            true,
			// 	Computed:            true,
			// 	Default:             stringdefault.StaticString("0001-01-01T00:00:00Z"),
			// 	MarkdownDescription: docs.ScheduleSpecStartAtDocs,
			// },
			// "end_at": schema.StringAttribute{
			// 	CustomType:          timetypes.RFC3339Type{},
			// 	Optional:            true,
			// 	Computed:            true,
			// 	Default:             stringdefault.StaticString("0001-01-01T00:00:00Z"),
			// 	MarkdownDescription: docs.ScheduleSpecEndAtDocs,
			// },
			"jitter": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("0s"),
				MarkdownDescription: docs.ScheduleSpecJitterDocs,
			},
			"time_zone": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: docs.ScheduleSpecTimeZoneDocs,
			},
		},
	}
}

///////////////////////////////////////////////////////////////////////////////

func GetDefaultStartWorkflowAction() basetypes.ObjectValue {
	return types.ObjectValueMust(
		map[string]attr.Type{
			"task_timeout": types.StringType,
		},
		map[string]attr.Value{
			"task_timeout": types.StringValue("10s"),
		},
	)
}

func GetStartWorkflowActionSingleNestedAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed:            true,
		Optional:            true,
		MarkdownDescription: docs.ScheduleActionDocs,
		Default:             objectdefault.StaticValue(GetDefaultStartWorkflowAction()),
		Attributes: map[string]schema.Attribute{
			"task_queue": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: docs.ScheduleWATaskQueueDocs,
			},
			"workflow_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: docs.ScheduleWAIDDocs,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workflow": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: docs.ScheduleWAWorkflowDocs,
			},
			// "args": schema.ListAttribute{
			// 	Optional:            true,
			// 	ElementType:         types.StringType,
			// 	MarkdownDescription: scheduleWAArgDocs,
			// 	PlanModifiers: []planmodifier.List{
			// 		listplanmodifier.UseStateForUnknown(),
			// 	},
			// },
			"execution_timeout": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: docs.ScheduleWAWorkflowExecutionTimeoutDocs,
			},
			"run_timeout": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"task_timeout": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: docs.ScheduleWAWorkflowTaskTimeoutDocs,
				Default:             stringdefault.StaticString("10s"),
			},
			// // RetryPolicy - Retry policy for workflow. If a retry policy is specified, in case of workflow failure
			// // server will start new workflow execution if needed based on the retry policy.
			// //RetryPolicy *RetryPolicy
			// scheduleWARetryPolicyDocs
			// "memo_json": schema.StringAttribute{
			// 	Optional:            true,
			// 	MarkdownDescription: docs.scheduleWAMemoDocs,
			// 	// PlanModifiers: []planmodifier.String{
			// 	// 	stringplanmodifier.UseStateForUnknown(),
			// 	// },
			// },
			// "search_attributes_json": schema.StringAttribute{
			// 	Optional:            true,
			// 	MarkdownDescription: docs.scheduleWASearchAttributesDocs,
			// 	// PlanModifiers: []planmodifier.String{
			// 	// 	stringplanmodifier.UseStateForUnknown(),
			// 	// },
			// },
		},
	}
}

func getScheduleCalendarNestedAttributeObject() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"second":      getScheduleRangeListNestedAttribute(false, true, true, docs.ScheduleCalendarSecondDocs),
			"minute":      getScheduleRangeListNestedAttribute(false, true, true, docs.ScheduleCalendarMinuteDocs),
			"hour":        getScheduleRangeListNestedAttribute(false, true, true, docs.ScheduleCalendarHourDocs),
			"day":         getScheduleRangeListNestedAttribute(false, true, true, docs.ScheduleCalendarDayOfMonthDocs),
			"month":       getScheduleRangeListNestedAttribute(false, true, true, docs.ScheduleCalendarMonthDocs),
			"year":        getScheduleRangeListNestedAttribute(false, true, true, docs.ScheduleCalendarYearDocs),
			"day_of_week": getScheduleRangeListNestedAttribute(false, true, true, docs.ScheduleCalendarDayOfWeekDocs),
			"comment": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: docs.ScheduleCalendarCommentDocs,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func getScheduleRangeListNestedAttribute(required, computed, optional bool, desc string) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Required:            required,
		Computed:            computed,
		Optional:            optional,
		Description:         desc,
		MarkdownDescription: desc,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"start": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: docs.ScheduleRangeStartDocs,
				},
				"end": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: docs.ScheduleRangeEndDocs,
				},
				"step": schema.Int64Attribute{
					Computed:            true,
					MarkdownDescription: docs.ScheduleRangeStepDocs,
				},
			},
		},
	}
}
