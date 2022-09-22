package traefik

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateTraefikIngress(ctx *pulumi.Context) error {
	// Create Traefik Namespace
	_, err := corev1.NewNamespace(ctx, "traefik-namespace", &corev1.NamespaceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Namespace"),
		Metadata: metav1.ObjectMetaArgs{
			Name: pulumi.String("traefik"),
		},
	})
	if err != nil {
		return err
	}

	traefikName := "traefik-ingress-controller"

	// Create ClusterRole
	_, err = rbacv1.NewClusterRole(ctx, traefikName+"-cluster-role", &rbacv1.ClusterRoleArgs{
		Kind:       pulumi.String("ClusterRole"),
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Metadata: metav1.ObjectMetaArgs{
			Namespace: pulumi.String("traefik"),
			Name:      pulumi.String(traefikName),
		},
		Rules: rbacv1.PolicyRuleArray{
			rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String(""),
				},
				Resources: pulumi.StringArray{
					pulumi.String("services"),
					pulumi.String("endpoints"),
					pulumi.String("secrets"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("get"),
					pulumi.String("list"),
					pulumi.String("watch"),
				},
			},
			rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String("extensions"),
					pulumi.String("networking.k8s.io"),
				},
				Resources: pulumi.StringArray{
					pulumi.String("ingresses"),
					pulumi.String("ingressclasses"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("get"),
					pulumi.String("list"),
					pulumi.String("watch"),
				},
			},
			rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String("extensions"),
				},
				Resources: pulumi.StringArray{
					pulumi.String("ingresses/status"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("update"),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	// Create ClusterRoleBinding
	_, err = rbacv1.NewClusterRoleBinding(ctx, traefikName+"-cluster-role-binding", &rbacv1.ClusterRoleBindingArgs{
		Kind:       pulumi.String("ClusterRoleBinding"),
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Metadata: metav1.ObjectMetaArgs{
			Namespace: pulumi.String("traefik"),
			Name:      pulumi.String(traefikName),
		},
		RoleRef: rbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("ClusterRole"),
			Name:     pulumi.String(traefikName),
		},
		Subjects: rbacv1.SubjectArray{
			rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Namespace: pulumi.String("traefik"),
				Name:      pulumi.String(traefikName),
			},
		},
	})
	if err != nil {
		return err
	}

	// Create ServiceAccount
	_, err = corev1.NewServiceAccount(ctx, traefikName+"-service-account", &corev1.ServiceAccountArgs{
		Kind:       pulumi.String("ServiceAccount"),
		ApiVersion: pulumi.String("v1"),
		Metadata: metav1.ObjectMetaArgs{
			Namespace: pulumi.String("traefik"),
			Name:      pulumi.String(traefikName),
		},
	})
	if err != nil {
		return err
	}

	// Create Deployment
	_, err = appsv1.NewDeployment(ctx, traefikName+"-deployment", &appsv1.DeploymentArgs{
		Kind:       pulumi.String("Deployment"),
		ApiVersion: pulumi.String("apps/v1"),
		Metadata: metav1.ObjectMetaArgs{
			Namespace: pulumi.String("traefik"),
			Name:      pulumi.String(traefikName),
			Labels: pulumi.StringMap{
				"k8s-app": pulumi.String("traefik-ingress-lb"),
			},
		},
		Spec: appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"k8s-app": pulumi.String("traefik-ingress-lb"),
				},
			},
			Template: corev1.PodTemplateSpecArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name: pulumi.String("traefik-ingress-lb"),
					Labels: pulumi.StringMap{
						"k8s-app": pulumi.String("traefik-ingress-lb"),
					},
				},
				Spec: corev1.PodSpecArgs{
					ServiceAccountName:            pulumi.String(traefikName),
					TerminationGracePeriodSeconds: pulumi.Int(60),
					Containers: corev1.ContainerArray{
						corev1.ContainerArgs{
							Image: pulumi.String("traefik:v2.8"),
							Name:  pulumi.String("traefik-ingress-lb"),
							Ports: corev1.ContainerPortArray{
								corev1.ContainerPortArgs{
									Name:          pulumi.String("http"),
									ContainerPort: pulumi.Int(80),
								},
								corev1.ContainerPortArgs{
									Name:          pulumi.String("https"),
									ContainerPort: pulumi.Int(443),
								},
								corev1.ContainerPortArgs{
									Name:          pulumi.String("admin"),
									ContainerPort: pulumi.Int(8080),
								},
							},
							Args: pulumi.StringArray{
								pulumi.String("--api"),
								pulumi.String("--providers.kubernetesingress=true"),
								pulumi.String("--log.level=DEBUG"),
								pulumi.String("--entrypoints.http.address=:80"),
								pulumi.String("--entrypoints.https.address=:443"),
								pulumi.String("--entrypoints.https.http.tls=true"),
								pulumi.String("--entrypoints.https.http.redirections.entrypoint.to=https"),
								pulumi.String("--entrypoints.https.http.redirections.entrypoint.scheme=https"),
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
