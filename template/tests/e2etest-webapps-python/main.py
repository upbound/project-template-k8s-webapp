from .model.io.upbound.dev.meta.e2etest import v1alpha1 as e2etest
from .model.io.k8s.apimachinery.pkg.apis.meta import v1 as metav1
from .model.io.k8s.apimachinery.pkg.apis.core.meta import v1 as metacorev1

from .model.io.k8s.api.rbac import v1 as rbacv1
from .model.com.example.platform.webapp import v1alpha1 as platformv1alpha1

test = e2etest.E2ETest(
    metadata=metav1.ObjectMeta(
        name="e2etest-webapps-python",
    ),
    spec = e2etest.Spec(
        initResources=[
            rbacv1.ClusterRoleBinding(
                apiVersion="rbac.authorization.k8s.io/v1",
                kind ="ClusterRoleBinding",
                metadata=metacorev1.ObjectMeta(
                    name="crossplane-clusteradmin"
                ),
                roleRef=rbacv1.RoleRef(
                    apiGroup="rbac.authorization.k8s.io",
                    kind="ClusterRole",
                    name="cluster-admin"
                ),
                subjects=[
                    rbacv1.Subject(
                        kind="ServiceAccount",
                        name="crossplane",
                        namespace="crossplane-system"
                    )
                ]
            ).model_dump(exclude_unset=True)
        ],
        crossplane=e2etest.Crossplane(
            autoUpgrade=e2etest.AutoUpgrade(
                channel="Rapid",
            ),
        ),
        defaultConditions=[
            "Ready",
        ],
        manifests=[
            platformv1alpha1.WebApp(
                apiVersion="platform.example.com/v1alpha1",
                kind="WebApp",
                metadata=metav1.ObjectMeta(
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
                )
            ).model_dump(exclude_unset=True)
        ],
        extraResources=[],
        skipDelete=False,
        timeoutSeconds=4500,
    )
)
