package odl

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	commands "git.opendaylight.org/gerrit/p/coe.git/watcher/cmd"
)

var OdlCmd = &cobra.Command{
	Use:   "odl",
	Short: "odl watcher",
	Long:  "Watches kubernetes and transfers relevant information to OpenDaylight's COE engine",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Run ODL watcher")
		viper.ReadInConfig()

		var err error

		host := viper.GetString("odl.host")
		if host == "" {
			host, err = cmd.Flags().GetString("host")
			if err != nil {
				log.Panic(err)
			}
		}

		username := viper.GetString("odl.user")
		if username == "" {
			username, err = cmd.Flags().GetString("username")
			if err != nil {
				log.Panic(err)
			}
		}

		password := viper.GetString("odl.password")
		if password == "" {
			password, err = cmd.Flags().GetString("password")
			if err != nil {
				log.Panic(err)
			}
		}
		backend := New(host, username, password)

		Watch(commands.Config.Clientset, backend)
	},
}

func init() {
	OdlCmd.Flags().String("host", "http://127.0.0.1:8181", "ODL Server to connect to")
	OdlCmd.Flags().String("username", "admin", "ODL Username")
	OdlCmd.Flags().String("password", "admin", "ODL Password")
	commands.RootCmd.AddCommand(OdlCmd)
}
