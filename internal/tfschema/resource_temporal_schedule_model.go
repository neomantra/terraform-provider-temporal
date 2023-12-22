package tfschema

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ScheduleResourceModel describes the resource data model.
type ScheduleResourceModel struct {
	ScheduleId         types.String                         `tfsdk:"id"`
	ScheduleSpec       *ScheduleSpecModel                   `tfsdk:"schedule"`
	StartWorkflow      *ScheduleWorkflowActionResourceModel `tfsdk:"start_workflow"`
	Overlap            types.String                         `tfsdk:"overlap"`
	CatchupWindow      types.String                         `tfsdk:"catchup_window"`
	PauseOnFailure     types.Bool                           `tfsdk:"pause_on_failure"`
	Note               types.String                         `tfsdk:"note"`
	Paused             types.Bool                           `tfsdk:"paused"`
	RemainingActions   types.Int64                          `tfsdk:"remaining_actions"`
	TriggerImmediately types.Bool                           `tfsdk:"trigger_immediately"`
	//ScheduleBackfill []ScheduleBackfill `tfsdk:"schedule_backfill"`
	//Memo             types.String `tfsdk:"memo_json"`
	//SearchAttributes types.String `tfsdk:"search_attributes_json"`
}

type ScheduleSpecModel struct {
	//Intervals []ScheduleIntervalSpecModel `tfsdk:"intervals"`
	// Skip            []ScheduleCalendarSpecModel `tfsdk:"skip"`
	// Crons     types.List                  `tfsdk:"crons"`
	// Calendars []ScheduleCalendarSpecModel `tfsdk:"calendars"`
	// StartAt      timetypes.RFC3339 `tfsdk:"start_at"`
	// EndAt        timetypes.RFC3339 `tfsdk:"end_at"`
	Jitter       types.String `tfsdk:"jitter"`
	TimeZoneName types.String `tfsdk:"time_zone"`
}

type ScheduleCalendarSpecModel struct {
	// Second range to match (0-59). default: matches 0
	Second []ScheduleRangeModel `tfsdk:"second"`
	// Minute range to match (0-59). default: matches 0
	Minute []ScheduleRangeModel `tfsdk:"minute"`
	// Hour range to match (0-23). default: matches 0
	Hour []ScheduleRangeModel `tfsdk:"hour"`
	// DayOfMonth range to match (1-31).  default: matches all days
	DayOfMonth []ScheduleRangeModel `tfsdk:"day_of_month"`
	// Month range to match (1-12).  default: matches all months
	Month []ScheduleRangeModel `tfsdk:"month"`
	// Year range to match. default: empty that matches all years
	Year []ScheduleRangeModel `tfsdk:"year"`
	// DayOfWeek range to match (0-6; 0 is Sunday). default: matches all days of the week
	DayOfWeek []ScheduleRangeModel `tfsdk:"day_of_week"`
	// Comment - Description of the intention of this schedule.
	Comment types.String `tfsdk:"comment"`
}

type ScheduleRangeModel struct {
	// Start of the range (inclusive)
	Start types.Int64 `tfsdk:"start"`
	// End of the range (inclusive)
	// Optional: defaulted to Start
	End types.Int64 `tfsdk:"end"`
	// Step to be take between each value
	// Optional: defaulted to 1
	Step types.Int64 `tfsdk:"step"`
}

type ScheduleIntervalSpecModel struct {
	// Every - DURATION describes the period to repeat the interval.
	Every types.String `tfsdk:"every"`
	// Offset - DURATION is a fixed offset added to the intervals period. // Optional: Defaulted to 0
	Offset types.String `tfsdk:"offset"`
}

// TODO: can remove?
// type ScheduleActionModel struct {
// 	// As other Scheduled Actions are invented by Temporal, add them here.
// 	StartWorkflow ScheduleWorkflowActionResourceModel `tfsdk:"start_workflow"`
// }

// ScheduleWorkflowAction implements ScheduleAction to launch a workflow.
type ScheduleWorkflowActionResourceModel struct {
	WorkflowId types.String `tfsdk:"workflow_id"`
	Workflow   types.String `tfsdk:"workflow"`
	//Args                     types.List   `tfsdk:"args"`
	TaskQueue                types.String `tfsdk:"task_queue"`
	WorkflowExecutionTimeout types.String `tfsdk:"execution_timeout"`
	WorkflowRunTimeout       types.String `tfsdk:"run_timeout"`
	WorkflowTaskTimeout      types.String `tfsdk:"task_timeout"`
	// RetryPolicy - Retry policy for workflow. If a retry policy is specified, in case of workflow failure
	// server will start new workflow execution if needed based on the retry policy.
	//RetryPolicy *RetryPolicy
	//Memo             types.String `tfsdk:"memo_json"`
	//SearchAttributes types.String `tfsdk:"search_attributes_json"`
}
