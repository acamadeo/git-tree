package utils

import "os/exec"

func RunCmd_Stdout(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	output, _ := cmd.Output()
	return string(output)
}
