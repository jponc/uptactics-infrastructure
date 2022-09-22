package certmanager

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateCertManager(ctx *pulumi.Context, eksCluster *eks.Cluster) error {
	// Create CertManager from Yaml
	_, err := yaml.NewConfigFile(ctx, "certmanager", &yaml.ConfigFileArgs{
		File:      "certmanager/cert-manager.yaml",
		SkipAwait: false,
	}, pulumi.DependsOn([]pulumi.Resource{eksCluster}))
	if err != nil {
		return err
	}

	return nil
}
