package main

import (
	"elasticsearch-olivere/cmd/hanap/util"
	"fmt"

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

		searcher, err := util.NewSearcher()
		if err != nil {
			check(err)
		}
		fileList, err := searcher.Search(index, phrase)
		if err != nil {
			check(err)
		}
		for i, fileName := range fileList {
			fmt.Printf("FILE: %d. %s\n", i+1, fileName)
		}
	},
}

func init() {
	clientCmd.AddCommand(clientSearchCmd)
	clientSearchCmd.Flags().StringP("index", "i", "", "index group name e.g. golang, javascript, rust, solidity")
	clientSearchCmd.Flags().StringP("phrase", "p", "", "phrase to search")
}
