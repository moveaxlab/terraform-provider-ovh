package ovh

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/ovh/go-ovh/ovh"
)

func resourceKubernetesState(
	d *schema.ResourceData,
	meta interface{}) ([]*schema.ResourceData, error) {
	givenId := d.Id()
	splitId := strings.SplitN(givenId, "/", 2)
	if len(splitId) != 2 {
		return nil, fmt.Errorf("Import Id is not OVH_PROJECT_ID formatted ")
	}
	d.SetId(splitId[1])
	d.Set("service_name", splitId[0])
	results := make([]*schema.ResourceData, 1)
	results[0] = d
	//log.Printf("[DEBUG] result %s", results)
	return results, nil
}

func resourceKubernetes() *schema.Resource {
	return &schema.Resource{
		Create: resourceKubernetesCreate,
		Read:   resourceKubernetesRead,
		Update: resourceKubernetesUpdate,
		Delete: resourceKubernetesDelete,
		Importer: &schema.ResourceImporter{
			State: resourceKubernetesState,
		},

		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				DefaultFunc: schema.EnvDefaultFunc("OVH_PROJECT_ID", nil),
			},
			"name": {
				Type:     schema.TypeString,
				//Description: "This attribute is there for documentation purpose only and isnt passed to the OVH API as it may conflicts with http/tcp  ",
				Required: true,
				ForceNew: false,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
				//Description: "This attribute is there for documentation purpose only and isnt passed to the OVH API as it may conflicts with http/tcp  ",
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"version": {
				Type:     schema.TypeString,
				Required: true,
			},
			//"desired_nodes": {
			//	Type:     schema.TypeFloat,
			//	//Description: "This attribute is there for documentation purpose only and isnt passed to the OVH API as it may conflicts with http/tcp  ",
			//	Optional: true,
			//},
			//"min_nodes": {
			//	Type:     schema.TypeFloat,
			//	//Description: "This attribute is there for documentation purpose only and isnt passed to the OVH API as it may conflicts with http/tcp ",
			//	Optional: true,
			//},
			//"max_nodes": {
			//	Type:     schema.TypeFloat,
			//	//Description: "This attribute is there for documentation purpose only and isnt passed to the OVH API as it may conflicts with http/tcp ",
			//	Optional: true,
			//},
			//"nodes_prefix" : {
			//	Type: schema.TypeString,
			//	//Description: "This attribute is there for documentation purpose only and isnt passed to the OVH API as it may conflicts with http/tcp ",
			//	Optional: true,
			//},
			//"flavor" : {
			//	Type: schema.TypeString,
			//	//Description: "This attribute is there for documentation purpose only and isnt passed to the OVH API as it may conflicts with http/tcp ",
			//	Optional: true,
			//},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceKubernetesCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	projectId := d.Get("service_name").(string)
	params := &KubernetesCreateOpts{
		serviceName:  d.Get("service_name").(string),
		Name:         d.Get("name").(string),
		Version:      d.Get("version").(string),
		//DesiredNodes: d.Get("desired_nodes").(int),
		//MaxNodes:     d.Get("max_nodes").(int),
		//MinNodes:     d.Get("min_nodes").(int),
		//FlavorName:   d.Get("flavor").(string),
		Region:       d.Get("region").(string),
		//nodeName:     d.Get("nodes_prefix").(string),
	}

	r := &KubernetesCreateResponse{}

	log.Printf("[DEBUG] Will create a Managed Kuberenetes Cluster: %s", params)

	endpoint := fmt.Sprintf("/cloud/project/%s/kube", params.serviceName)

	err := config.OVHClient.Post(endpoint, params, r)
	if err != nil {
		return fmt.Errorf("calling %s with params %s:\n\t %q", endpoint, params, err)
	}

	log.Printf("[DEBUG] Waiting for Kubernetes Cluster %s:", r)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"INSTALLING"},
		Target:     []string{"READY"},
		Refresh:    waitForKubernetesActive(config.OVHClient, projectId, r.Id),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("waiting for Kubernetes Cluster (%s): %s", params, err)
	}
	log.Printf("[DEBUG] Created Kubernetes Cluster %s", r)

	//set id
	d.SetId(r.Id)

	return nil
}

func resourceKubernetesRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	projectId := d.Get("service_name").(string)

	r := &KubernetesCreateResponse{}

	log.Printf("[DEBUG] Will read Kubernetes Cluster for service: %s, id: %s", projectId, d.Id())

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", projectId, d.Id())

	d.Partial(true)
	err := config.OVHClient.Get(endpoint, r)
	if err != nil {
		return fmt.Errorf("Error calling %s:\n\t %q", endpoint, err)
	}

	err = readKubernetes(config, d, r)
	if err != nil {
		return err
	}
	d.Partial(false)

	log.Printf("[DEBUG] Read/List Kubernetes Clusters %s", r)
	return nil
}

func resourceKubernetesUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	projectId := d.Get("service_name").(string)
	params := &KubernetesUpdateOpts{
		Name: d.Get("name").(string),
	}

	log.Printf("[DEBUG] Will update Kubernetes Cluster: %s", params)

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", projectId, d.Id())

	err := config.OVHClient.Put(endpoint, params, nil)
	if err != nil {
		return fmt.Errorf("calling %s with params %s:\n\t %q", endpoint, params, err)
	}

	log.Printf("[DEBUG] Updated kubernetes cluster %s %s:", projectId, d.Id())

	return resourceKubernetesRead(d, meta)
}

func resourceKubernetesDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	projectId := d.Get("service_name").(string)
	id := d.Id()

	log.Printf("[DEBUG] Will delete kubernetes cluster for service: %s, and id: %s", projectId, id)

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", projectId, id)

	err := config.OVHClient.Delete(endpoint, nil)
	if err != nil {
		return fmt.Errorf("calling %s:\n\t %q", endpoint, err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{"DELETED"},
		Refresh:    waitForKubernetesDelete(config.OVHClient, projectId, id),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("deleting kubernetes cluster (%s): %s", id, err)
	}

	d.SetId("")

	log.Printf("[DEBUG] Deleted %s Kubernetes Cluster %s", projectId, id)
	return nil
}

func readKubernetes(config *Config, d *schema.ResourceData, r *KubernetesCreateResponse) error {

	d.Set("status", r.Status)

	d.SetId(r.Id)
	return nil
}

func kubernetesClusterExists(projectId, id string, c *ovh.Client) error {
	r := &KubernetesCreateResponse{}

	log.Printf("[DEBUG] Will Get kubernetes cluster : %s, id: %s", projectId, id)

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", projectId, id)

	err := c.Get(endpoint, r)
	if err != nil {
		return fmt.Errorf("calling %s:\n\t %q", endpoint, err)
	}
	log.Printf("[DEBUG] Get Kubernetes cluster: %s", r)

	return nil
}

// AttachmentStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// an Attachment Task.
func waitForKubernetesActive(c *ovh.Client, projectId, KubernetesId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r := &KubernetesCreateResponse{}
		endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", projectId, KubernetesId)
		err := c.Get(endpoint, r)
		if err != nil {
			return r, "", err
		}

		log.Printf("[DEBUG] Pending Kubernetes Cluster: %s", r)
		return r, r.Status, nil
	}
}

// AttachmentStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// an Attachment Task.
func waitForKubernetesDelete(c *ovh.Client, projectId, KubernetesId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		r := &KubernetesCreateResponse{}
		endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", projectId, KubernetesId)
		err := c.Get(endpoint, r)
		if err != nil {
			if err.(*ovh.APIError).Code == 404 {
				log.Printf("[DEBUG] Kubernetes Cluster %s on project %s deleted", KubernetesId, projectId)
				return r, "DELETED", nil
			} else {
				return r, "", err
			}
		}
		log.Printf("[DEBUG] Pending Kubernetes Cluster: %s", r)
		return r, r.Status, nil
	}
}
