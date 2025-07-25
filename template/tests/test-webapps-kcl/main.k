import models.com.example.platform.v1alpha1 as platformv1alpha1
import models.io.upbound.dev.meta.v1alpha1 as metav1alpha1
import models.io.k8s.api.apps.v1 as appsv1
import models.io.k8s.api.core.v1 as corev1

_items = [
    metav1alpha1.CompositionTest{
        metadata.name: "test-webapps-kcl"
        spec= {
            assertResources: [
                appsv1.Deployment{
                    metadata: {
                        name: "webservice"
                        namespace: "default"
                        annotations: {
                            "crossplane.io/composition-resource-name": "deployment"
                        }
                    }
                    spec: {
                        replicas: 1
                        selector: {
                            matchLabels: {
                                "app.kubernetes.io/name": "webservice"
                                app: "webservice"
                            }
                        }
                        template: {
                            metadata: {
                                labels: {
                                    "app.kubernetes.io/name": "webservice"
                                    app: "webservice"
                                }
                            }
                            spec: {
                                serviceAccountName: "default"
                                containers: [{
                                    name: "webservice"
                                    image: "nginx"
                                    imagePullPolicy: "Always"
                                    ports: [{
                                        name: "http"
                                        containerPort: 8080
                                        protocol: "TCP"
                                    }]
                                    resources: {
                                        requests: {
                                            memory: "64Mi"
                                            cpu: "250m"
                                        }
                                        limits: {
                                            memory: "1Gi"
                                            cpu: "1"
                                        }
                                    }
                                }]
                                restartPolicy: "Always"
                            }
                        }
                    }
                }

                platformv1alpha1.WebApp{
                    apiVersion: "platform.example.com/v1alpha1"
                    kind: "WebApp"
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
                    status: {
                        availableReplicas: 1
                        deploymentConditions: [
                            {
                                message: "Deployment has minimum availability."
                                reason: "MinimumReplicasAvailable"
                                status: "True"
                                type: "Available"
                            }
                        ]
                        url: ""
                    }
                }

                corev1.Service{
                    metadata:{
                        name: "webservice"
                        namespace: "default"
                    }
                    spec: {
                        selector: {
                            app: "webservice"
                        }
                        ports: [{
                            name: "http"
                            protocol: "TCP"
                            port: 80
                            targetPort: "http"
                        }]
                    }
                }

            ]
            observedResources: [
                appsv1.Deployment{
                    metadata: {
                        name: "webservice"
                        namespace: "default"
                        annotations: {
                            "crossplane.io/composition-resource-name": "deployment"
                        }
                    }
                    spec: {
                        replicas: 1
                        selector: {
                            matchLabels: {
                                "app.kubernetes.io/name": "webservice"
                                app: "webservice"
                            }
                        }
                        template: {
                            metadata: {
                                labels: {
                                    "app.kubernetes.io/name": "webservice"
                                    app: "webservice"
                                }
                            }
                            spec: {
                                serviceAccountName: "default"
                                containers: [{
                                    name: "webservice"
                                    image: "nginx"
                                    imagePullPolicy: "Always"
                                    ports: [{
                                        name: "http"
                                        containerPort: 8080
                                        protocol: "TCP"
                                    }]
                                    resources: {
                                        requests: {
                                            memory: "64Mi"
                                            cpu: "250m"
                                        }
                                        limits: {
                                            memory: "1Gi"
                                            cpu: "1"
                                        }
                                    }
                                }]
                                restartPolicy: "Always"
                            }
                        }
                    }
                    status: {
                        availableReplicas: 1
                        conditions: [
                            {
                                message: "Deployment has minimum availability."
                                reason: "MinimumReplicasAvailable"
                                status: "True"
                                type: "Available"
                            }
                        ]
                    }
                }

            ]
            compositionPath: "apis/webapps/composition.yaml"
            xrPath: "examples/webapps/example.yaml"
            xrdPath: "apis/webapps/definition.yaml"
            timeoutSeconds: 60
            validate: False
        }
    }
]
items= _items
