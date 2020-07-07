package ovh

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestKubernetesClusterDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testPreCheckKubernetes(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testKubernetesClusterDatasourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testKubernetesClusterDatasource("data.ovh_kubernetes_cluster.cluster"),
				),
			},
		},
	})
}

func testKubernetesClusterDatasource(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("can't find data: %s", n)
		}

		return nil
	}
}

var testKubernetesClusterDatasourceConfig = fmt.Sprintf(`
data "ovh_kubernetes_cluster" "cluster" {
  project_id = "%s"
  name = "%s"
}
`, os.Getenv("OVH_PUBLIC_CLOUD"), os.Getenv("OVH_KUBERNETES_CLUSTER_NAME"))
