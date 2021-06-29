package output

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OutForward", func() {

	var (
		builder *OutForward
	)

	Context("when building the configuration", func() {

		BeforeEach(func() {
			builder = NewOutForwardBuilder("**")
			builder.AddServer("192.168.1.3").
				WithName("myserver1").
				WithPort(24224).
				WithWeight(60)
			builder.AddServer("192.168.1.4").
				WithName("myserver2").
				WithPort(24224).
				WithWeight(60)
		})

		It("should create the config", func() {
			Expect(builder.AsList()).Should(
				ConsistOf([]string{
					"<match **>",
					"@type forward",
					"<server>",
					"name myserver1",
					"host 192.168.1.3",
					"port 24224",
					"weight 60",
					"</server>",
					"<server>",
					"name myserver2",
					"host 192.168.1.4",
					"port 24224",
					"weight 60",
					"</server>",
					"</match>",
				}))
		})
	})

})
