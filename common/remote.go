package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RemoteCmd wraps common parameters for running a command on a remote host
type RemoteCmd struct {
	Command         []string
	InteractiveFlag string
	NodeName        string
	Sudo            string
}

func buildArgs(parts ...[]string) (args []string) {
	for _, part := range parts {
		args = append(args, part...)
	}
	return
}

func (rc *RemoteCmd) getBaseCmd(isInteractive bool) (args []string) {
	extraArgs := []string{rc.InteractiveFlag, rc.NodeName}
	if !isInteractive || rc.InteractiveFlag == "" {
		extraArgs = extraArgs[1:]
	}
	if rc.Sudo != "" {
		extraArgs = append(extraArgs, rc.Sudo)
	}
	return buildArgs(rc.Command, extraArgs)
}

func (rc *RemoteCmd) run(name string, arg ...string) ([]byte, []byte, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command(name, arg...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func (rc *RemoteCmd) getDockerPid(containerID, runtime string, isInteractive bool) (pid string, err error) {
	cmd := buildArgs(rc.getBaseCmd(isInteractive), []string{"docker", "inspect", "--format", "'{{ .State.Pid }}'", containerID})
	out, _, err := rc.run(cmd[0], cmd[1:]...)
	if err != nil {
		return
	}
	pid = string(out)
	return
}

func (rc *RemoteCmd) getContainerdPid(containerID, runtime string, isInteractive bool) (pid string, err error) {
	cmd := buildArgs(rc.getBaseCmd(isInteractive), []string{"crictl", "inspect", containerID})
	out, _, err := rc.run(cmd[0], cmd[1:]...)
	if err != nil {
		return
	}
	var pidInfo struct {
		Info struct {
			Pid json.Number `json:"pid"`
		} `json:"info"`
	}
	err = json.Unmarshal(out, &pidInfo)
	if err != nil {
		return
	}
	pid = pidInfo.Info.Pid.String()
	return
}

// GetPid retrieves the PID for a given container ID on the specified runtime
func (rc *RemoteCmd) GetPid(containerID, runtime string) (pid string, err error) {
	if runtime == "docker" {
		return rc.getDockerPid(containerID, runtime, false)
	}
	if runtime == "containerd" {
		return rc.getContainerdPid(containerID, runtime, false)
	}
	err = errors.New("unsupported container runtime")
	return
}

// Enter executes nsenter on the remote host
func (rc *RemoteCmd) Enter(pid, nsArgs string, command []string) error {
	cmd := buildArgs(rc.getBaseCmd(true), []string{"nsenter", "-t", pid, nsArgs})
	if len(command) != 0 {
		cmd = append(cmd, "sh", "-c", strings.Join(command, " "))
	}
	nsenterCmd := exec.Command(cmd[0], cmd[1:]...)
	nsenterCmd.Stdout = os.Stdout
	nsenterCmd.Stdin = os.Stdin
	nsenterCmd.Stderr = os.Stderr
	return nsenterCmd.Run()
}
