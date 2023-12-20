package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNamespaceDataSource(t *testing.T) {
	resourcePath := "temporal_namespace.test-full"
	dataSourcePath := "data.temporal_namespace.data-source-test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: providerConfig + testAccNamespaceDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourcePath, "name", dataSourcePath, "name"),
					//resource.TestCheckResourceAttrPair(resourcePath, "id", dataSourcePath, "id"),
					resource.TestCheckResourceAttrPair(resourcePath, "description", dataSourcePath, "description"),
					resource.TestCheckResourceAttrPair(resourcePath, "owner_email", dataSourcePath, "owner_email"),
					resource.TestCheckResourceAttrPair(resourcePath, "retention_hours", dataSourcePath, "retention_hours"),
				),
			},
		},
	})
}

const testAccNamespaceDataSourceConfig = `
resource "temporal_namespace" "test-full" {
  name            = "namespace-data-source-test"
  retention_hours = 420
  description     = "test description"
  owner_email     = "test@test.com"
}
data "temporal_namespace" "data-source-test" {
  name = temporal_namespace.test-full.name
}
`
