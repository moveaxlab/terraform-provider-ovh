package ovh

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testKubernetesClusterConfig = fmt.Sprintf(`
resource "ovh_kubernetes_cluster" "cluster" {
	project_id  = "%s"
  	name = "Test Cluster "
	region = "GRA7"
	version = "1.18"
}
`, os.Getenv("OVH_PUBLIC_CLOUD"))

func TestKubernetesCluster_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testPreCheckKubernetes(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testKubernetesClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testKubernetesClusterConfig,
				Check: resource.ComposeTestCheckFunc(
					testKubernetesClusterExists("ovh_kubernetes_cluster.cluster", t),
				),
			},
		},
	})
}

func testKubernetesClusterExists(n string, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		if rs.Primary.Attributes["project_id"] == "" {
			return fmt.Errorf("No Project ID is set")
		}

		return cloudKubernetesClusterExists(rs.Primary.Attributes["project_id"], rs.Primary.ID, config.OVHClient)
	}
}

func testKubernetesClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovh_kubernetes_cluster" {
			continue
		}

		err := cloudKubernetesClusterExists(rs.Primary.Attributes["project_id"], rs.Primary.ID, config.OVHClient)
		if err == nil {
			return fmt.Errorf("cloud > Kubernetes Cluster still exists")
		}

	}
	return nil
}
