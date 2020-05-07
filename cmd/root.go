package cmd

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"github.com/Hsn723/kubectl-nse/common"
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
	if len(podList.Items) == 1 {
		pod = podList.Items[0]
		return
	}
	if nodeName == "" {
		err = errors.New("a node name must be provided when there are multiple pods matching a selector")
		return
	}
	var matchingPods []v1.Pod
	for _, p := range podList.Items {
		if p.Spec.NodeName == nodeName {
			matchingPods = append(matchingPods, p)
		}
	}
	switch l := len(matchingPods); l {
	case 0:
		err = errors.New("no matching pod has been found on the given node")
		return
	case 1:
		pod = matchingPods[0]
		return
	default:
		err = errors.New("multiple pods were found on the given node. Please retry by specifying the pod name")
		return
	}
}

func getContainerInfo(pod v1.Pod) (id string, runtime string, err error) {
	containers := pod.Status.ContainerStatuses
	if len(containers) > 1 && containerName == "" {
		err = errors.New("this pod has more than one container running. Specify container with -c")
		return
	}
	var containerID string
	if len(containers) == 1 {
		containerID = containers[0].ContainerID
	} else {
		for _, c := range containers {
			if c.Name == containerName {
				containerID = c.ContainerID
				break
			}
		}
	}
	if containerID == "" {
		err = errors.New("could not find a matching container")
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
