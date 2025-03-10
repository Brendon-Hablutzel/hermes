package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"hermes/app/types"
	"os"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/pages"
)

var _ types.ResourceStatus = PagesStatus{}

type PagesStatus struct {
	InstanceExists            bool   `json:"exists"`
	CanonicalDeploymentStatus string `json:"status"`
	CanonicalDeploymentUrl    string `json:"url"`
}

func (p PagesStatus) IsResourceStatus() {}

func (p PagesStatus) IsHealthy() bool {
	return p.CanonicalDeploymentStatus == "success"
}

func (p PagesStatus) Exists() bool {
	return p.InstanceExists
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
		// if errors.As(err, )
		var cloudflareErr *cloudflare.Error
		if errors.As(err, &cloudflareErr) {
			if cloudflareErr.StatusCode == 404 {
				return PagesStatus{
					InstanceExists: false,
				}, nil
			}
		}

		return PagesStatus{}, err
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
		InstanceExists:            true,
		CanonicalDeploymentStatus: project.CanonicalDeployment.LatestStage.Status,
		CanonicalDeploymentUrl:    project.CanonicalDeployment.URL,
	}, nil
}

func GetCloudflareClient() *cloudflare.Client {
	// reads variables from environment by default
	return cloudflare.NewClient()
}
