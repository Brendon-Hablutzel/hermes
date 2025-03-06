package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"

	"hermes/app/aws"
	"hermes/app/cloudflare"
	"hermes/app/types"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	cloudflare_sdk "github.com/cloudflare/cloudflare-go/v4"
)

func findProject(projects []types.ProjectDefinition, name string) (types.ProjectDefinition, bool) {
	projectIdx := slices.IndexFunc(
		projects,
		func(p types.ProjectDefinition) bool { return p.Name == name },
	)
	if projectIdx == -1 {
		return types.ProjectDefinition{}, false
	}

	project := projects[projectIdx]

	return project, true
}

func findDeployment(project types.ProjectDefinition, name string) (types.DeploymentDefinition, bool) {
	deployments := project.Deployments

	deploymentIdx := slices.IndexFunc(
		deployments,
		func(d types.DeploymentDefinition) bool { return d.Name == name },
	)
	if deploymentIdx == -1 {
		return types.DeploymentDefinition{}, false
	}

	deployment := deployments[deploymentIdx]

	return deployment, true
}

func findResource(deployment types.DeploymentDefinition, name string) (types.ResourceDefinition, bool) {
	resources := deployment.Resources

	resourceIdx := slices.IndexFunc(
		resources,
		func(r types.ResourceDefinition) bool { return r.Name == name },
	)
	if resourceIdx == -1 {
		return types.ResourceDefinition{}, false
	}

	resource := resources[resourceIdx]

	return resource, true
}

type GetProjectDefinitionResponse struct {
	Project types.ProjectDefinition `json:"project"`
}

