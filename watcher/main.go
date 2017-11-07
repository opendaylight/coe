package main

import (
	_ "git.opendaylight.org/gerrit/p/coe.git/watcher/backends/odl"
	_ "git.opendaylight.org/gerrit/p/coe.git/watcher/backends/std"

	commands "git.opendaylight.org/gerrit/p/coe.git/watcher/cmd"
)

func main() {
	commands.Execute()
}
