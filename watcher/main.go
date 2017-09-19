package main

import (
	commands "git.opendaylight.org/gerrit/p/coe.git/watcher/cmd"
	_ "git.opendaylight.org/gerrit/p/coe.git/watcher/backends/odl"
)

func main() {
	commands.Execute()
}