func (s *Server) GetProjectDefinitionHandler(w http.ResponseWriter, r *http.Request) {
	projectName := r.PathValue("project")

	project, found := findProject(s.projects, projectName)
	if !found {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	err := json.NewEncoder(w).Encode(GetProjectDefinitionResponse{
		Project: project,
	})

	if err != nil {
		log.Println("failed to encode get project definition response", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}

}

func (s *Server) GetResourceSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	projectName := r.PathValue("project")
	deploymentName := r.PathValue("deployment")
	resourceName := r.PathValue("resource")

	project, found := findProject(s.projects, projectName)
	if !found {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	deployment, found := findDeployment(project, deploymentName)
	if !found {
		http.Error(w, "deployment not found", http.StatusNotFound)
		return
	}

	resource, found := findResource(deployment, resourceName)
	if !found {
		http.Error(w, "resource not found", http.StatusNotFound)
		return
	}

	var status types.ResourceStatus
	var err error
	switch resource.Type {
	case types.ECSResource:
		status, err = aws.GetECSStatus(s.ecsClient, resource.Identifier)
	case types.RDSResource:
		status, err = aws.GetRDSStatus(s.rdsClient, resource.Identifier)
	case types.ELBResource:
		status, err = aws.GetELBStatus(s.elbClient, resource.Identifier)
	case types.CloudflarePagesResource:
		status, err = cloudflare.GetPagesStatus(s.cloudflareClient, resource.Identifier)
	default:
		log.Println("invalid resource type encountered", resource.Type)
		http.Error(w, "invalid resource type", http.StatusInternalServerError)
		return
	}

	if err != nil {
		log.Println("error getting resource details", resource.Type, resource.Identifier, err)
		http.Error(w, "error getting resource details", http.StatusInternalServerError)
		return
	}

	snapshot := types.ResourceSnapshot{
		Definition: resource,
		Status:     status,
	}

	err = json.NewEncoder(w).Encode(snapshot)
	if err != nil {
		log.Println("failed to encode get resource snapshot response", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// type GetDeploymentSnapshotResponse struct {
// 	Resources []types.ResourceSnapshot `json:"resources"`
// }

// func getResourceSnapshots(s *Server, resources []types.ResourceDefinition) []types.ResourceSnapshot {
// 	ch := make(chan types.ResourceSnapshot, len(resources))
// 	var wg sync.WaitGroup

// 	for _, resource := range resources {
// 		wg.Add(1)
// 		fmt.Println("fetching")
// 		switch resource.Type {
// 		case types.ECSResource:
// 			go func() {
// 				defer wg.Done()

// 				res, err := aws.GetECSStatus(s.ecsClient, resource.Identifier)

// 				if err != nil {
// 					fmt.Println("ecs fetch err", err)
// 					return
// 				}

// 				snapshot := types.ResourceSnapshot{
// 					Definition: resource,
// 					Status:     res,
// 				}

// 				ch <- snapshot
// 			}()
// 		case types.RDSResource:
// 			go func() {
// 				defer wg.Done()

// 				res, err := aws.GetRDSStatus(s.rdsClient, resource.Identifier)

// 				if err != nil {
// 					fmt.Println("rds fetch err", err)
// 					return
// 				}

// 				snapshot := types.ResourceSnapshot{
// 					Definition: resource,
// 					Status:     res,
// 				}

// 				ch <- snapshot
// 			}()
// 		default:
// 			fmt.Println("invalid resource type")
// 		}
// 	}

// 	wg.Wait()
// 	close(ch)

// 	snapshots := []types.ResourceSnapshot{}
// 	for resource := range ch {
// 		snapshots = append(snapshots, resource)
// 	}

// 	return snapshots
// }

// func (s *Server) GetDeploymentSnapshotHandler(w http.ResponseWriter, r *http.Request) {
// 	projectName := r.PathValue("project")
// 	deploymentName := r.PathValue("deployment")

// 	project, found := findProject(s.projects, projectName)
// 	if !found {
// 		http.Error(w, "project not found", http.StatusNotFound)
// 	}

// 	deployment, found := findDeployment(project, deploymentName)
// 	if !found {
// 		http.Error(w, "deployment not found", http.StatusNotFound)
// 	}

// 	resources := getResourceSnapshots(s, deployment.Resources)

// 	resp := GetDeploymentSnapshotResponse{
// 		Resources: resources,
// 	}

// 	err := json.NewEncoder(w).Encode(resp)
// 	if err != nil {
// 		log.Println("failed to encode get resource snapshot response", err)
// 		http.Error(w, "failed to encode response", http.StatusInternalServerError)
// 	}
// }

type Server struct {
	ecsClient        *ecs.Client
	rdsClient        *rds.Client
	elbClient        *elasticloadbalancingv2.Client
	cloudflareClient *cloudflare_sdk.Client
	projects         []types.ProjectDefinition
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Println(r.Method, r.URL.Path)

			next.ServeHTTP(w, r)
		},
	)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		},
	)
}

func getProjectDefinitions() ([]types.ProjectDefinition, error) {
	data, err := os.ReadFile("projects.json")
	if err != nil {
		return []types.ProjectDefinition{}, err
	}

	var projectDefinitions []types.ProjectDefinition
	err = json.Unmarshal(data, &projectDefinitions)
	if err != nil {
		return []types.ProjectDefinition{}, err
	}

	return projectDefinitions, nil
}

func main() {
	for _, requiredEnvVar := range []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_REGION",
		"CLOUDFLARE_EMAIL",
		"CLOUDFLARE_API_KEY",
		"CLOUDFLARE_ACCOUNT_ID",
	} {
		_, found := os.LookupEnv(requiredEnvVar)

		if !found {
			fmt.Printf("required environment variable %s not found\n", requiredEnvVar)
			os.Exit(1)
		}
	}

	projectDefinitions, err := getProjectDefinitions()

	if err != nil {
		fmt.Println("error getting project definitions", err)
		os.Exit(1)
	}

	fmt.Println(projectDefinitions)

	ecsClient, err := aws.GetECSClient()
	if err != nil {
		log.Println("error getting ecs client", err)
		os.Exit(1)
	}

	rdsClient, err := aws.GetRDSClient()
	if err != nil {
		log.Println("error getting rds client", err)
		os.Exit(1)
	}

	elbClient, err := aws.GetELBClient()
	if err != nil {
		log.Println("error getting elb client", err)
		os.Exit(1)
	}

	cloudflareClient := cloudflare.GetCloudflareClient()

	server := &Server{
		ecsClient,
		rdsClient,
		elbClient,
		cloudflareClient,
		projectDefinitions,
	}

	router := http.NewServeMux()

	router.HandleFunc("/projects/{project}", server.GetProjectDefinitionHandler)
	// router.HandleFunc("/projects/{project}/deployments/{deployment}/snapshot", server.GetDeploymentSnapshotHandler)
	router.HandleFunc("/projects/{project}/deployments/{deployment}/resources/{resource}/snapshot", server.GetResourceSnapshotHandler)

	configuredRouter := corsMiddleware(loggingMiddleware(router))

	log.Println("Server running on :8080")
	err = http.ListenAndServe(":8080", configuredRouter)

	if err != nil {
		log.Println("error starting server", err)
		os.Exit(1)
	}
}
