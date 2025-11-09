package handlers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("#namespace", func(exp, path string) {
	Expect(namespace(path)).To(Equal(exp))
},
	Entry("from multi-parent directories", "openshift-operator-lifecycle-manager","/var/log/pods/openshift-operator-lifecycle-manager_packageserver-5f89579495-7g57f_706603d5-69bf-4e76-9f1a-7664d919afa3"),
	Entry("from single parent directory", "openshift-operator-lifecycle-manager","/openshift-operator-lifecycle-manager_packageserver-5f89579495-7g57f_706603d5-69bf-4e76-9f1a-7664d919afa3"),
)
