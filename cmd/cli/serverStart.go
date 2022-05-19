package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/renegmed/learn-elas-search-go/pkg/utils"
	"github.com/spf13/cobra"
)

var serverStartCmd = &cobra.Command{
	Use:   "start", // this is a sub-command
	Short: "Server service: start elasticsearch server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Start elasticsearch server")
		if err := startElasticSearchServer(args); err != nil {
			log.Fatalln("Start Server Failed:", utils.Error(err))
		}
	},
}

func startElasticSearchServer(keywords []string) error {
	cmd := exec.Command("/Users/rene/System/elasticsearch-6.6.0/bin/elasticsearch")
	//cmd := exec.Command("elasticsearch")
	err := cmd.Run()
	if err != nil {
		utils.Error(err)
	}
	return err
}

func init() {
	serverCmd.AddCommand(serverStartCmd)
}
