package std

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"git.opendaylight.org/gerrit/p/coe.git/watcher/backends"
	commands "git.opendaylight.org/gerrit/p/coe.git/watcher/cmd"
)

var Cmd = &cobra.Command{
	Use:   "std",
	Short: "std watcher",
	Long:  "Watches Kubernetes and print to stdout",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Run STD watcher")
		viper.ReadInConfig()

		backend := Backend{}

		backends.Watch(commands.Config.ClientSet, backend)
	},
}

func init() {
	commands.RootCmd.AddCommand(Cmd)
}
