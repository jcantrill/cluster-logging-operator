package k8shandler

import (
	"context"

	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/runtime"

	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/migrations"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	securityv1 "github.com/openshift/api/security/v1"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	cli "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconciling", func() {
	defer GinkgoRecover()

	_ = loggingv1.SchemeBuilder.AddToScheme(scheme.Scheme)
	_ = monitoringv1.AddToScheme(scheme.Scheme)
	_ = securityv1.AddToScheme(scheme.Scheme)
	_ = configv1.AddToScheme(scheme.Scheme)

	var (
		cluster = &loggingv1.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "instance",
				Namespace: constants.OpenshiftNS,
			},
			Spec: loggingv1.ClusterLoggingSpec{
				ManagementState: loggingv1.ManagementStateManaged,
				LogStore: &loggingv1.LogStoreSpec{
					Type: loggingv1.LogStoreTypeElasticsearch,
				},
				Collection: &loggingv1.CollectionSpec{
					Type:          loggingv1.LogCollectionTypeFluentd,
					CollectorSpec: loggingv1.CollectorSpec{},
				},
			},
		}

		fluentdSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.CollectorName,
				Namespace: cluster.GetNamespace(),
			},
		}
		fluentdCABundle = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.CollectorTrustedCAName,
				Namespace: cluster.GetNamespace(),
				Labels: map[string]string{
					constants.InjectTrustedCABundleLabel: "true",
				},
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       "ClusterLogging",
						Name:       "instance",
						APIVersion: "logging.openshift.io/v1",
						Controller: utils.GetPtr(true),
					},
				},
			},
			Data: map[string]string{
				constants.TrustedCABundleKey: "",
			},
		}
		// Adding ns and label to account for addSecurityLabelsToNamespace() added in LOG-2620
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"test": "true"},
				Name:   cluster.Namespace,
			},
		}
		extras = map[string]bool{}
	)

	Describe("Collection", func() {
		var (
			client         cli.Client
			clusterRequest *ClusterLoggingRequest
			spec           loggingv1.ClusterLogForwarderSpec
		)

		Context("when cluster proxy present", func() {
			var (
				customCABundle = `
                  -----BEGIN CERTIFICATE-----
                  <PEM_ENCODED_CERT1>
                  -----END CERTIFICATE-------
                  -----BEGIN CERTIFICATE-----
                  <PEM_ENCODED_CERT2>
                  -----END CERTIFICATE-------
                `
				trustedCABundleVolume = corev1.Volume{
					Name: constants.CollectorTrustedCAName,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: constants.CollectorTrustedCAName,
							},
							Items: []corev1.KeyToPath{
								{
									Key:  constants.TrustedCABundleKey,
									Path: constants.TrustedCABundleMountFile,
								},
							},
						},
					},
				}
				trustedCABundleVolumeMount = corev1.VolumeMount{
					Name:      constants.CollectorTrustedCAName,
					ReadOnly:  true,
					MountPath: constants.TrustedCABundleMountDir,
				}
			)
			BeforeEach(func() {
				cluster.TypeMeta.SetGroupVersionKind(loggingv1.GroupVersion.WithKind("ClusterLogging"))
				client = fake.NewFakeClient( //nolint
					cluster,
					fluentdSecret,
					fluentdCABundle,
					namespace,
				)
				clusterRequest = &ClusterLoggingRequest{
					Client:        client,
					Cluster:       cluster,
					EventRecorder: record.NewFakeRecorder(100),
					Forwarder:     runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName),
					ResourceNames: factory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)),
					ResourceOwner: utils.AsOwner(cluster),
				}
				extras[constants.MigrateDefaultOutput] = true
				spec, extras = migrations.MigrateClusterLogForwarderSpec(clusterRequest.Forwarder.Namespace, clusterRequest.Forwarder.Name, clusterRequest.Forwarder.Spec, clusterRequest.Cluster.Spec.LogStore, extras, clusterRequest.ResourceNames.InternalLogStoreSecret)
				clusterRequest.Forwarder.Spec = spec
			})

			It("should use the injected custom CA bundle for the collector", func() {
				// Reconcile w/o custom CA bundle
				Expect(clusterRequest.CreateOrUpdateCollection()).To(Succeed())

				// Inject custom CA bundle into collector config map
				injectedCABundle := fluentdCABundle.DeepCopy()
				injectedCABundle.Data[constants.TrustedCABundleKey] = customCABundle
				Expect(client.Update(context.TODO(), injectedCABundle)).Should(Succeed())

				// Reconcile with injected custom CA bundle
				Expect(clusterRequest.CreateOrUpdateCollection()).Should(Succeed())

				key := types.NamespacedName{Name: constants.CollectorName, Namespace: cluster.GetNamespace()}
				ds := &appsv1.DaemonSet{}
				Expect(client.Get(context.TODO(), key, ds)).Should(Succeed())

				bundleVar, found := utils.GetEnvVar(common.TrustedCABundleHashName, ds.Spec.Template.Spec.Containers[0].Env)
				Expect(found).To(BeTrue(), "Exp. the trusted bundle CA hash to be added to the collector container")
				Expect(collector.CalcTrustedCAHashValue(injectedCABundle)).To(Equal(bundleVar.Value))
				Expect(ds.Spec.Template.Spec.Volumes).To(ContainElement(trustedCABundleVolume))
				Expect(ds.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElement(trustedCABundleVolumeMount))
			})
		})
		Context("when cluster proxy is not present", func() {
			var (
				trustedCABundleVolume = corev1.Volume{
					Name: constants.CollectorTrustedCAName,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: constants.CollectorTrustedCAName,
							},
							Items: []corev1.KeyToPath{
								{
									Key:  constants.TrustedCABundleKey,
									Path: constants.TrustedCABundleMountFile,
								},
							},
						},
					},
				}
				trustedCABundleVolumeMount = corev1.VolumeMount{
					Name:      constants.CollectorTrustedCAName,
					ReadOnly:  true,
					MountPath: constants.TrustedCABundleMountDir,
				}
			)
			BeforeEach(func() {
				client = fake.NewFakeClient( //nolint
					cluster,
					fluentdSecret,
					fluentdCABundle,
					namespace,
				)
				clusterRequest = &ClusterLoggingRequest{
					Client:        client,
					Cluster:       cluster,
					EventRecorder: record.NewFakeRecorder(100),
					Forwarder:     runtime.NewClusterLogForwarder(constants.OpenshiftNS, "bar"),
					ResourceNames: factory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)),
					ResourceOwner: utils.AsOwner(cluster),
				}
				extras[constants.MigrateDefaultOutput] = true
				spec, extras = migrations.MigrateClusterLogForwarderSpec(clusterRequest.Forwarder.Namespace, clusterRequest.Forwarder.Name, clusterRequest.Forwarder.Spec, clusterRequest.Cluster.Spec.LogStore, extras, clusterRequest.ResourceNames.InternalLogStoreSecret)
				clusterRequest.Forwarder.Spec = spec
			})

			//https://issues.redhat.com/browse/LOG-1859
			It("should continue to reconcile without error", func() {
				Expect(clusterRequest.CreateOrUpdateCollection()).Should(Succeed())

				key := types.NamespacedName{Name: constants.CollectorTrustedCAName, Namespace: cluster.GetNamespace()}
				fluentdCaBundle := &corev1.ConfigMap{}
				Expect(client.Get(context.TODO(), key, fluentdCaBundle)).Should(Succeed())
				Expect(fluentdCABundle.Data).To(Equal(fluentdCaBundle.Data))

				key = types.NamespacedName{Name: constants.CollectorName, Namespace: cluster.GetNamespace()}
				ds := &appsv1.DaemonSet{}
				Expect(client.Get(context.TODO(), key, ds)).Should(Succeed())

				bundleVar, found := utils.GetEnvVar(common.TrustedCABundleHashName, ds.Spec.Template.Spec.Containers[0].Env)
				Expect(found).To(BeTrue(), "Exp. the trusted bundle CA hash to be added to the collector container")
				Expect(bundleVar.Value).To(BeEmpty())
				Expect(ds.Spec.Template.Spec.Volumes).To(Not(ContainElement(trustedCABundleVolume)))
				Expect(ds.Spec.Template.Spec.Containers[0].VolumeMounts).To(Not(ContainElement(trustedCABundleVolumeMount)))
			})
		})

		Context("when creating prometheus rule for collector", func() {
			BeforeEach(func() {
				client = fake.NewFakeClient( //nolint
					cluster,
					fluentdSecret,
					fluentdCABundle,
					namespace,
				)
				clusterRequest = &ClusterLoggingRequest{
					Client:        client,
					Cluster:       cluster,
					EventRecorder: record.NewFakeRecorder(100),
					Forwarder: &loggingv1.ClusterLogForwarder{
						ObjectMeta: metav1.ObjectMeta{
							Name:      constants.SingletonName,
							Namespace: constants.OpenshiftNS,
						},
					},
					ResourceNames: factory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)),
					ResourceOwner: utils.AsOwner(cluster),
				}
				extras[constants.MigrateDefaultOutput] = true
				spec, extras = migrations.MigrateClusterLogForwarderSpec(clusterRequest.Forwarder.Namespace, clusterRequest.Forwarder.Name, clusterRequest.Forwarder.Spec, clusterRequest.Cluster.Spec.LogStore, extras, clusterRequest.ResourceNames.InternalLogStoreSecret)
				clusterRequest.Forwarder.Spec = spec
			})

			It("a fluentd collector should create the logging_collector alerts", func() {
				Expect(clusterRequest.CreateOrUpdateCollection()).To(Succeed())

				collectorKey := types.NamespacedName{Name: constants.CollectorName, Namespace: cluster.GetNamespace()}
				rule := &monitoringv1.PrometheusRule{}
				Expect(client.Get(context.TODO(), collectorKey, rule)).Should(Succeed())
				Expect(rule.Spec.Groups[0].Name).To(Equal("logging_collector.alerts"))
				Expect(rule.Spec.Groups[0].Rules[0].Alert).To(Equal("CollectorNodeDown"))
			})
			It("a vector collector should create the logging_collector alerts", func() {
				// Set collector to vector
				cluster.Spec.Collection.Type = loggingv1.LogCollectionTypeVector
				Expect(clusterRequest.CreateOrUpdateCollection()).To(Succeed())

				collectorKey := types.NamespacedName{Name: constants.CollectorName, Namespace: cluster.GetNamespace()}
				rule := &monitoringv1.PrometheusRule{}
				Expect(client.Get(context.TODO(), collectorKey, rule)).Should(Succeed())
				Expect(rule.Spec.Groups[0].Name).To(Equal("logging_collector.alerts"))
				Expect(rule.Spec.Groups[0].Rules[0].Alert).To(Equal("CollectorNodeDown"))
			})
		})
		Context("when removing collector", func() {
			BeforeEach(func() {
				client = fake.NewFakeClient( //nolint
					cluster,
					fluentdSecret,
					fluentdCABundle,
					namespace,
				)

				clusterRequest = &ClusterLoggingRequest{
					Client:        client,
					Cluster:       cluster,
					EventRecorder: record.NewFakeRecorder(100),
					Forwarder:     &loggingv1.ClusterLogForwarder{},
				}
				extras[constants.MigrateDefaultOutput] = true
				spec, extras = migrations.MigrateClusterLogForwarderSpec(clusterRequest.Forwarder.Namespace, clusterRequest.Forwarder.Name, clusterRequest.Forwarder.Spec, clusterRequest.Cluster.Spec.LogStore, extras, clusterRequest.ResourceNames.InternalLogStoreSecret)
				clusterRequest.Forwarder.Spec = spec
			})
		})

		Context("when creating a ClusterLogForwarder not named instance", func() {
			customCLFName := "custom-clf"
			vectorSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      customCLFName,
					Namespace: cluster.GetNamespace(),
				},
			}
			vectorCABundle := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      customCLFName + "-trustbundle",
					Namespace: cluster.GetNamespace(),
					Labels: map[string]string{
						constants.InjectTrustedCABundleLabel: "true",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							Name:       customCLFName,
							Kind:       "ClusterLogForwarder",
							APIVersion: "logging.openshift.io/v1",
							Controller: utils.GetPtr(true),
						},
					},
				},
				Data: map[string]string{
					constants.TrustedCABundleKey: "",
				},
			}

			fwder := runtime.NewClusterLogForwarder(constants.OpenshiftNS, customCLFName)

			var getObject = func(objName string, obj cli.Object) error {
				nsName := types.NamespacedName{Name: objName, Namespace: cluster.GetNamespace()}
				return client.Get(context.TODO(), nsName, obj)
			}

			BeforeEach(func() {
				client = fake.NewFakeClient( //nolint
					cluster,
					vectorSecret,
					vectorCABundle,
					namespace,
				)
				cluster.Spec.Collection = &loggingv1.CollectionSpec{
					Type: loggingv1.LogCollectionTypeVector,
				}
				clusterRequest = &ClusterLoggingRequest{
					Client:        client,
					Cluster:       cluster,
					EventRecorder: record.NewFakeRecorder(100),
					Forwarder:     fwder,
					ResourceNames: factory.GenerateResourceNames(*runtime.NewClusterLogForwarder(constants.OpenshiftNS, customCLFName)),
					ResourceOwner: utils.AsOwner(fwder),
				}
				extras[constants.MigrateDefaultOutput] = true
				spec, extras = migrations.MigrateClusterLogForwarderSpec(clusterRequest.Forwarder.Namespace, clusterRequest.Forwarder.Name, clusterRequest.Forwarder.Spec, clusterRequest.Cluster.Spec.LogStore, extras, clusterRequest.ResourceNames.InternalLogStoreSecret)
				clusterRequest.Forwarder.Spec = spec
			})
			It("should have appropriately named resources", func() {
				cluster.Spec.Collection.Type = loggingv1.LogCollectionTypeVector
				Expect(clusterRequest.CreateOrUpdateCollection()).To(Succeed())

				// Daemonset
				ds := &appsv1.DaemonSet{}
				Expect(getObject(customCLFName, ds)).Should(Succeed())

			})
		})
	})
})

