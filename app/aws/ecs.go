package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

type ECSStatus struct {
	Status       string       `json:"status"`
	TasksPending int          `json:"tasks_pending"`
	TasksRunning int          `json:"tasks_running"`
	Services     []ECSService `json:"services"`
}

func (e ECSStatus) IsResourceStatus() {}

type ECSService struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	DesiredCount int       `json:"desired_count"`
	PendingCount int       `json:"pending_count"`
	RunningCount int       `json:"running_count"`
}

func GetECSStatus(client *ecs.Client, clusterIdentifier string) (ECSStatus, error) {
	resp, err := client.DescribeClusters(context.TODO(), &ecs.DescribeClustersInput{
		Clusters: []string{clusterIdentifier},
	})

	if err != nil {
		return ECSStatus{}, err
	}

	if len(resp.Clusters) == 0 {
		return ECSStatus{}, fmt.Errorf("no cluster found")
	}

	firstCluster := resp.Clusters[0]

	listServicesResp, err := client.ListServices(context.TODO(), &ecs.ListServicesInput{
		Cluster: &clusterIdentifier,
	})

	if err != nil {
		return ECSStatus{}, err
	}

	servicesResp, err := client.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
		Cluster:  &clusterIdentifier,
		Services: listServicesResp.ServiceArns,
	})

	if err != nil {
		return ECSStatus{}, err
	}

	services := []ECSService{}
	for _, service := range servicesResp.Services {
		services = append(services,
			ECSService{
				Status:       *service.Status,
				CreatedAt:    *service.CreatedAt,
				DesiredCount: int(service.DesiredCount),
				PendingCount: int(service.PendingCount),
				RunningCount: int(service.RunningCount),
				Name:         *service.ServiceName,
			},
		)
	}

	return ECSStatus{
		Status:       *firstCluster.Status,
		TasksPending: int(firstCluster.PendingTasksCount),
		TasksRunning: int(firstCluster.RunningTasksCount),
		Services:     services,
	}, nil
}

func GetECSClient() (*ecs.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return ecs.NewFromConfig(cfg), nil
}
