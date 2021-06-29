package output
import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Logforwarder forward output", func() {

	var (
		builder *ForwardOutputBuilder
		secrets map[string]*corev1.Secret
	)

	BeforeEach(func() {
		builder = NewForwardOutputBuilder()
		secrets = nil
	})

	Context("when building the configuration", func(){
		It("should generate the desired forward label", func() {
			Expect(builder.AsList()).Should(
				ConsistOf([]string{
					"<label @SPECIAL_OUTPUT>",
					"<match **>",
					"@type forward",
					"heartbeat_type none",
					"keepalive true",
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
	})



})