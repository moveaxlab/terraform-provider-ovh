package ovh

import (
	"fmt"
)

// OVH Kubernetes
type KubernetesCreateOpts struct {
	serviceName		string		`json:"serviceName"`
	Name      		string   	`json:"name"`
	Region   		string	 	`json:"region"`
	Version			string		`json:"version"`
	//ProjectKubeCreationNodePool *KubernetesNodePool `json:"ProjectKubeCreationNodePool"`
}

//type KubernetesNodePool struct {
//	nodeName		string		`json:"name"`
//	DesiredNodes	int      	`json:"desiredNodes"`
//	FlavorName		string		`json:"flavorName"`
//	MaxNodes		int 		`json:"maxNodes"`
//	MinNodes		int			`json:"minNodes"`
//}


func (p *KubernetesCreateOpts) String() string {
	return fmt.Sprintf("projectId: %s, name:%s, region: %s, version: %s", p.serviceName, p.Name, p.Region, p.Version)
}

//func (opts *KubernetesCreateOpts) FromResource(d *schema.ResourceData) *KubernetesCreateOpts  {
//	kubernetesNodePool := opts.ProjectKubeCreationNodePool
//
//}

type KubernetesCreateResponse struct {
	Id      string                      `json:"id"`
	Status  string                      `json:"status"`
	Url		string						`json:"url"`
	NodesUrl string						`json:"nodesUrl"`
	Version  string						`json:"version"`

}

func (p *KubernetesCreateResponse) String() string {
	return fmt.Sprintf("Id: %s, Status: %s, Url: %s. NodesUrl: %s, Version: %s", p.Id, p.Status,p.Url, p.NodesUrl, p.Version )

}

type KubernetesUpdateOpts struct {
	Name string 					`json:"name"`

}


