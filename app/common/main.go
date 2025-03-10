package common

import (
	"fmt"

	"hermes/app/aws"
	"hermes/app/cloudflare"
	"hermes/app/types"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	cloudflare_sdk "github.com/cloudflare/cloudflare-go/v4"
)

type Clients struct {
	EcsClient        *ecs.Client
	RdsClient        *rds.Client
	ElbClient        *elasticloadbalancingv2.Client
	ApiGatewayClient *apigatewayv2.Client
	CloudflareClient *cloudflare_sdk.Client
}

func GetResourceStatus(c *Clients, resource types.ResourceDefinition) (types.ResourceStatus, error) {
	var status types.ResourceStatus
	var err error
	switch resource.Type {
	case types.ECSResource:
		status, err = aws.GetECSStatus(c.EcsClient, resource.Identifier)
	case types.RDSResource:
		status, err = aws.GetRDSStatus(c.RdsClient, resource.Identifier)
	case types.ELBResource:
		status, err = aws.GetELBStatus(c.ElbClient, resource.Identifier)
	case types.APIGatewayResource:
		status, err = aws.GetAPIGatewayStatus(c.ApiGatewayClient, resource.Identifier)
	case types.CloudflarePagesResource:
		status, err = cloudflare.GetPagesStatus(c.CloudflareClient, resource.Identifier)
	default:
		return nil, fmt.Errorf("invalid resource type encountered: %s", resource.Type)
	}

	if err != nil {
		return nil, err
	}

	return status, nil
}
