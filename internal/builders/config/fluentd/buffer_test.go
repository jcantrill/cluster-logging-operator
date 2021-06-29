package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Buffer configuration", func() {

	var (
		builder *Buffer
	)

	Context("when building a file buffer", func() {

		BeforeEach(func() {
			builder = NewBuffer(BufferTypeFile).
				WithPath("/var/log/mypath")
		})

		It("should include the path", func() {
			Expect(builder.AsList()).Should(
				ConsistOf([]string{
					"<buffer>",
					"@type file",
					"path /var/log/mypath",
					"</buffer>",
				}))
		})

		It("should exclude unallowed keys", func() {
			builder.Set("foo", "bar")
			Expect(builder.AsList()).Should(
				ConsistOf([]string{
					"<buffer>",
					"@type file",
					"path /var/log/mypath",
					"</buffer>",
				}))
		})
	})

})
