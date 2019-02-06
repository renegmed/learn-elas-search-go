package main

import (
	"elasticsearch-olivere/cmd/hanap/util"

	"github.com/spf13/cobra"
)

// hanap client search index phrase -i golang -p "hello world"
var clientSearchCmd = &cobra.Command{
	Use:   "search", // this is a sub-command
	Short: "Client service search",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		index, _ := cmd.Flags().GetString("index")
		phrase, _ := cmd.Flags().GetString("phrase")
		error := util.Search(index, phrase)
		check(error)
	},
}

func init() {
	clientCmd.AddCommand(clientSearchCmd)
	clientSearchCmd.Flags().StringP("index", "i", "", "index group name e.g. golang")
	clientSearchCmd.Flags().StringP("phrase", "p", "", "phrase to search")
}
