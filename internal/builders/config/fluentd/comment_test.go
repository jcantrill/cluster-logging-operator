package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Comment", func() {

	var (
		builder Comment
	)

	BeforeEach(func() {
		builder = Comment("my comment")
	})

	It("#AsList should return the comment as a formatted ruby comment", func() {
		Expect(builder.AsList()).Should(
			ConsistOf([]string{"# my comment"}))
	})

})
