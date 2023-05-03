package docs

///////////////////////////////////////////////////////////////////////////////

const DurationDocs = `A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w", "y".`

const ScheduleIDDocs = `The business identifier of the schedule.`

const ScheduleScheduleDocs = `Describes when Actions should be taken.`

const ScheduleActionDocs = `Which Action to take. Currently only start_workflow is supported.`

const ScheduleOverlapDocs = `Controls what happens when an Action would be started by a Schedule at the same time that an older Action is still
running. This can be changed after a Schedule has taken some Actions, and some changes might produce
unintuitive results. In general, the later policy overrides the earlier policy.
Optional: defaulted to SCHEDULE_OVERLAP_POLICY_SKIP`

const ScheduleCatchupWindowDocs = `The Temporal Server might be down or unavailable at the time when a Schedule should take an Action.
When the Server comes back up, CatchupWindow controls which missed Actions should be taken at that point. The default is one
minute, which means that the Schedule attempts to take any Actions that wouldn't be more than one minute late. It
takes those Actions according to the Overlap. An outage that lasts longer than the Catchup
Window could lead to missed Actions.
Optional: defaulted to 1 minute`

const SchedulePauseOnFailureDocs = `When an Action times out or reaches the end of its Retry Policy the Schedule will pause.
With SCHEDULE_OVERLAP_POLICY_ALLOW_ALL, this pause might not apply to the next Action, because the next Action
might have already started previous to the failed one finishing. Pausing applies only to Actions that are scheduled
to start after the failed one finishes.
Optional: defaulted to false`

const ScheduleNoteDocs = `Informative human-readable message with contextual notes, e.g. the reason
a Schedule is paused. The system may overwrite this message on certain
conditions, e.g. when pause-on-failure happens.`

const SchedulePausedDocs = `Start in paused state. Optional: defaulted to false`

const ScheduleRemainingActionsDocs = `limit the number of Actions to take.
This number is decremented after each Action is taken, and Actions are not
taken when the number is '0' (unless ScheduleHandle.Trigger is called).
Optional: defaulted to zero`

const ScheduleTriggerImmediatelyDocs = `Trigger one Action immediately on creating the schedule.
Optional: defaulted to false`

const ScheduleScheduleBackfillDocs = `Runs though the specified time periods and takes Actions as if that time passed by right now, all at once. The
overlap policy can be overridden for the scope of the ScheduleBackfill.`

const ScheduleMemoDocs = `Optional non-indexed info that will be shown in list schedules.`

const ScheduleSearchAttributesDocs = `Optional indexed info that can be used in query of List schedules APIs (only
supported when Temporal server is using advanced visibility). The key and value type must be registered on Temporal server side.
Use GetSearchAttributes API to get valid key and corresponding value type.`

//////////////////////////////////////////////////////////////////////////////

const ScheduleSpecCronArgDocs = `CronExpressions -  CronExpressions-based specifications of times. CronExpressions is provided for easy migration from legacy Cron Workflows. For new
use cases, we recommend using ScheduleSpec.Calendars or ScheduleSpec.Intervals for readability and maintainability. Once a schedule is created all
expressions in CronExpressions will be translated to ScheduleSpec.Calendars on the server.

For example, \x60 0 12 * * MON-WED,FRI\x60 is every M/Tu/W/F at noon, and is equivalent to this ScheduleCalendarSpec:

client.ScheduleCalendarSpec{
		Second: []ScheduleRange{{}},
		Minute: []ScheduleRanges{{}},
		Hour: []ScheduleRange{{
			Start: 12,
		}},
		DayOfMonth: []ScheduleRange{
			{
				Start: 1,
				End:   31,
			},
		},
		Month: []ScheduleRange{
			{
				Start: 1,
				End:   12,
			},
		},
		DayOfWeek: []ScheduleRange{
			{
				Start: 1,
				End: 3,
			},
			{
				Start: 5,
			},
		},
	}

The string can have 5, 6, or 7 fields, separated by spaces, and they are interpreted in the
same way as a ScheduleCalendarSpec:
	- 5 fields:         Minute, Hour, DayOfMonth, Month, DayOfWeek
	- 6 fields:         Minute, Hour, DayOfMonth, Month, DayOfWeek, Year
	- 7 fields: Second, Minute, Hour, DayOfMonth, Month, DayOfWeek, Year

Notes:
	- If Year is not given, it defaults to *.
	- If Second is not given, it defaults to 0.
	- Shorthands @yearly, @monthly, @weekly, @daily, and @hourly are also
		accepted instead of the 5-7 time fields.
	- @every <interval>[/<phase>] is accepted and gets compiled into an
		IntervalSpec instead. <interval> and <phase> should be a decimal integer
		with a unit suffix s, m, h, or d.
	- Optionally, the string can be preceded by CRON_TZ=<time zone name> or
		TZ=<time zone name>, which will get copied to ScheduleSpec.TimeZoneName. (In which case the ScheduleSpec.TimeZone field should be left empty.)
	- Optionally, "#" followed by a comment can appear at the end of the string.
	- Note that the special case that some cron implementations have for
		treating DayOfMonth and DayOfWeek as "or" instead of "and" when both
		are set is not implemented.
`

//////////////////////////////////////////////////////////////////////////////

const ScheduleWAIDDocs = `The business identifier of the workflow execution.
The workflow ID of the started workflow may not match this exactly,
it may have a timestamp appended for uniqueness.
Optional: defaulted to a uuid.`

const ScheduleWAWorkflowDocs = `Type name of the Workflow to run.`

const ScheduleWAArgDocs = `Arguments to pass to the workflow.`

const ScheduleWATaskQueueDocs = `The workflow tasks of the workflow are scheduled on the queue with this name.
This is also the name of the activity task queue on which activities are scheduled.`

const ScheduleWAWorkflowExecutionTimeoutDocs = `The timeout for duration of workflow execution.`

const ScheduleWAWorkflowRunTimeoutDocs = `The timeout for duration of a single workflow run.`

const ScheduleWAWorkflowTaskTimeoutDocs = `The timeout for processing workflow task from the time the worker pulled this task.`

const ScheduleWARetryPolicyDocs = `Retry policy for workflow. If a retry policy is specified, in case of workflow failure
server will start new workflow execution if needed based on the retry policy.`

const ScheduleWAMemoDocs = `Optional non-indexed info that will be shown in list workflow.`

const ScheduleWASearchAttributesDocs = `Optional indexed info that can be used in query of List/Scan/Count workflow APIs (only
 supported when Temporal server is using advanced visiblity). The key and value type must be registered on Temporal server side.`
