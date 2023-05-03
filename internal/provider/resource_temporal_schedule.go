package provider

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neomantra/terraform-provider-temporal/internal/tfschema"

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
	ScheduleId         types.String         `tfsdk:"id"`
	ScheduleSpec       *ScheduleSpecModel   `tfsdk:"schedule"`
	Action             *ScheduleActionModel `tfsdk:"action"`
	Overlap            types.String         `tfsdk:"overlap"`
	CatchupWindow      types.String         `tfsdk:"catchup_window"`
	PauseOnFailure     types.Bool           `tfsdk:"pause_on_failure"`
	Note               types.String         `tfsdk:"note"`
	Paused             types.Bool           `tfsdk:"paused"`
	RemainingActions   types.Int64          `tfsdk:"remaining_actions"`
	TriggerImmediately types.Bool           `tfsdk:"trigger_immediately"`
	//ScheduleBackfill []ScheduleBackfill `tfsdk:"schedule_backfill"`
	//Memo             types.String `tfsdk:"memo_json"`
	//SearchAttributes types.String `tfsdk:"search_attributes_json"`
}

type ScheduleSpecModel struct {
	// input.Description.Schedule.Spec = &temporalClient.ScheduleSpec{
	// Calendars       []ScheduleCalendarSpecModel `tfsdk:"calendars"`
	// Intervals       []ScheduleIntervalSpecModel `tfsdk:"intervals"`
	CronExpressions types.List `tfsdk:"cron"`
	// Skip            []ScheduleCalendarSpecModel `tfsdk:"skip"`
	// StartAt         types.String                `tfsdk:"start_at"` // time.Time
	// EndAt           types.String                `tfsdk:"end_at"`   // time.Time
	// Jitter          types.Int64                 `tfsdk:"jitter"`
	// TimeZoneName    types.String                `tfsdk:"time_zone"`
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

type ScheduleIntervalSpecModel struct {
	// Every - DURATION describes the period to repeat the interval.
	Every types.String `tfsdk:"every"`
	// Offset - DURATION is a fixed offset added to the intervals period. // Optional: Defaulted to 0
	Offset types.String `tfsdk:"offset"`
}

type ScheduleActionModel struct {
	// As other Scheduled Actions are invented by Temporal, add them here.
	StartWorkflow *ScheduleWorkflowActionResourceModel `tfsdk:"start_workflow"`
}

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

//////////////////////////////////////////////////////////////////////////////

func (r *ScheduleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (r *ScheduleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tfschema.MakeResourceScheduleSchema()
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

//////////////////////////////////////////////////////////////////////////////

func (r *ScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data *ScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	var action *temporalClient.ScheduleWorkflowAction
	// dataConverter := converter.GetDefaultDataConverter()

	if data.Action != nil && data.Action.StartWorkflow != nil {
		wfAction := data.Action.StartWorkflow
		action = &temporalClient.ScheduleWorkflowAction{
			ID:                       wfAction.WorkflowId.ValueString(),
			Workflow:                 wfAction.Workflow.ValueString(),
			TaskQueue:                wfAction.TaskQueue.ValueString(),
			WorkflowExecutionTimeout: toDuration(wfAction.WorkflowExecutionTimeout),
			WorkflowRunTimeout:       toDuration(wfAction.WorkflowRunTimeout),
			WorkflowTaskTimeout:      toDuration(wfAction.WorkflowTaskTimeout),
		}
		// action.Args action.Memo action.SearchAttributes
	}

	scheduleOptions := temporalClient.ScheduleOptions{
		ID: data.ScheduleId.ValueString(),
		// TODO: we must express all this in Terraform!
		// Spec: &temporalClient.ScheduleSpec{
		// 	Intervals: []temporalClient.ScheduleIntervalSpec{
		// 		{Every: 10},
		// 	},
		// },
		Action:             action,
		Overlap:            temporalEnums.SCHEDULE_OVERLAP_POLICY_SKIP, // TODO: convert from enum
		CatchupWindow:      toDuration(data.CatchupWindow),
		PauseOnFailure:     data.PauseOnFailure.ValueBool(),
		Note:               data.Note.ValueString(),
		Paused:             data.Paused.ValueBool(),
		RemainingActions:   int(data.RemainingActions.ValueInt64()),
		TriggerImmediately: data.TriggerImmediately.ValueBool(),
		// ScheduleBackfill
	}
	// Args Memo SearchAttributes

	// Invoke Create on the Temporal API
	scheduleHandle, err := r.tclient.ScheduleClient().Create(ctx, scheduleOptions)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("Create: Unable to create Schedule: %s", err))
		return
	}
	data.ScheduleId = types.StringValue(scheduleHandle.GetID())

	// Fetch the new Schedule's description from the Server
	desc, err := scheduleHandle.Describe(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("Read: Unable to describe created Schedule %s : %s", scheduleHandle.GetID(), err))
		return
	}
	switch newAction := desc.Schedule.Action.(type) {
	case *temporalClient.ScheduleWorkflowAction:
		if data.Action == nil {
			data.Action = &ScheduleActionModel{}
		}
		if data.Action.StartWorkflow == nil {
			data.Action.StartWorkflow = &ScheduleWorkflowActionResourceModel{}
		}
		data.Action.StartWorkflow.WorkflowId = types.StringValue(newAction.ID)
		data.Action.StartWorkflow.Workflow = getWorkflowName(newAction.Workflow)
		data.Action.StartWorkflow.TaskQueue = types.StringValue(newAction.TaskQueue)
		data.Action.StartWorkflow.WorkflowExecutionTimeout = fromDuration(newAction.WorkflowExecutionTimeout)
		data.Action.StartWorkflow.WorkflowRunTimeout = fromDuration(newAction.WorkflowRunTimeout)
		data.Action.StartWorkflow.WorkflowTaskTimeout = fromDuration(newAction.WorkflowTaskTimeout)
		// Args Memo SearchAttributes
	}

	// Save data into Terraform state
	tflog.Info(ctx, fmt.Sprintf("Created Schedule resource %s", data.ScheduleId.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//////////////////////////////////////////////////////////////////////////////

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

	// Convert from the Temporal Schedule model to the Terraform data model
	if policy := desc.Schedule.Policy; policy != nil {
		data.Overlap = types.StringValue(policy.Overlap.String())
		data.CatchupWindow = fromDuration(policy.CatchupWindow)
		data.PauseOnFailure = types.BoolValue(policy.PauseOnFailure)
	}
	if state := desc.Schedule.State; state != nil {
		data.Note = fromString(state.Note)
		data.Paused = types.BoolValue(state.Paused)
		data.RemainingActions = types.Int64Value(int64(state.RemainingActions))
		if data.TriggerImmediately.IsUnknown() || data.TriggerImmediately.IsNull() {
			// TriggerImmediately isn't stored in Temporal state, so we clear it out here
			data.TriggerImmediately = types.BoolValue(false)
		}
		// LimitedActions ?
	}
	// Memo SearchAttributes

	switch action := desc.Schedule.Action.(type) {
	case *temporalClient.ScheduleWorkflowAction:
		if data.Action == nil {
			data.Action = &ScheduleActionModel{}
		}
		if data.Action.StartWorkflow == nil {
			data.Action.StartWorkflow = &ScheduleWorkflowActionResourceModel{}
		}
		data.Action.StartWorkflow.WorkflowId = types.StringValue(action.ID)
		data.Action.StartWorkflow.Workflow = getWorkflowName(action.Workflow)
		data.Action.StartWorkflow.TaskQueue = types.StringValue(action.TaskQueue)
		data.Action.StartWorkflow.WorkflowExecutionTimeout = fromDuration(action.WorkflowExecutionTimeout)
		data.Action.StartWorkflow.WorkflowRunTimeout = fromDuration(action.WorkflowRunTimeout)
		data.Action.StartWorkflow.WorkflowTaskTimeout = fromDuration(action.WorkflowTaskTimeout)
	}

	// Save updated data into Terraform state
	tflog.Info(ctx, fmt.Sprintf("Read ScheduledWorkflow resource %s", data.ScheduleId.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//////////////////////////////////////////////////////////////////////////////

func (r *ScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var data *ScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Invoke Create on the Temporal API
	err := r.tclient.ScheduleClient().GetHandle(ctx, data.ScheduleId.ValueString()).Update(ctx, temporalClient.ScheduleUpdateOptions{
		DoUpdate: func(input temporalClient.ScheduleUpdateInput) (*temporalClient.ScheduleUpdate, error) {
			// update action
			var wfAction *temporalClient.ScheduleWorkflowAction
			switch actionType := input.Description.Schedule.Action.(type) {
			case *temporalClient.ScheduleWorkflowAction:
				wfAction = actionType
			}
			if data.Action != nil && data.Action.StartWorkflow != nil {
				if wfAction == nil {
					wfAction = &temporalClient.ScheduleWorkflowAction{}
					input.Description.Schedule.Action = wfAction
				}
				workflow := data.Action.StartWorkflow
				wfAction.ID = workflow.WorkflowId.ValueString()
				wfAction.Workflow = workflow.Workflow.ValueString()
				wfAction.TaskQueue = workflow.TaskQueue.ValueString()
				wfAction.WorkflowExecutionTimeout = toDuration(workflow.WorkflowExecutionTimeout)
				wfAction.WorkflowRunTimeout = toDuration(workflow.WorkflowRunTimeout)
				wfAction.WorkflowTaskTimeout = toDuration(workflow.WorkflowTaskTimeout)
			}

			policy := input.Description.Schedule.Policy
			policy.Overlap = temporalEnums.SCHEDULE_OVERLAP_POLICY_SKIP // TODO: convert from enum
			policy.CatchupWindow = toDuration(data.CatchupWindow)
			policy.PauseOnFailure = data.PauseOnFailure.ValueBool()

			input.Description.Schedule.Spec = &temporalClient.ScheduleSpec{
				// Calendars:       []internal.ScheduleCalendarSpec{},
				// Intervals:       []internal.ScheduleIntervalSpec{},
				// CronExpressions: []string{},
				// Skip:            []internal.ScheduleCalendarSpec{},
				// StartAt:         time.Time{},
				// EndAt:           time.Time{},
				// Jitter:          0,
				// TimeZoneName:    "",
			}
			state := input.Description.Schedule.State
			state.Note = data.Note.ValueString()
			state.Paused = data.Paused.ValueBool()
			state.RemainingActions = int(data.RemainingActions.ValueInt64())
			//LimitedActions:   false,

			return &temporalClient.ScheduleUpdate{
				Schedule: &input.Description.Schedule,
			}, nil
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("Update: Unable to Update Schedule %s : %s", data.ScheduleId.ValueString(), err))
		return
	}

	// Save updated data into Terraform state
	tflog.Info(ctx, fmt.Sprintf("Update ScheduledWorkflow resource %s", data.ScheduleId.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

//////////////////////////////////////////////////////////////////////////////

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
	tflog.Debug(ctx, fmt.Sprintf("Deleted ScheduledWorkflow resource %s", data.ScheduleId.ValueString()))
}

//////////////////////////////////////////////////////////////////////////////

func (r *ScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

//////////////////////////////////////////////////////////////////////////////

func toDuration(val types.String) time.Duration {
	duration, _ := time.ParseDuration(val.ValueString())
	return duration
}

func fromDuration(duration time.Duration) types.String {
	if duration == 0 {
		return types.StringNull()
	}
	return types.StringValue(duration.String())
}

func fromString(str string) types.String {
	if str == "" {
		return types.StringNull()
	}
	return types.StringValue(str)
}

func getKind(fType reflect.Type) reflect.Kind {
	if fType == nil {
		return reflect.Invalid
	}
	return fType.Kind()
}

// getWorkflowName returns the name of a workflow from a generic Workflow field.  Only supports string names.
func getWorkflowName(workflow interface{}) types.String {
	fType := reflect.TypeOf(workflow)
	switch getKind(fType) {
	case reflect.String:
		return types.StringValue(reflect.ValueOf(workflow).String())
	default:
		return types.StringUnknown()
	}
}
