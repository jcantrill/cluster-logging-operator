//go:build vector

package splunk

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Forwarding to Splunk", func() {
	var (
		framework *functional.CollectorFunctionalFramework
	)
	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
	})

	It("should accept application logs", func() {

		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToSplunkOutput()
		Expect(framework.Deploy()).To(BeNil())

		timestamp := "2020-11-04T18:13:59.061892+00:00"
		applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())
		logs, err := framework.ReadApplicationLogsFrom(logging.OutputTypeSplunk)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		outputTestLog := logs[0]
		outputLogTemplate := functional.NewApplicationLogTemplate()
		Expect(outputTestLog).To(FitLogFormatTemplate(outputLogTemplate))

	})

})
