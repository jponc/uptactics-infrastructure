package vpc

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func CreateInfrastructure(ctx *pulumi.Context) (pulumi.StringInput, []pulumi.StringInput, []pulumi.StringInput, error) {
	conf := config.New(ctx, "")
	vpcName := conf.Require("vpcName")
	vpcCidr := conf.Require("vpcCidr")

	// Creates the VPC
	vpc, err := ec2.NewVpc(ctx, vpcName, &ec2.VpcArgs{
		CidrBlock:          pulumi.String(vpcCidr),
		EnableDnsHostnames: pulumi.Bool(true),
		EnableDnsSupport:   pulumi.Bool(true),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(vpcName),
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Create IGW
	igwName := conf.Require("igwName")

	igw, err := ec2.NewInternetGateway(ctx, igwName, &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(igwName),
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// Create Subnets
	clusterName := conf.Require("clusterName")
	subnetNamesString := conf.Require("subnetNames")
	subnetCidrsString := conf.Require("subnetCidrs")
	subnetAZsString := conf.Require("subnetAZs")

	subnetNames := strings.Split(subnetNamesString, ",")
	subnetCidrs := strings.Split(subnetCidrsString, ",")
	subnetAZs := strings.Split(subnetAZsString, ",")

	publicSubnetIds := []pulumi.StringInput{}
	privateSubnetIds := []pulumi.StringInput{}

	for i, subnetName := range subnetNames {
		subnetCidr := subnetCidrs[i]
		subnetAZ := subnetAZs[i]
		isPrivate := strings.Contains(subnetName, "private")

		tags := pulumi.StringMap{
			"Name": pulumi.String(subnetName),
			fmt.Sprintf("kubernetes.io/cluster/%s", clusterName): pulumi.String("owned"),
		}

		if isPrivate {
			tags["kubernetes.io/role/internal-elb"] = pulumi.String("1")
		} else {
			tags["kubernetes.io/role/elb"] = pulumi.String("1")
		}

		subnet, err := ec2.NewSubnet(ctx, subnetName, &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String(subnetCidr),
			AvailabilityZone: pulumi.String(subnetAZ),
			Tags:             tags,
		})
		if err != nil {
			return nil, nil, nil, err
		}

		if isPrivate {
			privateSubnetIds = append(privateSubnetIds, subnet.ID())
		} else {
			publicSubnetIds = append(publicSubnetIds, subnet.ID())
		}
	}

	// Create NAT Gateway
	natGwName := conf.Require("natGwName")

	eip, err := ec2.NewEip(ctx, natGwName, &ec2.EipArgs{
		Vpc: pulumi.Bool(true),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(natGwName),
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	ngw, err := ec2.NewNatGateway(ctx, natGwName, &ec2.NatGatewayArgs{
		AllocationId: eip.ID(),
		SubnetId:     publicSubnetIds[0],
		Tags: pulumi.StringMap{
			"Name": pulumi.String(natGwName),
		},
	}, pulumi.DependsOn([]pulumi.Resource{igw}))
	if err != nil {
		return nil, nil, nil, err
	}

	// Route Tables
	privateRTName := conf.Require("privateRTName")
	privateRT, err := ec2.NewRouteTable(ctx, privateRTName, &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Routes: ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock:    pulumi.String("0.0.0.0/0"),
				NatGatewayId: ngw.ID(),
			},
		},
		Tags: pulumi.StringMap{
			"Name": pulumi.String(privateRTName),
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	publicRTName := conf.Require("publicRTName")
	publicRT, err := ec2.NewRouteTable(ctx, publicRTName, &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Routes: ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock: pulumi.String("0.0.0.0/0"),
				GatewayId: igw.ID(),
			},
		},
		Tags: pulumi.StringMap{
			"Name": pulumi.String(publicRTName),
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	for i, privateSubnetId := range privateSubnetIds {
		rtaName := fmt.Sprintf("u-staging-rta-private-%d", i+1)
		_, err := ec2.NewRouteTableAssociation(ctx, rtaName, &ec2.RouteTableAssociationArgs{
			SubnetId:     privateSubnetId,
			RouteTableId: privateRT.ID(),
		})
		if err != nil {
			return nil, nil, nil, err
		}
	}

	for i, publicSubnetId := range publicSubnetIds {
		rtaName := fmt.Sprintf("u-staging-rta-public-%d", i+1)
		_, err := ec2.NewRouteTableAssociation(ctx, rtaName, &ec2.RouteTableAssociationArgs{
			SubnetId:     publicSubnetId,
			RouteTableId: publicRT.ID(),
		})
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return vpc.ID(), privateSubnetIds, publicSubnetIds, nil
}
