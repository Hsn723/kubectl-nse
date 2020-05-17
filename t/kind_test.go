package test

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var commonArgs = []string{
	"../artifacts/kubectl-nse-linux-amd64",
	"--kind",
	"--nsargs='-n'",
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyPollingInterval(time.Second)
	SetDefaultEventuallyTimeout(5 * time.Minute)
	fmt.Println("Setting up additional pods")
	_, _, err := Exec("kubectl", "apply", "-f", "kind/test-pods.yaml")
	Expect(err).NotTo(HaveOccurred())
	checkSync := func() error {
		_, _, err := Exec("kubectl", "get", "pod", "-A", "-o", "json", "|", "jq", ".items[].status.phase | grep -qv Running")
		if err == nil {
			return errors.New("not yet")
		}
		return nil
	}
	Context("wait for sync", func() {
		Eventually(checkSync).Should(Succeed())
	})
})

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test")
}

var _ = Describe("Test", func() {
	Context("when no pod or selector is provided", func() {
		It("should warn", func() {
			_, e, _ := Exec(append(commonArgs, "uname")...)
			Expect(string(e)).To(ContainSubstring("Error: a selector must be provided when no pod name is provided"))
		})
	})
	Context("when specified pod doesn't exist", func() {
		It("should fail", func() {
			_, e, _ := Exec(append(commonArgs, "-p", "non-existing-pod")...)
			Expect(e).To(ContainSubstring("Error: exit status 1"))
		})
	})
	Context("when selector has no pods", func() {
		It("should warn", func() {
			_, e, _ := Exec(append(commonArgs, "-l", "test=no-matching-pod")...)
			Expect(e).To(ContainSubstring("Error: no matching pod has been found"))
		})
	})
	Context("when specifying pod ID", func() {
		It("should succeed", func() {
			_, e, _ := Exec(append(commonArgs, "-n", "kube-system", "-p", "kube-scheduler-kind-control-plane", "uname")...)
			ExpectNoTTYError(e)
		})
	})
	Context("when selector matches but is not on the given node", func() {
		It("should fail", func() {
			_, e, _ := Exec(append(commonArgs, "-n", "kube-system", "-l", "component=kube-scheduler", "--node", "kind-worker")...)
			Expect(e).To(ContainSubstring("Error: found pod is not running in the given node"))
		})
	})
	Context("when selector matches", func() {
		It("should succeed", func() {
			_, e, _ := Exec(append(commonArgs, "-n", "kube-system", "-l", "component=kube-scheduler", "uname")...)
			ExpectNoTTYError(e)
		})
	})
	Context("when selector has multiple matches", func() {
		It("should succeed", func() {
			_, e, _ := Exec(append(commonArgs, "-n", "kube-system", "-l", "app=kindnet", "--node", "kind-worker", "uname")...)
			ExpectNoTTYError(e)
		})
	})
	Context("when selected pod has multiple containers", func() {
		It("should succeed", func() {
			_, e, _ := Exec(append(commonArgs, "-l", "test=single-pod-multiple-containers", "-c", "busybox-2", "uname")...)
			ExpectNoTTYError(e)
		})
	})
})

func ExpectNoTTYError(e []byte) {
	Expect(e).To(ContainSubstring("the input device is not a TTY"))
}

func ExecWithInput(input []byte, args ...string) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("sh", "-c", strings.Join(args, " "))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func Exec(args ...string) (stdout, stderr []byte, err error) {
	return ExecWithInput(nil, args...)
}
