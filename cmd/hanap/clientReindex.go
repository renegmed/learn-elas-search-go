package main

import (
	"fmt"

	"github.com/renegmed/learn-elas-search-go/cmd/hanap/util"

	"github.com/spf13/cobra"
)

// hanap client reindex file -f ./index_file_go.csv -i golang -s .go
var clientReindexCmd = &cobra.Command{
	Use:   "reindex", // this is a sub-command
	Short: "Client service reindex",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		f, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Println(err)
			return
		}

		indx, err := cmd.Flags().GetString("index")
		if err != nil {
			fmt.Println(err)
			return
		}

		suffix, err := cmd.Flags().GetString("suffix")
		if err != nil {
			fmt.Println(err)
			return
		}

		searcher, err := util.NewSearcher()
		if err != nil {
			fmt.Println(err)
		}

		//result, err := searcher.Reindex(f, ".go", "golang")
		result, err := searcher.Reindex(f, suffix, indx)
		if err != nil {
			fmt.Printf("Error while reindex %s, %v\n", indx, err)
		} else {
			fmt.Printf("RESULT of reindexing, %s\n", result)
		}

	},
}

func init() {
	clientCmd.AddCommand(clientReindexCmd)
	clientReindexCmd.Flags().StringP("file", "f", "", "csv file for indexing")
	clientReindexCmd.Flags().StringP("index", "i", "", "elasticsearch index e.g. golang, javascript, rust, solidity")
	clientReindexCmd.Flags().StringP("suffix", "s", "", "file suffix e.g. .go, .js, .rs, .sol")

}
