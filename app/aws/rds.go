package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

type RDSStatus struct {
	Status        string `json:"status"`
	InstanceClass string `json:"instance_class"`
}

func (r RDSStatus) IsResourceStatus() {}

func GetRDSStatus(client *rds.Client, dbIdentifier string) (RDSStatus, error) {
	resp, err := client.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbIdentifier),
	})

	if err != nil {
		return RDSStatus{}, err
	}

	if len(resp.DBInstances) == 0 {
		return RDSStatus{}, fmt.Errorf("no db found")
	}

	firstDb := resp.DBInstances[0]

	return RDSStatus{
		Status:        *firstDb.DBInstanceStatus,
		InstanceClass: *firstDb.DBInstanceClass,
	}, nil
}

func GetRDSClient() (*rds.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	return rds.NewFromConfig(cfg), nil
}
