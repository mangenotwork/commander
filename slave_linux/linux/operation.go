package linux

import (
	"gitee.com/mangenotework/commander/common/cmd"
	"strings"
)

func Kill(pid string, arg ...string) string {
	argStr := strings.Join(arg, " ")
	return cmd.LinuxSendCommand("kill "+ argStr+ " "+ pid)
}