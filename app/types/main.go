package types

type ResourceType string

const (
	ECSResource             ResourceType = "aws-ecs"
	RDSResource             ResourceType = "aws-rds"
	ELBResource             ResourceType = "aws-elb"
	CloudflarePagesResource ResourceType = "cloudflare-pages"
)

type ResourceDefinition struct {
	Name       string       `json:"name"`
	Identifier string       `json:"identifier"`
	Type       ResourceType `json:"type"`
}

type DeploymentDefinition struct {
	Name      string               `json:"name"`
	Resources []ResourceDefinition `json:"resources"`
}

type ProjectDefinition struct {
	Name        string                 `json:"name"`
	Deployments []DeploymentDefinition `json:"deployments"`
}

type ResourceStatus interface {
	IsResourceStatus()
	IsHealthy() bool
	GetStatusString() string
}

type ResourceSnapshot struct {
	Definition ResourceDefinition `json:"definition"`
	Status     ResourceStatus     `json:"status"`
	Healthy    bool               `json:"healthy"`
}
