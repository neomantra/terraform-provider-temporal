package tfschema

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	temporalEnums "go.temporal.io/api/enums/v1"
	//"github.com/neomantra/terraform-provider-temporal/internal/tfschema"
	"github.com/neomantra/terraform-provider-temporal/internal/docs"
)

func MakeResourceScheduleSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Schedule resource",
		Blocks: map[string]schema.Block{
			"schedule": makeScheduleBlock(),
			"action":   makeActionBlock(),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: docs.ScheduleIDDocs,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"overlap": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(temporalEnums.SCHEDULE_OVERLAP_POLICY_SKIP.String()),
				MarkdownDescription: docs.ScheduleOverlapDocs,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"catchup_window": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("1m0s"),
				MarkdownDescription: docs.ScheduleCatchupWindowDocs,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pause_on_failure": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: docs.SchedulePauseOnFailureDocs,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"remaining_actions": schema.Int64Attribute{
				Computed:            true,
				Optional:            true,
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
				// PlanModifiers: []planmodifier.Bool{
				// boolplanmodifier.UseStateForUnknown(),
				// },
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

			// // Action - Which Action to take.
			// Action ScheduleAction

			// Overlap - Controls what happens when an Action would be started by a Schedule at the same time that an older Action is still
			// running. This can be changed after a Schedule has taken some Actions, and some changes might produce
			// unintuitive results. In general, the later policy overrides the earlier policy.
			//
			// Optional: defaulted to SCHEDULE_OVERLAP_POLICY_SKIP
			// TODO Overlap enumspb.ScheduleOverlapPolicy

			// // ScheduleBackfill - Runs though the specified time periods and takes Actions as if that time passed by right now, all at once. The
			// // overlap policy can be overridden for the scope of the ScheduleBackfill.
			// TODO ScheduleBackfill []ScheduleBackfill
		},
	}
}

func makeScheduleBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: docs.ScheduleScheduleDocs,
		Blocks:              map[string]schema.Block{
			// "calendars": schema.SingleNestedBlock{},
			// "intervals": schema.SingleNestedBlock{},
			// "skip":      schema.SingleNestedBlock{},
			// "start_at":  schema.SingleNestedBlock{},
			// "end_at":    schema.SingleNestedBlock{},
			// "jitter":    schema.SingleNestedBlock{},
			// "time_zone": schema.SingleNestedBlock{},
		},
		Attributes: map[string]schema.Attribute{
			"cron": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: docs.ScheduleSpecCronArgDocs,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func makeActionBlock() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		MarkdownDescription: docs.ScheduleActionDocs,
		Blocks: map[string]schema.Block{
			"start_workflow": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
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
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					// "args": schema.ListAttribute{
					// 	Optional:            true,
					// 	ElementType:         types.StringType,
					// 	MarkdownDescription: scheduleWAArgDocs,
					// 	PlanModifiers: []planmodifier.List{
					// 		listplanmodifier.UseStateForUnknown(),
					// 	},
					// },
					"task_queue": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: docs.ScheduleWATaskQueueDocs,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"execution_timeout": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: docs.ScheduleWAWorkflowExecutionTimeoutDocs,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"run_timeout": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: docs.ScheduleWAWorkflowRunTimeoutDocs,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"task_timeout": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: docs.ScheduleWAWorkflowTaskTimeoutDocs,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
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
			},
		},
	}
}
