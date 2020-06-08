package main

import (
	"github.com/renegmed/learn-elas-search-go/cmd/hanap/util"

	"github.com/spf13/cobra"
)

var index string

// $ hanap client destroy index -i golang
var clientDestroyCmd = &cobra.Command{
	Use:   "destroy", // this is a sub-command
	Short: "Client service: destroy index",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		index, _ := cmd.Flags().GetString("index")

		searcher, err := util.NewSearcher()
		if err != nil {
			check(err)
		}
		err = searcher.Destroy(index)
		//err := util.Destroy(index)
		check(err)
	},
}

func init() {
	clientCmd.AddCommand(clientDestroyCmd)
	clientDestroyCmd.Flags().StringVarP(&index, "index", "i", "", "index group name e.g. golang")
}
