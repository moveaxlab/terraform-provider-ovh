package ovh

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceKubernetesClusterRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_PROJECT_ID", nil),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kubernetes_version": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},

			"next_upgrade_versions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"is_up_to_date": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"control_plane_is_up_to_date": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"nodes_url": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"update_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"client_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"client_certificate": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"cluster_ca_certificate": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"kubeconfig": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceKubernetesClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	projectId := d.Get("project_id").(string)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Will read public cloud region %s for project: %s", name, projectId)
	d.Partial(true)

	cluster, err := findKubernetesCluster(config, projectId, name)

	if err != nil {
		return err
	}

	err = readCloudKubernetesCluster(projectId, config, d, &cluster)

	if err != nil {
		return fmt.Errorf("error while reading cloud config: %s", err)
	}

	d.Partial(false)
	d.SetId(cluster.Id)

	return nil
}

func findKubernetesCluster(config *Config, projectId string, name string) (cluster KubernetesClusterResponse, err error) {
	cluster = KubernetesClusterResponse{}
	response := []string{}
	endpoint := fmt.Sprintf("/cloud/project/%s/kube", projectId)
	err = config.OVHClient.Get(endpoint, &response)

	if err != nil {
		return
	}

	for i := 0; i < len(response); i++ {
		id := response[i]
		endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", projectId, id)
		err = config.OVHClient.Get(endpoint, &cluster)

		if err != nil {
			return
		}

		if cluster.Name == name {
			err = nil
			return
		}
	}
	err = fmt.Errorf("Cannot find cluster name : %s", name)
	return
}
