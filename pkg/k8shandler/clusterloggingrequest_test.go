package k8shandler

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

var _ = Describe("ClusterLoggingRequest", func() {
	var (
		clr *ClusterLoggingRequest
	)
	Context("#isManaged", func() {
		BeforeEach(func() {
			clr = &ClusterLoggingRequest{
				Cluster: &logging.ClusterLogging{},
			}
		})
		It("should return true for managed state", func() {
			clr.Cluster.Spec.ManagementState = logging.ManagementStateManaged
			Expect(clr.isManaged()).To(BeTrue())
		})
		It("should return true for an unset managed state", func() {
			Expect(clr.isManaged()).To(BeTrue())

		})
		It("should return false for an unmanaged state", func() {
			clr.Cluster.Spec.ManagementState = logging.ManagementStateUnmanaged
			Expect(clr.isManaged()).To(BeFalse())
		})
	})
})
