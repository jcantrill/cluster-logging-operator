package filter

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filter", func() {

	var (
		builder *RecordTransformerFilter
	)

	Context("when building a record_transformer", func() {

		BeforeEach(func() {
			builder = NewRecordTransformerFilterBuilder("**").
				EnableRuby(true).
				AddToRecord("msg_size", "${record.to_s.length}")
		})

		It("should allow altering a record", func() {
			Expect(builder.AsList()).Should(
				ConsistOf([]string{
					"<filter **>",
					"@type record_transformer",
					"enable_ruby true",
					"<record>",
					"msg_size ${record.to_s.length}",
					"</record>",
					"</filter>",
				}))
		})
	})

})
