package pipeline

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Logforwarder pipeline", func() {

	var (
		builder *PipelineToOutputs
		pipeline logging.PipelineSpec
	)

	BeforeEach(func() {
		builder = NewPipelineToOutputsBuilder(pipeline)
	})

	Context("when building the configuration", func(){
		It("should generate the desired pipeline", func() {
			Expect(builder.AsList()).Should(
				ConsistOf([]string{
					"<label @SPECIAL_OUTPUT>",
					"<match **>",
					"@type forward",
					"heartbeat_type none",
					"keepalive true",
					"<security>",
					"self_hostname #{ENV['NODE_NAME']}",
					"shared_key mykey",
					"</security>",
					"<buffer>",
					"@type file",
					"path /var/lib/fluentd/special_output",
					"queued_chunks_limit_size",
					"#{ENV['BUFFER_QUEUE_LIMIT'] || '1024' }",
					"total_limit_size \"#{ENV['TOTAL_LIMIT_SIZE'] ||  8589934592 }\"",
					"chunk_limit_size \"#{ENV['BUFFER_SIZE_LIMIT'] || '1m'}",
					"flush_mode interval",
					"flush_interval 5s",
					"flush_at_shutdown true",
					"flush_thread_count 2",
					"retry_type exponential_backoff",
					"retry_wait 1s",
					"retry_max_interval 60s",
					"retry_timeout 10m",
					"overflow_action block",
					"</buffer>",
					"<server>",
					"host fluentdserver.security.example.com",
					"port 24224",
					"</server>",
					"</match>",
					"</label>",
				}))
		})

		Context("when there are pipeline labels", func(){

		})
		Context("when configured to parse json", func(){

		})

		Context("when there is more then one output", func(){
			BeforeEach(func(){
				pipeline = logging.PipelineSpec{
					Name:       "apps-pipeline",
					InputRefs:  []string{"myInput"},
					OutputRefs: []string{"apps-es-1", "apps-es-2"},
				}
				builder = NewPipelineToOutputsBuilder(pipeline)
			})
			It("should generate the desired pipeline", func() {
				Expect(builder.String()).Should(
				EqualTrimLines(`
  <label @APPS_PIPELINE>
    <match **>
      @type copy
      <store>
        @type relabel
        @label @APPS_ES_1
      </store>
      <store>
        @type relabel
        @label @APPS_ES_2
     </store>
    </match>
  </label>`))
			})
		})
	})




})