import models.com.example.platform.v1alpha1 as platformv1alpha1
import models.io.upbound.dev.meta.v1alpha1 as metav1alpha1
import models.io.k8s.api.rbac.v1 as rbacv1

_items = [
    metav1alpha1.E2ETest{
        metadata.name: "e2etest-webapps-kcl"
        spec= {
            initResources: [
                rbacv1.ClusterRoleBinding{
                    metadata.name: "crossplane-clusteradmin"
                    roleRef: {
                        apiGroup: "rbac.authorization.k8s.io"
                        kind: "ClusterRole"
                        name: "cluster-admin"
                    }
                    subjects: [
                        {
                            kind: "ServiceAccount"
                            name: "crossplane"
                            namespace: "crospslane-system"
                        }
                    ]
                }
            ]
            crossplane.autoUpgrade.channel: "Rapid"
            defaultConditions: ["Ready"]
            manifests: [
                platformv1alpha1.WebApp{
                    metadata: {
                        name: "webservice"
                        namespace: "default"
                    }
                    spec: {
                        parameters: {
                            image: "nginx"
                            port: 8080
                            replicas: 1
                            service: {
                                enabled: True
                            }
                            ingress: {
                                enabled: False
                            }
                            serviceAccount: "default"
                        }
                    }
                }
            ]
            extraResources: []
            skipDelete: False
            timeoutSeconds: 4500
        }
    }
]
items= _items