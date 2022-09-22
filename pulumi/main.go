package main

import (
	"uptactics/certmanager"
	"uptactics/eks"
	"uptactics/traefik"
	"uptactics/vpc"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vpcId, privateSubnetIds, publicSubnetIds, err := vpc.CreateInfrastructure(ctx)
		if err != nil {
			return err
		}

		eksCluster, err := eks.CreateInfrastructure(ctx, vpcId, privateSubnetIds, publicSubnetIds)
		if err != nil {
			return err
		}

		err = traefik.CreateTraefikIngress(ctx, eksCluster)
		if err != nil {
			return err
		}

		err = certmanager.CreateCertManager(ctx, eksCluster)
		if err != nil {
			return err
		}

		return nil
	})
}
