package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"

	"github.com/buger/jsonparser"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	networkingv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/networking/v1"
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

		// Create `petstore-svc`
		petSvc, err := corev1.NewService(ctx, "petstore-svc", &corev1.ServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("petstore-svc"),
			},
			Spec: corev1.ServiceSpecArgs{
				Selector: appLabels,
				Ports: corev1.ServicePortArray{
					corev1.ServicePortArgs{
						Port:       pulumi.Int(80),
						TargetPort: pulumi.Int(8080),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		// Parse `openapi.json`
		oas, err := ioutil.ReadFile("openapi.json")
		if err != nil {
			return err
		}
		prefix, err := jsonparser.GetString(oas, "servers", "[0]", "url")
		if err != nil {
			return err
		}
		pathParamRe := regexp.MustCompile(`\{.+\}`)
		pathParamMap := map[string]string{
			"{petId}":    `(\d+)`,
			"{orderId}":  `(\d+)`,
			"{username}": `([A-Za-z0-9]+)`,
		}
		paths := networkingv1.HTTPIngressPathArray{}
		jsonparser.ObjectEach(oas, func(key, value []byte, dataType jsonparser.ValueType, offset int) error {
			path := fmt.Sprintf("%s%s", prefix, key)
			path = pathParamRe.ReplaceAllStringFunc(path, func(s string) string {
				return pathParamMap[s]
			})
			paths = append(paths, networkingv1.HTTPIngressPathArgs{
				Path:     pulumi.String(path),
				PathType: pulumi.String("Exact"),
				Backend: networkingv1.IngressBackendArgs{
					Service: networkingv1.IngressServiceBackendArgs{
						Name: petSvc.Metadata.Elem().Name().Elem(),
						Port: networkingv1.ServiceBackendPortArgs{
							Number: petSvc.Spec.Elem().Ports().Index(pulumi.Int(0)).Port(),
						},
					},
				},
			})
			return nil
		}, "paths")

		// Get minikube IP
		minikubeIpOutput, err := exec.Command("minikube", "ip").Output()
		if err != nil {
			return err
		}
		minikubeIp := string(minikubeIpOutput)
		minikubeIp = strings.TrimSpace(minikubeIp)

		return nil
	})
}
