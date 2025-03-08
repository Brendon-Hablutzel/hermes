package cloudflare

import (
	"context"
	"fmt"
	"hermes/app/types"
	"os"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/pages"
)

var _ types.ResourceStatus = PagesStatus{}

type PagesStatus struct {
	CanonicalDeploymentStatus string
	CanonicalDeploymentUrl    string
}

func (p PagesStatus) IsResourceStatus() {}

func (p PagesStatus) IsHealthy() bool {
	return p.CanonicalDeploymentStatus == "success"
}

func (p PagesStatus) GetStatusString() string {
	return p.CanonicalDeploymentStatus
}

func GetPagesStatus(client *cloudflare.Client, projectName string) (PagesStatus, error) {
	accountId, found := os.LookupEnv("CLOUDFLARE_ACCOUNT_ID")

	if !found {
		return PagesStatus{}, fmt.Errorf("CLOUDFLARE_ACCOUNT_ID not found")
	}

	project, err := client.Pages.Projects.Get(
		context.TODO(),
		projectName,
		pages.ProjectGetParams{
			AccountID: cloudflare.F(accountId),
		},
	)

	if err != nil {
		return PagesStatus{}, nil
	}

	// TODO: specific deployments + pagination
	// deployments, err := client.Pages.Projects.Deployments.List(
	// 	context.TODO(),
	// 	projectName, pages.ProjectDeploymentListParams{
	// 		AccountID: cloudflare.F(accountId),
	// 	},
	// )

	// if err != nil {
	// 	return CloudflarePagesDetails{}, nil
	// }

	// j, _ := json.MarshalIndent(deployments, "", "  ")
	// fmt.Println(string(j))

	return PagesStatus{
		CanonicalDeploymentStatus: project.CanonicalDeployment.LatestStage.Status,
		CanonicalDeploymentUrl:    project.CanonicalDeployment.URL,
	}, nil
}

func GetCloudflareClient() *cloudflare.Client {
	// reads variables from environment by default
	return cloudflare.NewClient()
}
