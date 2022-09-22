package kubernetes

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func InitializeKubernetesProvider(ctx *pulumi.Context) (*kubernetes.Provider, error) {
	conf := config.New(ctx, "")
	contextName := conf.Require("kubernetesContext")

	provider, err := kubernetes.NewProvider(ctx, contextName, &kubernetes.ProviderArgs{
		Context: pulumi.String(contextName),
	})
	if err != nil {
		return nil, err
	}

  return provider, nil
}
