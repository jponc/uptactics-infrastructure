package main

import (
	"uptactics/eks"
	"uptactics/vpc"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vpcId, privateSubnetIds, publicSubnetIds, err := vpc.CreateInfrastructure(ctx)
		if err != nil {
			return err
		}
		err = eks.CreateInfrastructure(ctx, vpcId, privateSubnetIds, publicSubnetIds)
		if err != nil {
			return err
		}

		// Export the name of the bucket
		// ctx.Export("bucketName", bucket.ID())
		return nil
	})
}
