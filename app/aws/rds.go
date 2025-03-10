package aws

import (
	"context"
	"errors"
	"fmt"
	"hermes/app/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
)

var _ types.ResourceStatus = RDSStatus{}

type RDSStatus struct {
	InstanceExists bool   `json:"exists"`
	Status         string `json:"status"`
	InstanceClass  string `json:"instance_class"`
}

func (r RDSStatus) IsResourceStatus() {}

func (r RDSStatus) IsHealthy() bool {
	return r.Status == "available"
}

func (r RDSStatus) Exists() bool {
	return r.InstanceExists
}

func (r RDSStatus) GetStatusString() string {
	return r.Status
}

func GetRDSStatus(client *rds.Client, dbIdentifier string) (RDSStatus, error) {
	resp, err := client.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbIdentifier),
	})

	if err != nil {
		var notFound *rds_types.DBInstanceNotFoundFault
		if errors.As(err, &notFound) {
			return RDSStatus{
				InstanceExists: false,
			}, nil
		}

		return RDSStatus{}, err
	}

	if len(resp.DBInstances) == 0 {
		return RDSStatus{}, fmt.Errorf("no db found")
	}

	firstDb := resp.DBInstances[0]

	return RDSStatus{
		InstanceExists: true,
		Status:         *firstDb.DBInstanceStatus,
		InstanceClass:  *firstDb.DBInstanceClass,
	}, nil
}

func GetRDSClient() (*rds.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return rds.NewFromConfig(cfg), nil
}
