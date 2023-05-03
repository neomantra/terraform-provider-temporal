package provider

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	ScheduleId types.String `tfsdk:"id"`
	// TODOSpec               types.Object `tfsdk:"spec"`
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
	resp.Schema = schema.Schema{
		MarkdownDescription: "Schedule resource",
		Blocks: map[string]schema.Block{
			//"schedule": makeScheduleAttributeSchema(),
			"action": schema.SingleNestedBlock{
				MarkdownDescription: scheduleActionDocs,
				Blocks: map[string]schema.Block{
					"start_workflow": schema.SingleNestedBlock{
						//Required:   false,
						Attributes: makeStartWorkflowAttributes(),
					},
				},
			},
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: scheduleIDDocs,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"overlap": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(temporalEnums.SCHEDULE_OVERLAP_POLICY_SKIP.String()),
				MarkdownDescription: scheduleOverlapDocs,
				PlanModifiers: []planmodifier.String{
					//TODO
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"catchup_window": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("1m0s"),
				MarkdownDescription: scheduleCatchupWindowDocs,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pause_on_failure": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: schedulePauseOnFailureDocs,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"note": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: scheduleNoteDocs,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"paused": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: schedulePausedDocs,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"remaining_actions": schema.Int64Attribute{
				Computed:            true,
				Optional:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: scheduleRemainingActionsDocs,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"trigger_immediately": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: scheduleTriggerImmediatelyDocs,
				// PlanModifiers: []planmodifier.Bool{
				// boolplanmodifier.UseStateForUnknown(),
				// },
			},
			// TODO
			// "schedule_backfill": schema.BoolAttribute{
			// 	Optional:            true,
			// 	MarkdownDescription: scheduleScheduleBackfillDocs,
			// 	// PlanModifiers: []planmodifier.String{
			// 	// 	stringplanmodifier.UseStateForUnknown(),
			// 	// },
			// },
			// "memo_json": schema.StringAttribute{
			// 	Optional:            true,
			// 	MarkdownDescription: scheduleMemoDocs,
			// 	// PlanModifiers: []planmodifier.String{
			// 	// 	stringplanmodifier.UseStateForUnknown(),
			// 	// },
			// },
			// "search_attributes_json": schema.StringAttribute{
			// 	Optional:            true,
			// 	MarkdownDescription: scheduleSearchAttributesDocs,
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

func makeStartWorkflowAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"workflow_id": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			MarkdownDescription: scheduleWAIDDocs,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"workflow": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: scheduleWAWorkflowDocs,
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
			MarkdownDescription: scheduleWATaskQueueDocs,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"execution_timeout": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: scheduleWAWorkflowExecutionTimeoutDocs,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"run_timeout": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: scheduleWAWorkflowRunTimeoutDocs,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"task_timeout": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: scheduleWAWorkflowTaskTimeoutDocs,
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
		// 	MarkdownDescription: scheduleWAMemoDocs,
		// 	// PlanModifiers: []planmodifier.String{
		// 	// 	stringplanmodifier.UseStateForUnknown(),
		// 	// },
		// },
		// "search_attributes_json": schema.StringAttribute{
		// 	Optional:            true,
		// 	MarkdownDescription: scheduleWASearchAttributesDocs,
		// 	// PlanModifiers: []planmodifier.String{
		// 	// 	stringplanmodifier.UseStateForUnknown(),
		// 	// },
		// },
	}
}

//////////////////////////////////////////////////////////////////////////////

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

//////////////////////////////////////////////////////////////////////////////

const durationDocs = `A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w", "y".`

const scheduleIDDocs = `The business identifier of the schedule.`

const scheduleScheduleDocs = `Describes when Actions should be taken.`

const scheduleActionDocs = `Which Action to take. Currently only start_workflow is supported.`

const scheduleOverlapDocs = `Controls what happens when an Action would be started by a Schedule at the same time that an older Action is still
running. This can be changed after a Schedule has taken some Actions, and some changes might produce
unintuitive results. In general, the later policy overrides the earlier policy.
Optional: defaulted to SCHEDULE_OVERLAP_POLICY_SKIP`

const scheduleCatchupWindowDocs = `The Temporal Server might be down or unavailable at the time when a Schedule should take an Action.
When the Server comes back up, CatchupWindow controls which missed Actions should be taken at that point. The default is one
minute, which means that the Schedule attempts to take any Actions that wouldn't be more than one minute late. It
takes those Actions according to the Overlap. An outage that lasts longer than the Catchup
Window could lead to missed Actions.
Optional: defaulted to 1 minute`

const schedulePauseOnFailureDocs = `When an Action times out or reaches the end of its Retry Policy the Schedule will pause.
With SCHEDULE_OVERLAP_POLICY_ALLOW_ALL, this pause might not apply to the next Action, because the next Action
might have already started previous to the failed one finishing. Pausing applies only to Actions that are scheduled
to start after the failed one finishes.
Optional: defaulted to false`

const scheduleNoteDocs = `Informative human-readable message with contextual notes, e.g. the reason
a Schedule is paused. The system may overwrite this message on certain
conditions, e.g. when pause-on-failure happens.`

const schedulePausedDocs = `Start in paused state. Optional: defaulted to false`

const scheduleRemainingActionsDocs = `limit the number of Actions to take.
This number is decremented after each Action is taken, and Actions are not
taken when the number is '0' (unless ScheduleHandle.Trigger is called).
Optional: defaulted to zero`

const scheduleTriggerImmediatelyDocs = `Trigger one Action immediately on creating the schedule.
Optional: defaulted to false`

const scheduleScheduleBackfillDocs = `Runs though the specified time periods and takes Actions as if that time passed by right now, all at once. The
overlap policy can be overridden for the scope of the ScheduleBackfill.`

const scheduleMemoDocs = `Optional non-indexed info that will be shown in list schedules.`

const scheduleSearchAttributesDocs = `Optional indexed info that can be used in query of List schedules APIs (only
supported when Temporal server is using advanced visibility). The key and value type must be registered on Temporal server side.
Use GetSearchAttributes API to get valid key and corresponding value type.`

//////////////////////////////////////////////////////////////////////////////

const scheduleWAIDDocs = `The business identifier of the workflow execution.
The workflow ID of the started workflow may not match this exactly,
it may have a timestamp appended for uniqueness.
Optional: defaulted to a uuid.`

const scheduleWAWorkflowDocs = `Type name of the Workflow to run.`

const scheduleWAArgDocs = `Arguments to pass to the workflow.`

const scheduleWATaskQueueDocs = `The workflow tasks of the workflow are scheduled on the queue with this name.
This is also the name of the activity task queue on which activities are scheduled.`

const scheduleWAWorkflowExecutionTimeoutDocs = `The timeout for duration of workflow execution.`

const scheduleWAWorkflowRunTimeoutDocs = `The timeout for duration of a single workflow run.`

const scheduleWAWorkflowTaskTimeoutDocs = `The timeout for processing workflow task from the time the worker pulled this task.`

const scheduleWARetryPolicyDocs = `Retry policy for workflow. If a retry policy is specified, in case of workflow failure
server will start new workflow execution if needed based on the retry policy.`

const scheduleWAMemoDocs = `Optional non-indexed info that will be shown in list workflow.`

const scheduleWASearchAttributesDocs = `Optional indexed info that can be used in query of List/Scan/Count workflow APIs (only
 supported when Temporal server is using advanced visiblity). The key and value type must be registered on Temporal server side.`