var _ = Describe("compareFluentdCollectorStatus", func() {
	defer GinkgoRecover()
	var (
		lhs loggingv1.FluentdCollectorStatus
		rhs loggingv1.FluentdCollectorStatus
	)

	BeforeEach(func() {
		lhs = loggingv1.FluentdCollectorStatus{
			DaemonSet:  constants.CollectorName,
			Conditions: map[string]loggingv1.ClusterConditions{},
			Nodes:      map[string]string{},
			Pods:       map[loggingv1.PodStateType][]string{},
		}
		rhs = loggingv1.FluentdCollectorStatus{
			DaemonSet:  constants.CollectorName,
			Conditions: map[string]loggingv1.ClusterConditions{},
			Nodes:      map[string]string{},
			Pods:       map[loggingv1.PodStateType][]string{},
		}
	})
	It("should succeed if everything is the same", func() {
		Expect(compareFluentdCollectorStatus(lhs, rhs)).To(BeTrue())
	})
	It("should determine different names to be different", func() {
		rhs.DaemonSet = "foo"
		Expect(compareFluentdCollectorStatus(lhs, rhs)).To(BeFalse())
	})

	Context("when evaluating nodes", func() {
		It("should fail if they are different lengths", func() {
			rhs.Nodes["foo"] = "bar"
			Expect(compareFluentdCollectorStatus(lhs, rhs)).To(BeFalse())
		})
		It("should fail if the entries are different", func() {
			rhs.Nodes["foo"] = "bar"
			lhs.Nodes["foo"] = "xyz"
			Expect(compareFluentdCollectorStatus(lhs, rhs)).To(BeFalse())
		})

	})

	Context("when evaluating Pods", func() {
		It("should fail if they are different lengths", func() {
			rhs.Pods[loggingv1.PodStateTypeFailed] = []string{"abc"}
			Expect(compareFluentdCollectorStatus(lhs, rhs)).To(BeFalse())
		})
		It("should fail if the entries are different", func() {
			rhs.Pods[loggingv1.PodStateTypeFailed] = []string{"abc"}
			lhs.Pods[loggingv1.PodStateTypeFailed] = []string{"123"}
			Expect(compareFluentdCollectorStatus(lhs, rhs)).To(BeFalse())
		})

	})

})
