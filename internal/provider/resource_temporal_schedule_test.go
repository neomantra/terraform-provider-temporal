package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScheduleResource_Minimal(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccScheduleResourceConfig_Minimal("task-queue-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("temporal_schedule.test-minimal", "id", "test-schedule-minimal"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "temporal_schedule.test-minimal",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				// ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			},
			// Update and Read testing
			{
				Config: testAccScheduleResourceConfig_Minimal("task-queue-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("temporal_schedule.test-minimal", "start_workflow.task_queue", "task-queue-2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccScheduleResourceConfig_Minimal(taskQueue string) string {
	return fmt.Sprintf(`
resource "temporal_schedule" "test-minimal" {
  id = "test-schedule-minimal"
  start_workflow = {
    workflow   = "my-workflow-type"
    task_queue = "%s"   // one dynamic template for testing
  }
}`, taskQueue)
}

// ///////////////////////////////////////////////////////////////////////////////

func TestAccScheduleResource_Full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccScheduleResourceConfig_Full("task-queue-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("temporal_schedule.test-full", "id", "test-schedule-full"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "temporal_schedule.test-full",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				// ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			},
			// Update and Read testing
			{
				Config: testAccScheduleResourceConfig_Full("task-queue-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("temporal_schedule.test-full", "start_workflow.task_queue", "task-queue-2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccScheduleResourceConfig_Full(taskQueue string) string {
	return fmt.Sprintf(`
resource "temporal_schedule" "test-full" {
  id = "test-schedule-full"
  schedule = {
   #crons = ["30 2 * * 5"]
   #start_at  = "2023-04-20T04:20:01Z"
   #end_at    = "2024-08-31T19:00:01Z"
   jitter    = "15s"
   time_zone = "Antarctica/South_Pole"
  }
  start_workflow = {
    workflow_id       = "my-workflow-id"
    task_queue        = "%s"   // one dynamic template for testing
    workflow          = "my-workflow-type"
    # args            = ["one", "two", 3]
    # execution_timeout = "30m0s"
    # run_timeout       = "1h0m0s"
    # task_timeout      = "43s"
  }
  catchup_window      = "15s"
  pause_on_failure    = true
  note                = "hello"
  paused              = false
  remaining_actions   = 3
  trigger_immediately = false
}`, taskQueue)
}
