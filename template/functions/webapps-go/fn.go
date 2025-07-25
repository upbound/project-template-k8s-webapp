package main

import (
	"context"
	"encoding/json"

	"dev.upbound.io/models/com/example/platform/v1alpha1"
	appsv1 "dev.upbound.io/models/io/k8s/apps/v1"
	coremetav1 "dev.upbound.io/models/io/k8s/core/meta/v1"
	corev1 "dev.upbound.io/models/io/k8s/core/v1"
	networkingv1 "dev.upbound.io/models/io/k8s/networking/v1"
	resourcev1 "dev.upbound.io/models/io/k8s/resource/v1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/function-sdk-go/errors"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
	"k8s.io/utils/ptr"
)

// Function is your composition function.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())
	rsp := response.To(req, response.DefaultTTL)

	observedComposite, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get xr"))
		return rsp, nil
	}

	observedComposed, err := request.GetObservedComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get observed resources"))
		return rsp, nil
	}

	var xr v1alpha1.WebApp
	if err := convertViaJSON(&xr, observedComposite.Resource); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot convert xr"))
		return rsp, nil
	}

	params := xr.Spec.Parameters
	if params == nil {
		response.Fatal(rsp, errors.New("missing parameters"))
		return rsp, nil
	}

	// We'll collect our desired composed resources into this map, then convert
	// them to the SDK's types and set them in the response when we return.
	desiredComposed := make(map[resource.Name]any)
	defer func() {
		desiredComposedResources, err := request.GetDesiredComposedResources(req)
		if err != nil {
			response.Fatal(rsp, errors.Wrap(err, "cannot get desired resources"))
			return
		}

		for name, obj := range desiredComposed {
			c := composed.New()
			if err := convertViaJSON(c, obj); err != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot convert %s to unstructured", name))
				return
			}
			dc := &resource.DesiredComposed{Resource: c}

			// Check if this resource should be marked as ready
			if c.GetAnnotations()["go.upbound.io/ready"] == "True" {
				dc.Ready = resource.ReadyTrue
			}

			desiredComposedResources[name] = dc
		}

		if err := response.SetDesiredComposedResources(rsp, desiredComposedResources); err != nil {
			response.Fatal(rsp, errors.Wrap(err, "cannot set desired resources"))
			return
		}
	}()

	// Create Deployment
	deployment := &appsv1.Deployment{
		APIVersion: ptr.To(appsv1.DeploymentAPIVersionAppsV1),
		Kind:       ptr.To(appsv1.DeploymentKindDeployment),
		Metadata: &coremetav1.ObjectMeta{
			Name:      xr.Metadata.Name,
			Namespace: xr.Metadata.Namespace,
			Labels: &map[string]string{
				"app.kubernetes.io/name": *xr.Metadata.Name,
			},
		},
		Spec: &appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(*params.Replicas)),
			Selector: &coremetav1.LabelSelector{
				MatchLabels: &map[string]string{
					"app.kubernetes.io/name": *xr.Metadata.Name,
					"app":                    *xr.Metadata.Name,
				},
			},
			// ToDo(haarchri): remove this
			Strategy: &appsv1.IoK8SApiAppsV1DeploymentStrategy{},
			Template: &corev1.PodTemplateSpec{
				Metadata: &coremetav1.ObjectMeta{
					Labels: &map[string]string{
						"app.kubernetes.io/name": *xr.Metadata.Name,
						"app":                    *xr.Metadata.Name,
					},
				},
				Spec: &corev1.PodSpec{
					ServiceAccountName: params.ServiceAccount,
					Containers: &[]corev1.Container{{
						Name:            xr.Metadata.Name,
						Image:           params.Image,
						ImagePullPolicy: ptr.To("Always"),
						Ports: &[]corev1.ContainerPort{{
							Name:          ptr.To("http"),
							ContainerPort: ptr.To(int32(*params.Port)),
							Protocol:      ptr.To("TCP"),
						}},
						Resources: &corev1.ResourceRequirements{
							Requests: &map[string]resourcev1.Quantity{
								"memory": "64Mi",
								"cpu":    "250m",
							},
							Limits: &map[string]resourcev1.Quantity{
								"memory": "1Gi",
								"cpu":    "1",
							},
						},
					}},
					RestartPolicy: ptr.To("Always"),
				},
			},
		},
		// ToDo(haarchri): remove this
		Status: &appsv1.IoK8SApiAppsV1DeploymentStatus{},
	}

	// Check if deployment is ready
	observedDeployment, ok := observedComposed["deployment"]
	if ok && observedDeployment.Resource != nil {
		var obsDeployment appsv1.Deployment
		if err := convertViaJSON(&obsDeployment, observedDeployment.Resource); err == nil {
			if obsDeployment.Status != nil && obsDeployment.Status.Conditions != nil {
				for _, c := range *obsDeployment.Status.Conditions {
					if c.Type != nil && *c.Type == "Available" &&
						c.Status != nil && *c.Status == "True" {
						if deployment.Metadata.Annotations == nil {
							deployment.Metadata.Annotations = &map[string]string{}
						}
						(*deployment.Metadata.Annotations)["go.upbound.io/ready"] = "True"
						break
					}
				}
			}
		}
	}
	desiredComposed["deployment"] = deployment

	// Create Service if enabled
	if params.Service != nil && params.Service.Enabled != nil && *params.Service.Enabled {
		service := &corev1.Service{
			APIVersion: ptr.To(corev1.ServiceAPIVersionV1),
			Kind:       ptr.To(corev1.ServiceKindService),
			Metadata: &coremetav1.ObjectMeta{
				Name:      xr.Metadata.Name,
				Namespace: xr.Metadata.Namespace,
			},
			Spec: &corev1.ServiceSpec{
				Selector: &map[string]string{
					"app": *xr.Metadata.Name,
				},
				Ports: &[]corev1.ServicePort{{
					Name:       ptr.To("http"),
					Protocol:   ptr.To("TCP"),
					Port:       ptr.To(int32(80)),
					TargetPort: ptr.To("http"),
				}},
			},
			// ToDo(haarchri): remove this
			Status: &corev1.ServiceStatus{
				LoadBalancer: &corev1.LoadBalancerStatus{},
			},
		}

		// Check if service is ready
		observedService, ok := observedComposed["service"]
		if ok && observedService.Resource != nil {
			var obsService corev1.Service
			if err := convertViaJSON(&obsService, observedService.Resource); err == nil {
				if obsService.Spec != nil && obsService.Spec.ClusterIP != nil && *obsService.Spec.ClusterIP != "" {
					if service.Metadata.Annotations == nil {
						service.Metadata.Annotations = &map[string]string{}
					}
					(*service.Metadata.Annotations)["go.upbound.io/ready"] = "True"
				}
			}
		}
		desiredComposed["service"] = service
	}

	// Create Ingress if enabled
	if params.Ingress != nil && params.Ingress.Enabled != nil && *params.Ingress.Enabled {
		ingress := &networkingv1.Ingress{
			APIVersion: ptr.To(networkingv1.IngressAPIVersionNetworkingK8SIoV1),
			Kind:       ptr.To(networkingv1.IngressKindIngress),
			Metadata: &coremetav1.ObjectMeta{
				Name:      xr.Metadata.Name,
				Namespace: xr.Metadata.Namespace,
				Annotations: &map[string]string{
					"kubernetes.io/ingress.class":                       "alb",
					"alb.ingress.kubernetes.io/scheme":                  "internet-facing",
					"alb.ingress.kubernetes.io/target-type":             "ip",
					"alb.ingress.kubernetes.io/healthcheck-path":        "/health",
					"alb.ingress.kubernetes.io/listen-ports":            `[{"HTTP": 80}]`,
					"alb.ingress.kubernetes.io/target-group-attributes": "stickiness.enabled=true,stickiness.lb_cookie.duration_seconds=60",
				},
			},
			Spec: &networkingv1.IngressSpec{
				Rules: &[]networkingv1.IngressRule{{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: &[]networkingv1.HTTPIngressPath{{
							Path:     ptr.To("/"),
							PathType: ptr.To("Prefix"),
							Backend: &networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: xr.Metadata.Name,
									Port: &networkingv1.ServiceBackendPort{
										Number: ptr.To(int32(80)),
									},
								},
							},
						}},
					},
				}},
			},
		}

		// Check if ingress is ready
		observedIngress, ok := observedComposed["ingress"]
		if ok && observedIngress.Resource != nil {
			var obsIngress networkingv1.Ingress
			if err := convertViaJSON(&obsIngress, observedIngress.Resource); err == nil {
				if obsIngress.Status != nil && obsIngress.Status.LoadBalancer != nil &&
					obsIngress.Status.LoadBalancer.Ingress != nil && len(*obsIngress.Status.LoadBalancer.Ingress) > 0 {
					firstIngress := (*obsIngress.Status.LoadBalancer.Ingress)[0]
					if firstIngress.Hostname != nil && *firstIngress.Hostname != "" {
						if ingress.Metadata.Annotations == nil {
							ingress.Metadata.Annotations = &map[string]string{}
						}
						(*ingress.Metadata.Annotations)["go.upbound.io/ready"] = "True"
					}
				}
			}
		}
		desiredComposed["ingress"] = ingress
	}

	// Update XR status
	desiredXR, err := request.GetDesiredCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get desired composite resource"))
		return rsp, nil
	}

	// Convert desired XR to WebApp
	var desiredWebApp v1alpha1.WebApp
	desiredWebApp.APIVersion = ptr.To(v1alpha1.WebAppAPIVersionplatformExampleComV1Alpha1)
	desiredWebApp.Kind = ptr.To(v1alpha1.WebAppKindWebApp)
	if err := convertViaJSON(&desiredWebApp, desiredXR.Resource); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot convert desired xr"))
		return rsp, nil
	}

	// Update status fields
	if desiredWebApp.Status == nil {
		desiredWebApp.Status = &v1alpha1.WebAppStatus{}
	}

	// Set deployment conditions
	if observedDeployment, ok := observedComposed["deployment"]; ok && observedDeployment.Resource != nil {
		var obsDeployment appsv1.Deployment
		if err := convertViaJSON(&obsDeployment, observedDeployment.Resource); err == nil {
			if obsDeployment.Status != nil {
				if obsDeployment.Status.Conditions != nil {
					deploymentConditions := []v1alpha1.WebAppStatusDeploymentConditionsItem{}
					for _, c := range *obsDeployment.Status.Conditions {
						condition := v1alpha1.WebAppStatusDeploymentConditionsItem{
							Type:    c.Type,
							Status:  c.Status,
							Message: c.Message,
							Reason:  c.Reason,
						}
						if c.LastUpdateTime != nil {
							condition.LastUpdateTime = ptr.To(c.LastUpdateTime.String())
						}
						if c.LastTransitionTime != nil {
							condition.LastTransitionTime = ptr.To(c.LastTransitionTime.String())
						}
						deploymentConditions = append(deploymentConditions, condition)
					}
					desiredWebApp.Status.DeploymentConditions = &deploymentConditions
				} else {
					// Set empty conditions if no conditions exist
					deploymentConditions := []v1alpha1.WebAppStatusDeploymentConditionsItem{}
					desiredWebApp.Status.DeploymentConditions = &deploymentConditions
				}
				if obsDeployment.Status.AvailableReplicas != nil {
					desiredWebApp.Status.AvailableReplicas = ptr.To(int(*obsDeployment.Status.AvailableReplicas))
				} else {
					// Set default value when no available replicas
					desiredWebApp.Status.AvailableReplicas = ptr.To(0)
				}
			} else {
				// Set defaults when status is nil
				deploymentConditions := []v1alpha1.WebAppStatusDeploymentConditionsItem{}
				desiredWebApp.Status.DeploymentConditions = &deploymentConditions
				desiredWebApp.Status.AvailableReplicas = ptr.To(0)
			}
		}
	} else {
		// Set defaults when deployment doesn't exist
		deploymentConditions := []v1alpha1.WebAppStatusDeploymentConditionsItem{}
		desiredWebApp.Status.DeploymentConditions = &deploymentConditions
		desiredWebApp.Status.AvailableReplicas = ptr.To(0)
	}

	// Set ingress URL
	if observedIngress, ok := observedComposed["ingress"]; ok && observedIngress.Resource != nil {
		var obsIngress networkingv1.Ingress
		if err := convertViaJSON(&obsIngress, observedIngress.Resource); err == nil {
			if obsIngress.Status != nil && obsIngress.Status.LoadBalancer != nil &&
				obsIngress.Status.LoadBalancer.Ingress != nil && len(*obsIngress.Status.LoadBalancer.Ingress) > 0 {
				firstIngress := (*obsIngress.Status.LoadBalancer.Ingress)[0]
				if firstIngress.Hostname != nil {
					desiredWebApp.Status.URL = firstIngress.Hostname
				} else {
					// Set empty string when hostname is nil
					desiredWebApp.Status.URL = ptr.To("")
				}
			} else {
				// Set empty string when no load balancer ingress
				desiredWebApp.Status.URL = ptr.To("")
			}
		} else {
			// Set empty string when conversion fails
			desiredWebApp.Status.URL = ptr.To("")
		}
	} else {
		// Set empty string when ingress doesn't exist
		desiredWebApp.Status.URL = ptr.To("")
	}

	// Convert back to unstructured
	if err := convertViaJSON(desiredXR.Resource, &desiredWebApp); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot convert desired webapp back to unstructured"))
		return rsp, nil
	}

	if err := response.SetDesiredCompositeResource(rsp, desiredXR); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot set desired composite resource"))
		return rsp, nil
	}

	return rsp, nil
}

func convertViaJSON(to, from any) error {
	bs, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, to)
}
