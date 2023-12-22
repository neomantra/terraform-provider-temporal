package provider

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neomantra/terraform-provider-temporal/internal/tfschema"

	temporalEnums "go.temporal.io/api/enums/v1"
	temporalClient "go.temporal.io/sdk/client"
	temporalSdk "go.temporal.io/sdk/temporal"
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

//////////////////////////////////////////////////////////////////////////////

func (r *ScheduleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (r *ScheduleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tfschema.GetResourceScheduleSchema()
}

func (r *ScheduleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	tdata, ok := req.ProviderData.(TemporalProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("ScheduleResource expected temporalClient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.tclient = tdata.tclient
}

//////////////////////////////////////////////////////////////////////////////

func (r *ScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform PLAN data into the models
	var plan tfschema.ScheduleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Marshal Terraform type to Temporal Schedule
	scheduleOptions := temporalClient.ScheduleOptions{
		ID:                 plan.ScheduleId.ValueString(),
		Overlap:            temporalEnums.SCHEDULE_OVERLAP_POLICY_SKIP, // TODO: convert from enum
		CatchupWindow:      toDuration(plan.CatchupWindow),
		PauseOnFailure:     plan.PauseOnFailure.ValueBool(),
		Note:               plan.Note.ValueString(),
		Paused:             plan.Paused.ValueBool(),
		RemainingActions:   int(plan.RemainingActions.ValueInt64()),
		TriggerImmediately: plan.TriggerImmediately.ValueBool(),
		// ScheduleBackfill
	}

	if plan.ScheduleSpec != nil {
		// scheduleOptions.Spec.StartAt, diags = plan.ScheduleSpec.StartAt.ValueRFC3339Time()
		// resp.Diagnostics.Append(diags...)
		// scheduleOptions.Spec.EndAt, diags = plan.ScheduleSpec.EndAt.ValueRFC3339Time()
		// resp.Diagnostics.Append(diags...)
		//scheduleOptions.Spec.CronExpressions = toStringArray(ctx, plan.ScheduleSpec.Crons)
		scheduleOptions.Spec.Jitter = toDuration(plan.ScheduleSpec.Jitter)
		scheduleOptions.Spec.TimeZoneName = plan.ScheduleSpec.TimeZoneName.ValueString()
	}

	if plan.StartWorkflow != nil {
		scheduleOptions.Action = &temporalClient.ScheduleWorkflowAction{
			ID:                       plan.StartWorkflow.WorkflowId.ValueString(),
			Workflow:                 plan.StartWorkflow.Workflow.ValueString(),
			TaskQueue:                plan.StartWorkflow.TaskQueue.ValueString(),
			WorkflowExecutionTimeout: toDuration(plan.StartWorkflow.WorkflowExecutionTimeout),
			WorkflowRunTimeout:       toDuration(plan.StartWorkflow.WorkflowRunTimeout),
			WorkflowTaskTimeout:      toDuration(plan.StartWorkflow.WorkflowTaskTimeout),
		}
	}
	// Args Memo SearchAttributes

	// Invoke Create on the Temporal API
	scheduleHandle, err := r.tclient.ScheduleClient().Create(ctx, scheduleOptions)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("ScheduleResource Create: Unable to create Schedule: %s", err))
		return
	}
	createdScheduleID := scheduleHandle.GetID()

	// Fetch the new Schedule's description from the Server
	desc, err := scheduleHandle.Describe(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("ScheduleResource Create: Unable to describe created Schedule %s : %s", createdScheduleID, err))
		return
	}

	// Construct final STATE, including Temporal-created values
	state := tfschema.ScheduleResourceModel{}
	assignScheduleResourceModelFromDescription(createdScheduleID, desc.Schedule, &state)

	// Save data into Terraform state
	tflog.Info(ctx, fmt.Sprintf("Created Schedule resource ID: %s createID: %s", state.ScheduleId.ValueString(), createdScheduleID))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

//////////////////////////////////////////////////////////////////////////////

func (r *ScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var plan *tfschema.ScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the Schedule's description from the Server
	scheduleID := plan.ScheduleId.ValueString()
	desc, err := r.tclient.ScheduleClient().GetHandle(ctx, plan.ScheduleId.ValueString()).Describe(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("ScheduleResource Read: Unable to describe Schedule %s : %s", scheduleID, err))
		return
	}

	// Construct final STATE, including Temporal-created values
	state := tfschema.ScheduleResourceModel{}
	assignScheduleResourceModelFromDescription(scheduleID, desc.Schedule, &state)

	// Save updated data into Terraform state
	tflog.Info(ctx, fmt.Sprintf("Read ScheduledWorkflow resource %s", state.ScheduleId.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

//////////////////////////////////////////////////////////////////////////////

func (r *ScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform PLAN data into the models
	var plan tfschema.ScheduleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Invoke Update on the Temporal API
	scheduleID := plan.ScheduleId.ValueString()
	err := r.tclient.ScheduleClient().GetHandle(ctx, scheduleID).Update(ctx, temporalClient.ScheduleUpdateOptions{
		DoUpdate: func(input temporalClient.ScheduleUpdateInput) (*temporalClient.ScheduleUpdate, error) {
			// update action
			scheduleUpdate := &temporalClient.ScheduleUpdate{
				Schedule: &temporalClient.Schedule{
					Policy: &temporalClient.SchedulePolicies{
						Overlap:        temporalEnums.SCHEDULE_OVERLAP_POLICY_SKIP, // TODO: convert from enum
						CatchupWindow:  toDuration(plan.CatchupWindow),
						PauseOnFailure: plan.PauseOnFailure.ValueBool(),
					},
					State: &temporalClient.ScheduleState{
						Note:             plan.Note.ValueString(),
						RemainingActions: int(plan.RemainingActions.ValueInt64()),
					},
					Spec: &temporalClient.ScheduleSpec{},
				},
			}
			if plan.StartWorkflow == nil {
				resp.Diagnostics.AddError("Temporal Error", "ScheduleResource Update: StartWorkflow was nil -- report this to devs")
				return nil, temporalSdk.ErrSkipScheduleUpdate
			}
			scheduleUpdate.Schedule.Action = &temporalClient.ScheduleWorkflowAction{
				ID:                       plan.StartWorkflow.WorkflowId.ValueString(),
				Workflow:                 plan.StartWorkflow.Workflow.ValueString(),
				TaskQueue:                plan.StartWorkflow.TaskQueue.ValueString(),
				WorkflowExecutionTimeout: toDuration(plan.StartWorkflow.WorkflowExecutionTimeout),
				WorkflowRunTimeout:       toDuration(plan.StartWorkflow.WorkflowRunTimeout),
				WorkflowTaskTimeout:      toDuration(plan.StartWorkflow.WorkflowTaskTimeout),
			}
			if plan.ScheduleSpec != nil {
				scheduleSpec := scheduleUpdate.Schedule.Spec
				//scheduleSpec.CronExpressions = toStringArray(ctx, plan.ScheduleSpec.Crons)
				// scheduleSpec.StartAt, diags = plan.ScheduleSpec.StartAt.ValueRFC3339Time()
				// resp.Diagnostics.Append(diags...)
				// scheduleSpec.EndAt, diags = plan.ScheduleSpec.EndAt.ValueRFC3339Time()
				// resp.Diagnostics.Append(diags...)

				scheduleSpec.Jitter = toDuration(plan.ScheduleSpec.Jitter)
				scheduleSpec.TimeZoneName = plan.ScheduleSpec.TimeZoneName.ValueString()
			}

			// TODO
			// // Schedule - Describes when Actions should be taken.
			// Spec *ScheduleSpec
			// // State - this schedules state
			// State *ScheduleState

			return scheduleUpdate, nil
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("ScheduleResource Update: Unable to Update Schedule %s : %s", scheduleID, err))
		return
	}

	// Fetch the new Schedule's description from the Server
	desc, err := r.tclient.ScheduleClient().GetHandle(ctx, plan.ScheduleId.ValueString()).Describe(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("ScheduleResource Update: Unable to describe created Schedule %s : %s", scheduleID, err))
		return
	}

	// Construct final STATE, including Temporal-created values
	state := tfschema.ScheduleResourceModel{}
	assignScheduleResourceModelFromDescription(scheduleID, desc.Schedule, &state)

	// Save updated data into Terraform state
	tflog.Info(ctx, fmt.Sprintf("Update ScheduledWorkflow resource %s", state.ScheduleId.ValueString()))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

//////////////////////////////////////////////////////////////////////////////

func (r *ScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform prior state data into the model
	var state tfschema.ScheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the Schedule from the Server
	scheduleID := state.ScheduleId.ValueString()
	err := r.tclient.ScheduleClient().GetHandle(ctx, scheduleID).Delete(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Temporal Error", fmt.Sprintf("ScheduleResource Delete: Unable to delete Schedule : %s", err))
	}
	tflog.Debug(ctx, fmt.Sprintf("Deleted ScheduledWorkflow resource %s", scheduleID))
}

//////////////////////////////////////////////////////////////////////////////

func (r *ScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

//////////////////////////////////////////////////////////////////////////////
// TODO: move us to a package

func toDuration(val types.String) time.Duration {
	duration, _ := time.ParseDuration(val.ValueString())
	return duration
}

func fromDuration(duration time.Duration) types.String {
	if duration == 0 {
		//return types.StringNull()
		return types.StringValue("0s")
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

func assignScheduleResourceModelFromDescription(createdScheduleID string, sched temporalClient.Schedule, model *tfschema.ScheduleResourceModel) {
	// Construct final STATE, including Temporal-created values
	model.ScheduleId = types.StringValue(createdScheduleID)
	if policy := sched.Policy; policy != nil {
		model.Overlap = types.StringValue(policy.Overlap.String())
		model.CatchupWindow = fromDuration(policy.CatchupWindow)
		model.PauseOnFailure = types.BoolValue(policy.PauseOnFailure)
	}
	if scheduleState := sched.State; scheduleState != nil {
		model.Note = fromString(scheduleState.Note)
		model.Paused = types.BoolValue(scheduleState.Paused)
		model.RemainingActions = types.Int64Value(int64(scheduleState.RemainingActions))
		// TriggerImmediately isn't stored in Temporal state, so we clear it out here
		model.TriggerImmediately = types.BoolValue(false)
	}
	// TODO: memo, search attributes, scheduleBackfill

	if sched.Spec != nil {
		model.ScheduleSpec = &tfschema.ScheduleSpecModel{
			Jitter:       fromDuration(sched.Spec.Jitter),
			TimeZoneName: types.StringValue(sched.Spec.TimeZoneName),
			//Calendars:    fromTemporalCalendars(sched.Spec.Calendars),
			// StartAt: timetypes.NewRFC3339TimeValue(sched.Spec.StartAt),
			// EndAt:   timetypes.NewRFC3339TimeValue(sched.Spec.EndAt),
		}
	}

	switch newAction := sched.Action.(type) {
	case *temporalClient.ScheduleWorkflowAction:
		model.StartWorkflow = &tfschema.ScheduleWorkflowActionResourceModel{
			WorkflowId:               types.StringValue(newAction.ID),
			Workflow:                 getWorkflowName(newAction.Workflow),
			TaskQueue:                types.StringValue(newAction.TaskQueue),
			WorkflowExecutionTimeout: fromDuration(newAction.WorkflowExecutionTimeout),
			WorkflowRunTimeout:       fromDuration(newAction.WorkflowRunTimeout),
			WorkflowTaskTimeout:      fromDuration(newAction.WorkflowTaskTimeout),
		}
	}
}

func toStringArray(ctx context.Context, list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var strArr []string
	list.ElementsAs(ctx, &strArr, true)
	return strArr
}

func fromStringArray(strArr []string) types.List {
	elems := make([]attr.Value, len(strArr))
	for i, str := range strArr {
		elems[i] = types.StringValue(str)
	}
	list, _ := types.ListValue(types.StringType, elems)
	return list
}

func toTemporalCalendars(calendarModels []tfschema.ScheduleCalendarSpecModel) []temporalClient.ScheduleCalendarSpec {
	var tempCalendars []temporalClient.ScheduleCalendarSpec
	for _, calModel := range calendarModels {
		tempCalendar := temporalClient.ScheduleCalendarSpec{
			Second:     toTemporalScheduleRanges(calModel.Second),
			Minute:     toTemporalScheduleRanges(calModel.Minute),
			Hour:       toTemporalScheduleRanges(calModel.Hour),
			DayOfMonth: toTemporalScheduleRanges(calModel.DayOfMonth),
			Month:      toTemporalScheduleRanges(calModel.Month),
			Year:       toTemporalScheduleRanges(calModel.Year),
			DayOfWeek:  toTemporalScheduleRanges(calModel.DayOfWeek),
			Comment:    calModel.Comment.ValueString(),
		}
		tempCalendars = append(tempCalendars, tempCalendar)
	}
	return tempCalendars
}

func fromTemporalCalendars(tempModels []temporalClient.ScheduleCalendarSpec) []tfschema.ScheduleCalendarSpecModel {
	var calModels []tfschema.ScheduleCalendarSpecModel
	for _, tempModel := range tempModels {
		calModel := tfschema.ScheduleCalendarSpecModel{
			Second:     fromTemporalScheduleRanges(tempModel.Second),
			Minute:     fromTemporalScheduleRanges(tempModel.Minute),
			Hour:       fromTemporalScheduleRanges(tempModel.Hour),
			DayOfMonth: fromTemporalScheduleRanges(tempModel.DayOfMonth),
			Month:      fromTemporalScheduleRanges(tempModel.Month),
			Year:       fromTemporalScheduleRanges(tempModel.Year),
			DayOfWeek:  fromTemporalScheduleRanges(tempModel.DayOfWeek),
			Comment:    types.StringValue(tempModel.Comment),
		}
		calModels = append(calModels, calModel)
	}
	return calModels
}

func toTemporalScheduleRanges(ranges []tfschema.ScheduleRangeModel) []temporalClient.ScheduleRange {
	var tempRanges []temporalClient.ScheduleRange
	for _, rangeModel := range ranges {
		tempRange := temporalClient.ScheduleRange{
			Start: int(rangeModel.Start.ValueInt64()),
			End:   int(rangeModel.End.ValueInt64()),
			Step:  int(rangeModel.Step.ValueInt64()),
		}
		tempRanges = append(tempRanges, tempRange)
	}
	return tempRanges
}

func fromTemporalScheduleRanges(tempRanges []temporalClient.ScheduleRange) []tfschema.ScheduleRangeModel {
	var rangeModels []tfschema.ScheduleRangeModel
	for _, tempRange := range tempRanges {
		rangeModel := tfschema.ScheduleRangeModel{
			Start: types.Int64Value(int64(tempRange.Start)),
			End:   types.Int64Value(int64(tempRange.End)),
			Step:  types.Int64Value(int64(tempRange.Step)),
		}
		rangeModels = append(rangeModels, rangeModel)
	}
	return rangeModels
}
