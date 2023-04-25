package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccScheduleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccScheduleResourceConfig("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("temporal_schedule.test", "id", "example-id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "temporal_schedule.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				// ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			},
			// Update and Read testing
			// {
			// 	Config: testAccScheduleResourceConfig("two"),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("temporal_schedule.test", "configurable_attribute", "two"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccScheduleResourceConfig(configurableAttribute string) string {
	return `
resource "temporal_schedule" "test" {
  id = "example-id"
}`
}
