package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNamespaceResource_Minimal(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccNamespaceResourceConfig_Minimal(48),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("temporal_namespace.test-minimal", "name", "test-namespace-minimal"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "temporal_namespace.test-minimal",
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
				Config: testAccNamespaceResourceConfig_Minimal(24),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("temporal_namespace.test-minimal", "retention_hours", "24"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccNamespaceResourceConfig_Minimal(hours int64) string {
	return fmt.Sprintf(`
resource "temporal_namespace" "test-minimal" {
  name            = "test-namespace-minimal"
  retention_hours = %d
}`, hours)
}

// ///////////////////////////////////////////////////////////////////////////////

func TestAccNamespaceResource_Full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccNamespaceResourceConfig_Full("test-1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("temporal_namespace.test-full", "name", "test-namespace-full"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "temporal_namespace.test-full",
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
				Config: testAccNamespaceResourceConfig_Full("test-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("temporal_namespace.test-full", "description", "test-2"),
					resource.TestCheckResourceAttr("temporal_namespace.test-full", "owner_email", "test@test.com"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccNamespaceResourceConfig_Full(desc string) string {
	return fmt.Sprintf(`
resource "temporal_namespace" "test-full" {
  name            = "test-namespace-full"
  description     = "%s"
  owner_email     = "test@test.com"
  retention_hours = 48
}`, desc)
}
