package handlers_test

import (
	"os"
	"path"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/cmd/log-node-agent/internal/handlers"
)

var _ = Describe("handlers", func() {

	const (
		crioRoot      = "openshift-operator-lifecycle-manager_packageserver-5f89579495-7g57f_706603d5-69bf-4e76-9f1a-7664d919afa3"
		expNamespace  = "openshift-operator-lifecycle-manager"
		varlogpods    = "/var/log/pods"
		targetRoot    = "/var/lib/ocp-logging"
		containerName = "mine"
	)

	//DescribeTable("#extractNamespace", func(exp, path string) {
	//	Expect(extractNamespace(path)).To(Equal(exp))
	//},
	//	Entry("from multi-parent directories", expNamespace, "/var/log/pods/"+crioRoot),
	//	Entry("from single parent directory", expNamespace, "/"+crioRoot),
	//	Entry("from parent with container directory", expNamespace, "/"+crioRoot+"/mine"),
	//	Entry("from file log stream", expNamespace, "/"+crioRoot+"/mine/0.log"),
	//)
	//
	//DescribeTable("#stream", func(exp, path string) {
	//	Expect(stream(path)).To(Equal(exp))
	//},
	//	Entry("from multi-parent directories", crioRoot, "/var/log/pods/"+crioRoot),
	//	Entry("from single parent directory", crioRoot, "/"+crioRoot),
	//	Entry("from parent with container directory", crioRoot+"/mine", "/"+crioRoot+"/mine"),
	//	Entry("from file log stream", crioRoot+"/mine/0.log", "/"+crioRoot+"/mine/0.log"),
	//)

	DescribeTable("#NewContainerLogStream", func(streamPath string) {
		relPath := strings.Replace(streamPath, "/var/log/pods/", "", 1)
		act := handlers.NewContainerLogStream(targetRoot, streamPath)
		Expect(act.Namespace).To(Equal(expNamespace))
		Expect(act.OldName).To(Equal(streamPath))
		Expect(act.NewName).To(Equal(path.Join(targetRoot, expNamespace, relPath)))
	},
		Entry("directory in /var/lib/pods", "/var/log/pods/"+crioRoot),
		Entry("container directory for the pod", "/var/log/pods/"+crioRoot+"/"+containerName),
		Entry("log file for a container", "/var/log/pods/"+crioRoot+"/"+containerName+"/0.log"),
	)
	Context("ContainerLogStream#DirEntries", func() {
		It("should create entries for container directories", func() {
			cls := handlers.ContainerLogStream{
				TargetPathRoot: targetRoot,
				Namespace:      expNamespace,
				OldName:        varlogpods + "/" + crioRoot,
				NewName:        path.Join(targetRoot, expNamespace, crioRoot),
				Os: FakePathOperator{
					StatFn: func(name string) (os.FileInfo, error) {
						return FakeFileInfo{IsADir: true}, nil
					},
					ReadDirFn: func(name string) ([]os.DirEntry, error) {
						return []os.DirEntry{
							FakeDirectoryEntry{
								AName: containerName,
							},
						}, nil
					},
				},
			}
			act := cls.DirEntries()[0]
			Expect(act.Namespace).To(Equal(expNamespace), "Namespace")
			Expect(act.OldName).To(Equal(path.Join(varlogpods, crioRoot, containerName)), "OldName")
			Expect(act.NewName).To(Equal(path.Join(targetRoot, expNamespace, crioRoot, containerName)), "NewName")
		})
	})

})
