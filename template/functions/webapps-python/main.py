from crossplane.function import resource
from crossplane.function.proto.v1 import run_function_pb2 as fnv1
from .model.io.k8s.api.apps import v1 as appsv1
from .model.io.k8s.api.core import v1 as corev1
from .model.io.k8s.api.networking import v1 as networkingv1
from .model.com.example.platform.webapp import v1alpha1 as platformv1alpha1
from .model.io.k8s.apimachinery.pkg.apis.core.meta import v1 as coremetav1
def compose(req: fnv1.RunFunctionRequest, rsp: fnv1.RunFunctionResponse):
    oxr = platformv1alpha1.WebApp(**req.observed.composite.resource)
    ocds = req.observed.resources

    # Create a Status object to collect updates
    status = platformv1alpha1.Status()

    deployment = appsv1.Deployment(
        metadata=coremetav1.ObjectMeta(
            name=oxr.metadata.name,
            namespace=oxr.metadata.namespace,
            labels={
                "app.kubernetes.io/name": oxr.metadata.name
            },
        ),
        spec=appsv1.DeploymentSpec(
            replicas=oxr.spec.parameters.replicas,
            selector=coremetav1.LabelSelector(
                matchLabels={
                    "app.kubernetes.io/name": oxr.metadata.name,
                    "app": oxr.metadata.name
                }
            ),
            template=corev1.PodTemplateSpec(
                metadata=coremetav1.ObjectMeta(
                    labels={
                        "app.kubernetes.io/name": oxr.metadata.name,
                        "app": oxr.metadata.name
                    }
                ),
                spec=corev1.PodSpec(
                    serviceAccountName=oxr.spec.parameters.serviceAccount,
                    containers=[
                        corev1.Container(
                            name=oxr.metadata.name,
                            image=oxr.spec.parameters.image,
                            imagePullPolicy="Always",
                            ports=[
                                corev1.ContainerPort(
                                    containerPort=int(oxr.spec.parameters.port)
                                )
                            ],
                            resources=corev1.ResourceRequirements(
                                requests={
                                    "memory": "64Mi",
                                    "cpu": "250m"
                                },
                                limits={
                                    "memory": "1Gi",
                                    "cpu": "1"
                                }
                            )
                        )
                    ],
                    restartPolicy="Always"
                )
            )
        )
    )

    if "deployment" in ocds:
        observed_deployment = appsv1.Deployment(**ocds["deployment"].resource)
        if observed_deployment.status and observed_deployment.status.conditions:
            for condition in observed_deployment.status.conditions:
                if condition.type == "Available" and condition.status == "True":
                    rsp.desired.resources["deployment"].ready = True
                    break

    resource.update(rsp.desired.resources["deployment"], deployment)

    if oxr.spec.parameters.service and oxr.spec.parameters.service.enabled:
        service = corev1.Service(
            metadata=coremetav1.ObjectMeta(
                name=oxr.metadata.name,
                namespace=oxr.metadata.namespace,
            ),
            spec=corev1.ServiceSpec(
                selector={
                    "app": oxr.metadata.name
                },
                ports=[
                    corev1.ServicePort(
                        name="http",
                        protocol="TCP",
                        port=80,
                        targetPort=int(oxr.spec.parameters.port)
                    )
                ]
            )
        )

        if "service" in ocds:
            observed_service = corev1.Service(**ocds["service"].resource)
            if observed_service.spec and observed_service.spec.clusterIP:
                rsp.desired.resources["service"].ready = True
        resource.update(rsp.desired.resources["service"], service)

    if oxr.spec.parameters.ingress and oxr.spec.parameters.ingress.enabled:
        ingress = networkingv1.Ingress(
            metadata=coremetav1.ObjectMeta(
                name=oxr.metadata.name,
                namespace=oxr.metadata.namespace,
                annotations={
                    "kubernetes.io/ingress.class": "alb",
                    "alb.ingress.kubernetes.io/scheme": "internet-facing",
                    "alb.ingress.kubernetes.io/target-type": "ip",
                    "alb.ingress.kubernetes.io/healthcheck-path": "/health",
                    "alb.ingress.kubernetes.io/listen-ports": '[{"HTTP": 80}]',
                    "alb.ingress.kubernetes.io/target-group-attributes": "stickiness.enabled=true,stickiness.lb_cookie.duration_seconds=60"
                }
            ),
            spec=networkingv1.IngressSpec(
                rules=[
                    networkingv1.IngressRule(
                        http=networkingv1.HTTPIngressRuleValue(
                            paths=[
                                networkingv1.HTTPIngressPath(
                                    path="/",
                                    pathType="Prefix",
                                    backend=networkingv1.IngressBackend(
                                        service=networkingv1.IngressServiceBackend(
                                            name=oxr.metadata.name,
                                            port=networkingv1.ServiceBackendPort(
                                                number=80
                                            )
                                        )
                                    )
                                )
                            ]
                        )
                    )
                ]
            )
        )

        if "ingress" in ocds:
            observed_ingress = networkingv1.Ingress(**ocds["ingress"].resource)
            if (observed_ingress.status and
                observed_ingress.status.loadBalancer and
                observed_ingress.status.loadBalancer.ingress and
                len(observed_ingress.status.loadBalancer.ingress) > 0 and
                observed_ingress.status.loadBalancer.ingress[0].hostname):
                rsp.desired.resources["ingress"].ready = True
        resource.update(rsp.desired.resources["ingress"], ingress)

    # Set status with defaults
    if "deployment" in ocds:
        observed_deployment = appsv1.Deployment(**ocds["deployment"].resource)
        if observed_deployment.status and observed_deployment.status.conditions:
            status.deploymentConditions = []
            for condition in observed_deployment.status.conditions:
                condition_dict = condition.model_dump(exclude_none=True)
                # Convert datetime objects to ISO format strings
                if 'lastTransitionTime' in condition_dict and condition_dict['lastTransitionTime']:
                    condition_dict['lastTransitionTime'] = condition_dict['lastTransitionTime'].isoformat()
                if 'lastUpdateTime' in condition_dict and condition_dict['lastUpdateTime']:
                    condition_dict['lastUpdateTime'] = condition_dict['lastUpdateTime'].isoformat()
                status.deploymentConditions.append(condition_dict)
        else:
            status.deploymentConditions = []
        status.availableReplicas = observed_deployment.status.availableReplicas if observed_deployment.status and observed_deployment.status.availableReplicas else 0
    else:
        status.deploymentConditions = []
        status.availableReplicas = 0

    if "ingress" in ocds:
        observed_ingress = networkingv1.Ingress(**ocds["ingress"].resource)
        status.url = (
            observed_ingress.status.loadBalancer.ingress[0].hostname
            if (observed_ingress.status and
                observed_ingress.status.loadBalancer and
                observed_ingress.status.loadBalancer.ingress and
                len(observed_ingress.status.loadBalancer.ingress) > 0 and
                observed_ingress.status.loadBalancer.ingress[0].hostname)
            else ""
        )
    else:
        status.url = ""

    resource.update(rsp.desired.composite, {"status": status.model_dump(exclude_none=True)})
