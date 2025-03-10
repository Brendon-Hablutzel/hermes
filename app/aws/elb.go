package aws

import (
	"context"
	"errors"
	"fmt"
	"hermes/app/types"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elb_types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

var _ types.ResourceStatus = ELBStatus{}

type ELBStatus struct {
	InstanceExists bool                            `json:"exists"`
	Status         elb_types.LoadBalancerStateEnum `json:"status"`
	DNSName        string                          `json:"dns_name"`
}

func (e ELBStatus) IsResourceStatus() {}

func (e ELBStatus) IsHealthy() bool {
	return e.Status == "active"
}

func (e ELBStatus) Exists() bool {
	return e.InstanceExists
}

func (e ELBStatus) GetStatusString() string {
	return string(e.Status)
}

func GetELBStatus(client *elasticloadbalancingv2.Client, elbName string) (ELBStatus, error) {
	result, err := client.DescribeLoadBalancers(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancersInput{
		Names: []string{elbName},
	})

	if err != nil {
		var notFound *elb_types.LoadBalancerNotFoundException
		if errors.As(err, &notFound) {
			return ELBStatus{
				InstanceExists: false,
			}, nil
		}

		return ELBStatus{}, err
	}

	if len(result.LoadBalancers) == 0 {
		return ELBStatus{}, fmt.Errorf("no load balancers found")
	}

	loadBalancer := result.LoadBalancers[0]

	return ELBStatus{
		InstanceExists: true,
		Status:         loadBalancer.State.Code,
		DNSName:        *loadBalancer.DNSName,
	}, nil
}

func GetELBClient() (*elasticloadbalancingv2.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return elasticloadbalancingv2.NewFromConfig(cfg), nil
}
