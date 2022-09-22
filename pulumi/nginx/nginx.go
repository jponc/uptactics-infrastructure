package nginx

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateNginx(ctx *pulumi.Context) error {
	// Create Nginx Deployment
	_, err := appsv1.NewDeployment(ctx, "nginx-deployment", &appsv1.DeploymentArgs{
		Kind:       pulumi.String("Deployment"),
		ApiVersion: pulumi.String("apps/v1"),
		Metadata: metav1.ObjectMetaArgs{
			Name: pulumi.String("my-nginx"),
		},
		Spec: appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"run": pulumi.String("my-nginx"),
				},
			},
			Template: corev1.PodTemplateSpecArgs{
				Metadata: metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"run": pulumi.String("my-nginx"),
					},
				},
				Spec: corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						corev1.ContainerArgs{
							Image: pulumi.String("nginx"),
							Name:  pulumi.String("my-nginx"),
							Ports: corev1.ContainerPortArray{
								corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(80),
								},
							},
						},
					},
				},
			},
		},
	})

	// Create Nginx Service
	_, err = corev1.NewService(ctx, "nginx-svc", &corev1.ServiceArgs{
		Kind:       pulumi.String("Service"),
		ApiVersion: pulumi.String("v1"),
		Metadata: metav1.ObjectMetaArgs{
			Name: pulumi.String("my-nginx"),
			Labels: pulumi.StringMap{
				"run": pulumi.String("my-nginx"),
			},
		},
		Spec: corev1.ServiceSpecArgs{
			Type: pulumi.String("LoadBalancer"),
			Ports: corev1.ServicePortArray{
				corev1.ServicePortArgs{
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(80),
					Protocol:   pulumi.String("TCP"),
				},
			},
			Selector: pulumi.StringMap{
				"run": pulumi.String("my-nginx"),
			},
		},
	})

	if err != nil {
		return err
	}

	return nil
}
