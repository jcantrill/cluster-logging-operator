package podman

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
)

var _ = Describe("tet", func() {
	var (
		p PodCommand
	)
	AfterEach(func() {
		// Expect(p.Remove().Error()).To(BeNil())
		// Expect(p).NotTo(BeNil())
	})
	It("", func() {
		p = Pod("fluent").
			WithImage("quay.io/openshift/origin-logging-fluentd:latest").
			AddVolume(utils.DefaultWorkingDir, "/etc/fluent/metrics").
			AddVolume("/tmp/run", "/opt/app-root/src").
			AddVolume("/tmp/run", "/etc/fluent").
			Run()
		Expect(p.State()).To(Equal("Running"))
	})
})
