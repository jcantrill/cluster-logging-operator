package source

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("in_tail configuration", func() {

	var (
		builder *Tail
	)

	Context("when building", func() {

		BeforeEach(func() {
			builder = NewTailBuilder("/var/log/mypath.log")
		})

		It("should require the path", func() {
			Expect(builder.AsList()).Should(
				ConsistOf([]string{
					"<source>",
					"@type tail",
					"path /var/log/mypath.log",
					"</source>",
				}))
		})
	})

})
