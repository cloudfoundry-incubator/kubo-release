package kubernetesdeployment_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	k8sdeploy "ensure-specs-running/Kubernetesdeployment"

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Kubernetesdeployment", func() {
	Describe("Discrepancies", func() {
		Context("when the actual set has a missing deployment", func() {
			It("detects a discrepancy", func() {
				expected := k8sdeploy.DeploymentSet{"deployment-1": k8sdeploy.NewDeployment(1, []string{"c1"})}
				actual := k8sdeploy.DeploymentSet{}

				Expect(k8sdeploy.Discrepancies(expected, actual)).To(ContainElement("unable to find any pods for deployment-1"))
			})
		})

		Context("when the actual set has an extra deployment", func() {
			It("does not report this as a discrepancy", func() {
				expected := k8sdeploy.DeploymentSet{}
				actual := k8sdeploy.DeploymentSet{"deployment-1": k8sdeploy.NewDeployment(1, []string{"c1"})}

				Expect(k8sdeploy.Discrepancies(expected, actual)).NotTo(ContainElement(HavePrefix("unable to find any pods for")))
			})
		})

		Context("when the actual set has an expected deployment", func() {
			Context("when the actual set's deployment has a different number of replicas", func() {
				It("detects a discrepancy", func() {
					expected := k8sdeploy.DeploymentSet{"deployment-1": k8sdeploy.NewDeployment(33, []string{"c1"})}
					actual := k8sdeploy.DeploymentSet{"deployment-1": k8sdeploy.NewDeployment(44, []string{"c1"})}

					Expect(k8sdeploy.Discrepancies(expected, actual)).
						To(ContainElement("expected replica count for deployment-1 is 33, found 44"))
				})
			})

			Context("when the actual set's deployment has a different set of containers", func() {
				It("detects a discrepancy", func() {
					expected := k8sdeploy.DeploymentSet{"deployment-1": k8sdeploy.NewDeployment(1, []string{"c1", "c2"})}
					actual := k8sdeploy.DeploymentSet{"deployment-1": k8sdeploy.NewDeployment(1, []string{"c3", "c4"})}

					Expect(k8sdeploy.Discrepancies(expected, actual)).
						To(ContainElement("expected ready containers [c1, c2] for deployment-1, found [c3, c4]"))
				})
			})

			Context("when the actual set's deployment has the same set of containers", func() {
				Context("when the actual containers are listed in a different order", func() {
					It("does not report this as a discrepancy", func() {
						expected := k8sdeploy.DeploymentSet{"deployment-1": k8sdeploy.NewDeployment(1, []string{"c1", "c2"})}
						actual := k8sdeploy.DeploymentSet{"deployment-1": k8sdeploy.NewDeployment(1, []string{"c2", "c1"})}

						Expect(k8sdeploy.Discrepancies(expected, actual)).NotTo(ContainElement(HavePrefix("expected ready containers")))
					})
				})

				Context("when the actual containers are listed in a different order", func() {
					It("does not report this as a discrepancy", func() {
						expected := k8sdeploy.DeploymentSet{"deployment-1": k8sdeploy.NewDeployment(1, []string{"c1", "c2"})}
						actual := k8sdeploy.DeploymentSet{"deployment-1": k8sdeploy.NewDeployment(1, []string{"c1", "c2"})}

						Expect(k8sdeploy.Discrepancies(expected, actual)).NotTo(ContainElement(HavePrefix("expected ready containers")))
					})
				})
			})
		})

		Describe("a comprehensive example with multiple discrepant deployments, some with multiple discrepancies", func() {
			It("detects multiple discrepancies", func() {
				expected := k8sdeploy.DeploymentSet{
					"missing-in-actual": k8sdeploy.NewDeployment(1, []string{"c0"}),
					"multiple-problems": k8sdeploy.NewDeployment(33, []string{"c1", "c2"}),
					"different-order":   k8sdeploy.NewDeployment(55, []string{"c5", "c6"}),
					"no-problems":       k8sdeploy.NewDeployment(77, []string{"c7"}),
				}
				actual := k8sdeploy.DeploymentSet{
					"multiple-problems": k8sdeploy.NewDeployment(44, []string{"c3", "c4"}),
					"different-order":   k8sdeploy.NewDeployment(66, []string{"c6", "c5"}),
					"no-problems":       k8sdeploy.NewDeployment(77, []string{"c7"}),
					"extra":             k8sdeploy.NewDeployment(88, []string{"c8"}),
				}

				Expect(k8sdeploy.Discrepancies(expected, actual)).To(ConsistOf(
					"unable to find any pods for missing-in-actual",
					"expected replica count for multiple-problems is 33, found 44",
					"expected ready containers [c1, c2] for multiple-problems, found [c3, c4]",
					"expected replica count for different-order is 55, found 66",
				))
			})
		})
	})

	Describe("AddPod", func() {
		var deployment *k8sdeploy.Deployment
		var pod corev1.Pod
		var expectDeploymentToMatch = func(expected *k8sdeploy.Deployment) {
			dummyName := "dummy-name"
			Expect(k8sdeploy.Discrepancies(
				k8sdeploy.DeploymentSet{dummyName: expected},
				k8sdeploy.DeploymentSet{dummyName: deployment},
			)).To(BeEmpty())
		}

		JustBeforeEach(func() {
			deployment.AddPod(pod)
		})

		Context("when the deployment is empty", func() {
			BeforeEach(func() {
				deployment = k8sdeploy.NewDeployment(0, []string{})
			})

			Context("when the pod has no containers", func() {
				BeforeEach(func() {
					pod = corev1.Pod{}
				})

				It("increments the replica count only", func() {
					expectDeploymentToMatch(k8sdeploy.NewDeployment(1, []string{}))
				})
			})

			Context("when the pod has containers but none are ready", func() {
				BeforeEach(func() {
					pod = corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
						{Name: "some-container", Ready: false},
						{Name: "another-container", Ready: false},
					}}}
				})

				It("increments the replica count only", func() {
					expectDeploymentToMatch(k8sdeploy.NewDeployment(1, []string{}))
				})
			})

			Context("when the pod has containers and all are ready", func() {
				BeforeEach(func() {
					pod = corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
						{Name: "some-container", Ready: true},
						{Name: "another-container", Ready: true},
					}}}
				})

				It("increments the replica count and adds all containers", func() {
					expectDeploymentToMatch(k8sdeploy.NewDeployment(1, []string{"some-container", "another-container"}))
				})
			})

			Context("when the pod has containers and only some are ready", func() {
				BeforeEach(func() {
					pod = corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
						{Name: "some-container", Ready: true},
						{Name: "another-container", Ready: false},
					}}}
				})

				It("increments the replica count and adds only the ready containers", func() {
					expectDeploymentToMatch(k8sdeploy.NewDeployment(1, []string{"some-container"}))
				})
			})
		})

		Context("when the deployment is non-empty", func() {
			BeforeEach(func() {
				deployment = k8sdeploy.NewDeployment(2, []string{"container-1a", "container-1b", "container-2a"})
			})

			Context("when the pod has no containers", func() {
				BeforeEach(func() {
					pod = corev1.Pod{}
				})

				It("increments the replica count only", func() {
					expectDeploymentToMatch(k8sdeploy.NewDeployment(3, []string{"container-1a", "container-1b", "container-2a"}))
				})
			})

			Context("when the pod has containers but none are ready", func() {
				BeforeEach(func() {
					pod = corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
						{Name: "some-container", Ready: false},
						{Name: "another-container", Ready: false},
					}}}
				})

				It("increments the replica count only", func() {
					expectDeploymentToMatch(k8sdeploy.NewDeployment(3, []string{"container-1a", "container-1b", "container-2a"}))
				})
			})

			Context("when the pod has containers and all are ready", func() {
				BeforeEach(func() {
					pod = corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
						{Name: "some-container", Ready: true},
						{Name: "another-container", Ready: true},
					}}}
				})

				It("increments the replica count and adds all containers", func() {
					expectDeploymentToMatch(k8sdeploy.NewDeployment(
						3,
						[]string{
							"some-container",
							"another-container",
							"container-1a",
							"container-1b",
							"container-2a",
						},
					))
				})
			})

			Context("when the pod has containers and only some are ready", func() {
				BeforeEach(func() {
					pod = corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
						{Name: "some-container", Ready: true},
						{Name: "another-container", Ready: false},
					}}}
				})

				It("increments the replica count and adds only the ready containers", func() {
					expectDeploymentToMatch(k8sdeploy.NewDeployment(
						3,
						[]string{
							"some-container",
							"container-1a",
							"container-1b",
							"container-2a",
						},
					))
				})
			})

			Context("when the pod has ready containers with the same name as in the deployment", func() {
				BeforeEach(func() {
					pod = corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
						{Name: "container-1a", Ready: true},
					}}}
				})

				It("adds the pod's containers (does not de-dupe)", func() {
					expectDeploymentToMatch(k8sdeploy.NewDeployment(
						3,
						[]string{
							"container-1a",
							"container-1a",
							"container-1b",
							"container-2a",
						},
					))
				})
			})
		})
	})
})
