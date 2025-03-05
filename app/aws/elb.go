package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

type ELBStatus struct {
	Status  types.LoadBalancerStateEnum `json:"status"`
	DNSName string                      `json:"dns_name"`
}

func (e ELBStatus) IsResourceStatus() {}

func GetELBStatus(client *elasticloadbalancingv2.Client, elbName string) (ELBStatus, error) {
	result, err := client.DescribeLoadBalancers(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancersInput{
		Names: []string{elbName},
	})

	if err != nil {
		return ELBStatus{}, err
	}

	if len(result.LoadBalancers) == 0 {
		return ELBStatus{}, fmt.Errorf("no load balancers found")
	}

	loadBalancer := result.LoadBalancers[0]

	return ELBStatus{
		Status:  loadBalancer.State.Code,
		DNSName: *loadBalancer.DNSName,
	}, nil
}

func GetELBClient() (*elasticloadbalancingv2.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return elasticloadbalancingv2.NewFromConfig(cfg), nil
}
