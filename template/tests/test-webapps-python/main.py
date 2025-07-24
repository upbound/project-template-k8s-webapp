from .model.io.upbound.dev.meta.compositiontest import v1alpha1 as compositiontest
from .model.io.k8s.apimachinery.pkg.apis.meta import v1
from .model.io.k8s.api.apps import v1 as appsv1
from .model.io.k8s.api.core import v1 as corev1
from .model.com.example.platform.webapp import v1alpha1 as platformv1alpha1

deployment = appsv1.Deployment(
    apiVersion = 'apps/v1',
    kind = 'Deployment',
    metadata=v1.ObjectMeta(
        name="webservice",
        namespace="default",
        annotations={
            "crossplane.io/composition-resource-name": "deployment"
        }
    ),
    spec=appsv1.DeploymentSpec(
        replicas=1,
        selector=v1.LabelSelector(
            matchLabels={
                "app.kubernetes.io/name": "webservice",
                "app": "webservice"
            }
        ),
        template=corev1.PodTemplateSpec(
            metadata=v1.ObjectMeta(
                labels={
                    "app.kubernetes.io/name": "webservice",
                    "app": "webservice"
                }
            ),
            spec=corev1.PodSpec(
                serviceAccountName="default",
                containers=[
                    corev1.Container(
                        name="webservice",
                        image="nginx",
                        imagePullPolicy="Always",
                        ports=[
                            corev1.ContainerPort(
                                containerPort=8080
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

webapp = platformv1alpha1.WebApp(
    apiVersion="platform.example.com/v1alpha1",
    kind="WebApp",
    metadata=v1.ObjectMeta(
        name="webservice",
        namespace="default"
    ),
    spec=platformv1alpha1.Spec(
        parameters=platformv1alpha1.Parameters(
            image="nginx",
            port=8080,
            replicas=1,
            service=platformv1alpha1.Service(
                enabled=True
            ),
            ingress=platformv1alpha1.Ingress(
                enabled=False
            ),
            serviceAccount="default"
        )
    ),
    status=platformv1alpha1.Status(
        availableReplicas=1,
        deploymentConditions=[
            platformv1alpha1.DeploymentCondition(
                message="Deployment has minimum availability.",
                reason="MinimumReplicasAvailable",
                status="True",
                type="Available"
            )
        ],
        url=""
    )
)

service = corev1.Service(
    apiVersion = 'v1',
    kind = 'Service',
    metadata=v1.ObjectMeta(
        name="webservice",
        namespace="default"
    ),
    spec=corev1.ServiceSpec(
        selector={
            "app": "webservice"
        },
        ports=[
            corev1.ServicePort(
                name="http",
                port=80,
                targetPort=8080
            )
        ]
    )
)

observed_deployment = appsv1.Deployment(
    apiVersion = 'apps/v1',
    kind = 'Deployment',
    metadata=v1.ObjectMeta(
        name="webservice",
        namespace="default",
        annotations={
            "crossplane.io/composition-resource-name": "deployment"
        }
    ),
    spec=appsv1.DeploymentSpec(
        replicas=1,
        selector=v1.LabelSelector(
            matchLabels={
                "app.kubernetes.io/name": "webservice",
                "app": "webservice"
            }
        ),
        template=corev1.PodTemplateSpec(
            metadata=v1.ObjectMeta(
                labels={
                    "app.kubernetes.io/name": "webservice",
                    "app": "webservice"
                }
            ),
            spec=corev1.PodSpec(
                serviceAccountName="default",
                containers=[
                    corev1.Container(
                        name="webservice",
                        image="nginx",
                        imagePullPolicy="Always",
                        ports=[
                            corev1.ContainerPort(
                                containerPort=8080
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
    ),
    status=appsv1.DeploymentStatus(
        availableReplicas=1,
        conditions=[
            appsv1.DeploymentCondition(
                message="Deployment has minimum availability.",
                reason="MinimumReplicasAvailable",
                status="True",
                type="Available"
            )
        ]
    )
)

test = compositiontest.CompositionTest(
    metadata=v1.ObjectMeta(
        name="test-webapps-python",
    ),
    spec = compositiontest.Spec(
        assertResources=[
            deployment.model_dump(exclude_unset=True),
            webapp.model_dump(exclude_unset=True),
            service.model_dump(exclude_unset=True)
        ],
        observedResources=[
            observed_deployment.model_dump(exclude_unset=True)
        ],
        compositionPath="apis/webapps/composition.yaml",
        xrPath="examples/webapps/example.yaml",
        xrdPath="apis/webapps/definition.yaml",
        timeoutSeconds=60,
        validate=False
    )
)