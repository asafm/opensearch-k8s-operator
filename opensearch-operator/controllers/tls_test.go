package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	opsterv1 "opensearch.opster.io/api/v1"
	"opensearch.opster.io/pkg/helpers"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("TLS Reconciler", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		clusterName = "cluster-test-tls"
		namespace   = clusterName
		timeout     = time.Second * 30
		interval    = time.Second * 1
	)
	Context("When Creating an OpenSearchCluster with TLS configured", func() {
		spec := opsterv1.OpenSearchCluster{
			ObjectMeta: metav1.ObjectMeta{Name: clusterName, Namespace: namespace},
			Spec: opsterv1.ClusterSpec{
				General: opsterv1.GeneralConfig{ServiceName: clusterName},
				Security: &opsterv1.Security{Tls: &opsterv1.TlsConfig{
					Transport: &opsterv1.TlsConfigTransport{
						Generate: true,
						PerNode:  true,
					},
					Http: &opsterv1.TlsConfigHttp{
						Generate: true,
					},
				}},
				NodePools: []opsterv1.NodePool{
					{
						Component: "masters",
						Replicas:  3,
						Roles:     []string{"master", "data"},
					},
				},
			}}

		It("Should create the namespace first", func() {
			Expect(CreateNamespace(k8sClient, &spec)).Should(Succeed())
			By("Create cluster ns ")
			Eventually(func() bool {
				return IsNsCreated(k8sClient, namespace)
			}, timeout, interval).Should(BeTrue())
		})

		It("should apply the cluster instance successfully", func() {
			Expect(k8sClient.Create(context.Background(), &spec)).Should(Succeed())
		})

		It("Should start a cluster successfully", func() {
			By("Checking for Statefulset")
			sts := appsv1.StatefulSet{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), client.ObjectKey{Name: clusterName + "-masters", Namespace: namespace}, &sts)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(*sts.Spec.Replicas).To(Equal(int32(3)))
			Expect(helpers.CheckVolumeExists(sts.Spec.Template.Spec.Volumes, sts.Spec.Template.Spec.Containers[0].VolumeMounts, clusterName+"-transport-cert", "transport-cert")).Should((BeTrue()))
			Expect(helpers.CheckVolumeExists(sts.Spec.Template.Spec.Volumes, sts.Spec.Template.Spec.Containers[0].VolumeMounts, clusterName+"-http-cert", "http-cert")).Should((BeTrue()))
			Expect(helpers.CheckVolumeExists(sts.Spec.Template.Spec.Volumes, sts.Spec.Template.Spec.Containers[0].VolumeMounts, clusterName+"-config", "config")).Should((BeTrue()))
		})

		It("Should set correct owner references", func() {
			cm := corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: clusterName + "-config", Namespace: namespace}, &cm)).To(Succeed())
			Expect(HasOwnerReference(&cm, &spec)).To(BeTrue())

			secret := corev1.Secret{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: clusterName + "-http-cert", Namespace: namespace}, &secret)).To(Succeed())
			Expect(HasOwnerReference(&secret, &spec)).To(BeTrue())

			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: clusterName + "-transport-cert", Namespace: namespace}, &secret)).To(Succeed())
			Expect(HasOwnerReference(&secret, &spec)).To(BeTrue())

			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: clusterName + "-ca", Namespace: namespace}, &secret)).To(Succeed())
			Expect(HasOwnerReference(&secret, &spec)).To(BeTrue())
		})
	})

})
