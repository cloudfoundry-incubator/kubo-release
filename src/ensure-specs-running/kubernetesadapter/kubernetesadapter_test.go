package kubernetesadapter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	k8sadapter "ensure-specs-running/kubernetesadapter"
)

var _ = Describe("Kubernetesadapter", func() {
	var adapter k8sadapter.Adapter
	var fakeClientset *fake.Clientset
	const namespace = "test-namespace"

	BeforeEach(func() {
		fakeClientset = new(fake.Clientset)
		adapter = k8sadapter.NewAdapter(fakeClientset, namespace)
	})

	Describe("Pods", func() {
		Context("when the kubernetes client has an error", func() {
			BeforeEach(func() {
				fakeClientset.PrependReactor("list", "pods", func(_ k8stesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("list-pods-error")
				})
			})

			It("bubbles up the error", func() {
				_, err := adapter.Pods()
				Expect(err).To(MatchError("list-pods-error"))
			})
		})

		Context("when the kubernetes client gets pods", func() {
			var pod1, pod2 corev1.Pod

			BeforeEach(func() {
				pod1 = corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}}
				pod2 = corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}}
				fakeClientset.PrependReactor("list", "pods", func(_ k8stesting.Action) (bool, runtime.Object, error) {
					return true, &corev1.PodList{Items: []corev1.Pod{pod1, pod2}}, nil
				})

			})

			It("returns pods", func() {
				pods, err := adapter.Pods()
				Expect(err).NotTo(HaveOccurred())
				Expect(pods).To(ConsistOf(pod1, pod2))
			})
		})
	})

	Describe("ExtractDeploymentName", func() {
		var pod = corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-name"}}

		Context("when the pod has no owner", func() {
			BeforeEach(func() {
				pod.ObjectMeta.OwnerReferences = []metav1.OwnerReference{}
			})

			It("returns an error", func() {
				_, err := adapter.ExtractDeploymentName(pod)
				Expect(err).To(MatchError("expected pod pod-name to have 1 owner, has 0"))
			})
		})

		Context("when the pod has multiple owners", func() {
			BeforeEach(func() {
				pod.ObjectMeta.OwnerReferences = []metav1.OwnerReference{{Name: "owner-1"}, {Name: "owner-2"}}
			})

			It("returns an error", func() {
				_, err := adapter.ExtractDeploymentName(pod)
				Expect(err).To(MatchError("expected pod pod-name to have 1 owner, has 2"))
			})
		})

		Context("when the pod has one owner", func() {
			BeforeEach(func() {
				pod.ObjectMeta.OwnerReferences = []metav1.OwnerReference{{Name: "owning-replica-set"}}
			})

			Context("when the kubernetes client has an error", func() {
				BeforeEach(func() {
					fakeClientset.PrependReactor("get", "replicasets", func(_ k8stesting.Action) (bool, runtime.Object, error) {
						return true, nil, errors.New("get-replicaset-error")
					})
				})

				It("bubbles up the error", func() {
					_, err := adapter.ExtractDeploymentName(pod)
					Expect(err).To(MatchError("get-replicaset-error"))
				})
			})

			Context("when the kubernetes client gets the owning replicaset, and its owning deployment, for the pod", func() {
				BeforeEach(func() {
					fakeClientset.PrependReactor("get", "replicasets", func(_ k8stesting.Action) (bool, runtime.Object, error) {
						rs := extv1beta1.ReplicaSet{ObjectMeta: metav1.ObjectMeta{
							Name:            "owning-replica-set",
							OwnerReferences: []metav1.OwnerReference{{Name: "owning-deployment"}},
						}}
						return true, &rs, nil
					})
					fakeClientset.PrependReactor("get", "deployments", func(_ k8stesting.Action) (bool, runtime.Object, error) {
						return true, &extv1beta1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "owning-deployment"}}, nil
					})
				})

				It("returns the deployment name", func() {
					name, err := adapter.ExtractDeploymentName(pod)
					Expect(err).NotTo(HaveOccurred())
					Expect(name).To(Equal("owning-deployment"))
				})
			})
		})
	})
})
