package aws

import (
	"context"
	"fmt"
	"hermes/app/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
)

var _ types.ResourceStatus = APIGatewayStatus{}

type APIGatewayStatus struct {
	Endpoint string `json:"endpoint"`
	Protocol string `json:"protocol"`
}

func (e APIGatewayStatus) IsResourceStatus() {}

func (e APIGatewayStatus) IsHealthy() bool {
	return true
}

func (e APIGatewayStatus) GetStatusString() string {
	return "active"
}

func GetAPIGatewayStatus(client *apigatewayv2.Client, apiName string) (APIGatewayStatus, error) {
	// TODO: pagination
	apis, err := client.GetApis(context.TODO(), &apigatewayv2.GetApisInput{})

	if err != nil {
		return APIGatewayStatus{}, err
	}

	var apiId string
	for _, api := range apis.Items {
		if *api.Name == apiName {
			apiId = *api.ApiId
			break
		}
	}

	if apiId == "" {
		return APIGatewayStatus{}, fmt.Errorf("no api with name %s found", apiName)
	}

	resp, err := client.GetApi(context.TODO(), &apigatewayv2.GetApiInput{
		ApiId: aws.String(apiId),
	})

	if err != nil {
		return APIGatewayStatus{}, err
	}

	// TODO: maybe more info

	return APIGatewayStatus{
		Endpoint: *resp.ApiEndpoint,
		Protocol: string(resp.ProtocolType),
	}, nil
}

func GetAPIGatewayClient() (*apigatewayv2.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return apigatewayv2.NewFromConfig(cfg), nil
}
