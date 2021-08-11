package main

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create `petstore-deploy`
		appLabels := pulumi.ToStringMap(map[string]string{
			"app":     "petstore3",
			"version": "v1.0.6",
		})
		_, err := appsv1.NewDeployment(ctx, "petstore-deploy", &appsv1.DeploymentArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("petstore-deploy"),
			},
			Spec: appsv1.DeploymentSpecArgs{
				Replicas: pulumi.Int(1),
				Selector: metav1.LabelSelectorArgs{
					MatchLabels: appLabels,
				},
				Template: corev1.PodTemplateSpecArgs{
					Metadata: metav1.ObjectMetaArgs{
						Labels: appLabels,
					},
					Spec: corev1.PodSpecArgs{
						Containers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name:  pulumi.String("petstore3"),
								Image: pulumi.String("swaggerapi/petstore3:1.0.6"),
								LivenessProbe: corev1.ProbeArgs{
									HttpGet: corev1.HTTPGetActionArgs{
										Path: pulumi.String("/"),
										Port: pulumi.Int(8080),
									},
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
	})
}
