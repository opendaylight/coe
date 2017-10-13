package main

import (
	_ "git.opendaylight.org/gerrit/p/coe.git/watcher/backends/odl"
	commands "git.opendaylight.org/gerrit/p/coe.git/watcher/cmd"
)

func main() {
	commands.Execute()
}
