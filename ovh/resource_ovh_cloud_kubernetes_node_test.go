package ovh

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testKubernetesNodeConfig = fmt.Sprintf(`
data "ovh_kubernetes_cluster" "cluster" {
	project_id  = "%s"
	name = "%s"
}
resource "ovh_kubernetes_node" "node" {
	project_id  = "%s"
	cluster_id = data.ovh_kubernetes_cluster.cluster.id
  	name = "test-node"
	flavor = "b2-7"
}
`, os.Getenv("OVH_PUBLIC_CLOUD"), os.Getenv("OVH_KUBERNETES_CLUSTER_NAME"), os.Getenv("OVH_PUBLIC_CLOUD"))

func TestKubernetesNode_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testPreCheckKubernetes(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckKubernetesNodeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testKubernetesNodeConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckKubernetesNodeExists("ovh_kubernetes_node.node", t),
				),
			},
		},
	})
}

func testCheckKubernetesNodeExists(n string, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		if rs.Primary.Attributes["project_id"] == "" {
			return fmt.Errorf("no Project ID is set")
		}

		return KubernetesNodeExists(rs.Primary.Attributes["project_id"], rs.Primary.Attributes["cluster_id"], rs.Primary.ID, config.OVHClient)
	}
}

func testCheckKubernetesNodeDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ovh_kubernetes_cluster" {
			continue
		}

		err := KubernetesNodeExists(rs.Primary.Attributes["project_id"], rs.Primary.Attributes["cluster_id"], rs.Primary.ID, config.OVHClient)
		if err == nil {
			return fmt.Errorf("cloud > Kubernetes Node still exist")
		}

	}
	return nil
}
