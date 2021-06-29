package output

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OutFile", func() {

	var (
		builder *OutFile
	)

	Context("when building the configuration", func() {

		BeforeEach(func() {
			builder = NewOutFileBuilder("mypattern", MatchTypeFile).
				WithPath("/var/log/mypath").
				WithCompress("gzip")
			builder.WithBuffer().
				WithTimeKey("1d").
				WithTimeKeyUseUTC(true).
				WithTimeKeyWait("10")
		})

		It("should create a valid config", func() {
			Expect(builder.AsList()).Should(
				ConsistOf([]string{
					"<match mypattern>",
					"@type file",
					"path /var/log/mypath",
					"compress gzip",
					"<buffer>",
					"timekey 1d",
					"timekey_use_utc true",
					"timekey_wait 10",
					"</buffer>",
					"</match>",
				}))
		})
	})

})
