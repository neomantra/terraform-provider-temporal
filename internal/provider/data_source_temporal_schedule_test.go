package provider

// // func TestAccScheduleDataSource(t *testing.T) {
// // 	resource.Test(t, resource.TestCase{
// // 		PreCheck:                 func() { testAccPreCheck(t) },
// // 		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
// // 		Steps: []resource.TestStep{
// // 			// Read testing
// // 			{
// // 				Config: providerConfig + testAccScheduleDataSourceConfig,
// // 				Check:  nil,
// // 				// Check: resource.ComposeAggregateTestCheckFunc(
// // 				// 	resource.TestCheckNoResourceAttr("data.temporal_schedule.fail", "id"),
// // 				// ),
// // 			},
// // 		},
// // 	})
// // }

// // const testAccScheduleDataSourceConfig = `
// // resource "temporal_schedule" "test-full" {
// //   id = "schedule-data-source-test"
// //   action {
// //     start_workflow {
// //       workflow   = "my-workflow-type"
// //       task_queue = "test-queue"
// //     }
// //   }
// // }
// // data "temporal_schedule" "data-source-test" {
// //  id = "schedule-data-source-test"
// // }
// //
//
