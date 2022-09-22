package eks

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func CreateInfrastructure(ctx *pulumi.Context, vpcId pulumi.StringInput, privateSubnetIds []pulumi.StringInput, publicSubnetIds []pulumi.StringInput) error {
	conf := config.New(ctx, "")
	clusterName := conf.Require("clusterName")
	clusterRole := conf.Require("clusterRole")
	clusterVersion := conf.Require("clusterVersion")

	// Create EKS Role
	eksRole, err := iam.NewRole(ctx, clusterRole, &iam.RoleArgs{
		Name: pulumi.String(clusterRole),
		AssumeRolePolicy: pulumi.String(`{
		    "Version": "2012-10-17",
		    "Statement": [{
		        "Effect": "Allow",
		        "Principal": {
		            "Service": "eks.amazonaws.com"
		        },
		        "Action": "sts:AssumeRole"
		    }]
		}`),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(clusterRole),
		},
	})
	if err != nil {
		return err
	}

	// Create EKS Policy Attachments
	eksPolicies := []string{
		"arn:aws:iam::aws:policy/AmazonEKSServicePolicy",
		"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
		"arn:aws:iam::aws:policy/AmazonEKSVPCResourceController",
	}

	for i, eksPolicy := range eksPolicies {
		attachmentName := fmt.Sprintf("%s-rpa-%d", clusterRole, i+1)
		_, err := iam.NewRolePolicyAttachment(ctx, attachmentName, &iam.RolePolicyAttachmentArgs{
			PolicyArn: pulumi.String(eksPolicy),
			Role:      eksRole.Name,
		})
		if err != nil {
			return err
		}
	}

	// Create a Security Group that we can use to actually connect to our cluster
	additionalSg, err := ec2.NewSecurityGroup(ctx, "cluster-sg", &ec2.SecurityGroupArgs{
		VpcId: vpcId,
		Egress: ec2.SecurityGroupEgressArray{
			ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
	})
	if err != nil {
		return err
	}

	// Create EKS Control Plane
	eksCluster, err := eks.NewCluster(ctx, clusterName, &eks.ClusterArgs{
		Name:    pulumi.String(clusterName),
		RoleArn: pulumi.StringInput(eksRole.Arn),
		Version: pulumi.String(clusterVersion),

		VpcConfig: &eks.ClusterVpcConfigArgs{
			EndpointPrivateAccess: pulumi.Bool(false),
			EndpointPublicAccess:  pulumi.Bool(true),
			PublicAccessCidrs: pulumi.StringArray{
				pulumi.String("0.0.0.0/0"),
			},
			SecurityGroupIds: pulumi.StringArray{
				additionalSg.ID().ToStringOutput(),
			},
			SubnetIds: pulumi.StringArray(append(privateSubnetIds, publicSubnetIds...)),
		},

		Tags: pulumi.StringMap{
			"Name": pulumi.String(clusterName),
		},
	})
	if err != nil {
		return err
	}

	// Export kubeconfig
	ctx.Export("kubeconfig", generateKubeconfig(eksCluster.Endpoint,
		eksCluster.CertificateAuthority.Data().Elem(), eksCluster.Name))

	// Create Fargate Profile Role
	fargateRoleName := conf.Require("fargateRoleName")
	fargateRole, err := iam.NewRole(ctx, fargateRoleName, &iam.RoleArgs{
		Name: pulumi.String(fargateRoleName),
		AssumeRolePolicy: pulumi.String(`{
		    "Version": "2012-10-17",
		    "Statement": [{
		        "Effect": "Allow",
		        "Principal": {
		            "Service": "eks-fargate-pods.amazonaws.com"
		        },
		        "Action": "sts:AssumeRole"
		    }]
		}`),
		Tags: pulumi.StringMap{
			"Name": pulumi.String(fargateRoleName),
		},
	})
	if err != nil {
		return err
	}

	// Create Farget Policy Attachment
	attachmentName := fmt.Sprintf("%s-rpa", fargateRoleName)
	_, err = iam.NewRolePolicyAttachment(ctx, attachmentName, &iam.RolePolicyAttachmentArgs{
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonEKSFargatePodExecutionRolePolicy"),
		Role:      fargateRole.Name,
	})
	if err != nil {
		return err
	}

	// Create AWS Fargate Profile

	fargateProfileName := conf.Require("fargateProfileName")
	_, err = eks.NewFargateProfile(ctx, fargateProfileName, &eks.FargateProfileArgs{
		ClusterName:         pulumi.String(clusterName),
		FargateProfileName:  pulumi.String(fargateProfileName),
		PodExecutionRoleArn: pulumi.StringInput(fargateRole.Arn),
		SubnetIds:           pulumi.StringArray(privateSubnetIds),
		Selectors: eks.FargateProfileSelectorArray{
			eks.FargateProfileSelectorArgs{
				Namespace: pulumi.String("kube-system"),
			},
			eks.FargateProfileSelectorArgs{
				Namespace: pulumi.String("default"),
			},
		},
		Tags: pulumi.StringMap{
			"Name": pulumi.String(fargateProfileName),
		},
	}, pulumi.DependsOn([]pulumi.Resource{eksCluster}))
	if err != nil {
		return err
	}

	return nil
}

// Create the KubeConfig Structure as per https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
func generateKubeconfig(clusterEndpoint pulumi.StringOutput, certData pulumi.StringOutput, clusterName pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.Sprintf(`{
        "apiVersion": "v1",
        "clusters": [{
            "cluster": {
                "server": "%s",
                "certificate-authority-data": "%s"
            },
            "name": "kubernetes",
        }],
        "contexts": [{
            "context": {
                "cluster": "kubernetes",
                "user": "aws",
            },
            "name": "aws",
        }],
        "current-context": "aws",
        "kind": "Config",
        "users": [{
            "name": "aws",
            "user": {
                "exec": {
                    "apiVersion": "client.authentication.k8s.io/v1alpha1",
                    "command": "aws-iam-authenticator",
                    "args": [
                        "token",
                        "-i",
                        "%s",
                    ],
                },
            },
        }],
    }`, clusterEndpoint, certData, clusterName)
}
