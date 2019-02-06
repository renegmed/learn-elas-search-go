package main

import (
	"elasticsearch-olivere/cmd/hanap/util"

	"github.com/spf13/cobra"
)

// hanap client reindex file -f ./index_file.csv
var clientReindexCmd = &cobra.Command{
	Use:   "reindex", // this is a sub-command
	Short: "Client service reindex",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		f, _ := cmd.Flags().GetString("file")
		util.Reindex(f)
	},
}

func init() {
	clientCmd.AddCommand(clientReindexCmd)
	clientReindexCmd.Flags().StringP("file", "f", "", "csv file for indexing")

}
