package common

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RemoteCmd wraps common parameters for running a command on a remote host
type RemoteCmd struct {
	Path               string
	PreArgs            []string
	PreArgsInteractive []string
}

// GetPid retrieves the PID for a given container ID on the specified runtime
func (rc *RemoteCmd) GetPid(containerID, runtime string) (pid string, err error) {
	if runtime == "docker" {
		args := append(rc.PreArgs, "docker", "inspect", "--format", "'{{ .State.Pid }}'", containerID)
		pidCmd := exec.Command(rc.Path, args...)
		out, e := pidCmd.Output()
		if e != nil {
			err = e
			return
		}
		pid = string(out)
		return
	}
	if runtime == "containerd" {
		args := append(rc.PreArgs, "crictl", "inspect", containerID)
		pidCmd := exec.Command(rc.Path, args...)
		out, e := pidCmd.Output()
		if err != nil {
			err = e
			return
		}
		var pidInfo struct {
			Info struct {
				Pid json.Number `json:"pid"`
			} `json:"info"`
		}
		e = json.Unmarshal(out, &pidInfo)
		if e != nil {
			err = e
			return
		}
		pid = pidInfo.Info.Pid.String()
		return
	}
	err = errors.New("unsupported container runtime")
	return
}

// Enter executes nsenter on the remote host
func (rc *RemoteCmd) Enter(pid, nsArgs string, command []string) error {
	args := append(rc.PreArgsInteractive, "nsenter", "-t", pid, nsArgs)
	if len(command) != 0 {
		args = append(args, "sh", "-c", strings.Join(command, " "))
	}
	nsenterCmd := exec.Command(rc.Path, args...)
	nsenterCmd.Stdout = os.Stdout
	nsenterCmd.Stdin = os.Stdin
	nsenterCmd.Stderr = os.Stderr
	return nsenterCmd.Run()
}
