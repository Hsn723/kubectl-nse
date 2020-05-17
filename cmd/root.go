package cmd

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"github.com/Hsn723/kubectl-nse/common"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

var (
	rootCmd = &cobra.Command{
		Use:   "kubectl nse",
		Short: "kubectl-nse allows to run a command in a pod's container using nsenter",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runRoot,
	}
	namespace     string
	nsArgs        string
	containerName string
	nodeName      string
	selector      string
	isKind        bool
	withSudo      bool
)

func init() {
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "specify namespace")
	rootCmd.Flags().StringVar(&nsArgs, "nsargs", nsArgs, "quoted options to pass to nsenter")
	rootCmd.Flags().StringVarP(&containerName, "container", "c", containerName, "specify container name for pods with multiple containers")
	rootCmd.Flags().StringVar(&nodeName, "node", nodeName, "specify node to which the pod belongs")
	rootCmd.Flags().StringVarP(&selector, "selector", "l", selector, "selector to filter on")
	rootCmd.Flags().BoolVar(&isKind, "kind", isKind, "the target cluster is a kind cluster")
	rootCmd.Flags().BoolVar(&withSudo, "sudo", withSudo, "execute commands on the remote host with sudo")
}

func runRoot(cmd *cobra.Command, args []string) (err error) {
	var pod v1.Pod
	if len(args) < 1 {
		if selector == "" {
			err = errors.New("a selector must be provided when no pod name is provided")
			return
		}
		pod, err = getPodBySelector()
	} else {
		pod, err = getPod(args[0])
	}
	if err != nil {
		return
	}

	if nodeName == "" {
		nodeName = pod.Spec.NodeName
	} else if nodeName != pod.Spec.NodeName {
		err = errors.New("found pod is not running in the given node")
		return
	}
	cmd.Println("Target node is " + nodeName)
	containerID, runtime, err := getContainerInfo(pod)
	if err != nil {
		return
	}
	cmd.Println("Got container ID " + containerID)
	var rc common.RemoteCmd
	if isKind {
		rc.Path = "docker"
		rc.PreArgs = []string{"exec", "-t", nodeName}
		rc.PreArgsInteractive = []string{"exec", "-it", nodeName}
	} else {
		rc.Path = "ssh"
		rc.PreArgs = []string{"-ttq", nodeName}
		rc.PreArgsInteractive = rc.PreArgs
	}
	if withSudo {
		rc.PreArgs = append(rc.PreArgs, "sudo")
		rc.PreArgsInteractive = append(rc.PreArgsInteractive, "sudo")
	}
	pid, err := rc.GetPid(containerID, runtime)
	if err != nil {
		return
	}
	if pid == "0" {
		err = errors.New("a container was found but with a PID of 0, this usually means it is not running")
		return
	}
	cmd.Println("Got PID " + pid)
	return rc.Enter(pid, nsArgs)
}

func getPodBaseArgs() []string {
	getPodArgs := []string{"get", "pod"}
	if namespace != "" {
		getPodArgs = append(getPodArgs, "-n", namespace)
	}
	getPodArgs = append(getPodArgs, "-o=json")
	return getPodArgs
}

func getPod(podName string) (pod v1.Pod, err error) {
	getPodArgs := getPodBaseArgs()
	getPodArgs = append(getPodArgs, podName)
	getPodCmd := exec.Command("kubectl", getPodArgs...)
	getPodOut, err := getPodCmd.Output()
	if err != nil {
		return
	}
	err = json.Unmarshal(getPodOut, &pod)
	if err != nil {
		return
	}
	return
}

func getPodBySelector() (pod v1.Pod, err error) {
	getPodArgs := getPodBaseArgs()
	getPodArgs = append(getPodArgs, "-l", selector)
	getPodCmd := exec.Command("kubectl", getPodArgs...)
	getPodOut, err := getPodCmd.Output()
	if err != nil {
		return
	}
	var podList v1.PodList
	err = json.Unmarshal(getPodOut, &podList)
	if err != nil {
		return
	}
	if len(podList.Items) == 0 {
		err = errors.New("no matching pod has been found")
		return
	}
	if len(podList.Items) == 1 {
		pod = podList.Items[0]
		return
	}
	if nodeName != "" {
		var matchingPods []v1.Pod
		for _, p := range podList.Items {
			if p.Spec.NodeName == nodeName {
				matchingPods = append(matchingPods, p)
			}
		}
		if len(matchingPods) == 0 {
			err = errors.New("no matching pod has been found on the given node")
			return
		}
		if len(matchingPods) == 1 {
			pod = matchingPods[0]
			return
		}
		return selectPod(matchingPods)
	}

	return selectPod(podList.Items)
}

func selectPod(podList []v1.Pod) (pod v1.Pod, err error) {
	var podChoices []string
	for _, p := range podList {
		podChoices = append(podChoices, p.Name)
	}
	prompt := promptui.Select{
		Label: "Multiple pods were matched, please select one:",
		Items: podChoices,
	}
	_, podName, err := prompt.Run()
	if err != nil {
		return
	}
	for _, p := range podList {
		if p.Name == podName {
			pod = p
			return
		}
	}
	err = errors.New("no matching pod has been found")
	return
}

func getContainerID(containers []v1.ContainerStatus) (containerID string, err error) {
	if len(containers) == 1 {
		container := containers[0]
		if containerName != "" && container.Name != containerName {
			err = errors.New("found a container, but it does not match the requested container name")
			return
		}
		containerID = container.ContainerID
		return
	}
	if containerName == "" {
		var choices []string
		for _, c := range containers {
			choices = append(choices, c.Name)
		}
		prompt := promptui.Select{
			Label: "Multiple containers were found, please select one:",
			Items: choices,
		}
		_, cn, e := prompt.Run()
		if e != nil {
			err = e
			return
		}
		containerName = cn
	}
	for _, c := range containers {
		if c.Name == containerName {
			containerID = c.ContainerID
			break
		}
	}
	if containerID == "" {
		err = errors.New("could not find a matching container")
		return
	}
	return
}

func getContainerInfo(pod v1.Pod) (id string, runtime string, err error) {
	containers := pod.Status.ContainerStatuses
	containerID, err := getContainerID(containers)
	if err != nil {
		return
	}
	containerInfoParts := strings.Split(containerID, "://")
	id = containerInfoParts[1]
	runtime = containerInfoParts[0]
	return
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
