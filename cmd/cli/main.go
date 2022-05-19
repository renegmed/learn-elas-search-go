package main

import (
	"github.com/spf13/cobra"
)

/*

 $ hanap --help

 $ hanap server start
 $ hanap server stop
 $ hanap server --help
 $
 $ curl -XGET 'localhost:9200/_cat/indices?v&pretty'
 $ curl -XGET 'localhost:9200/golang/_search?q=microservice&pretty'
 $ curl -XGET 'localhost:9200/golang,note/_search?pretty' -d '
  {
	  "query" : { "match_all" : {} },
	  "from" : 5,
	  "size" : 3
  }
 '
 $
 $ hanap client search index phrase -i golang -p '8080'
 $ hanap client destroy index -i golang
 $ hanap client reindex file -f ./index_file_go.csv -i golang -s .go
 $ hanap client destroy index -i gopackage
 $ hanap client reindex file -f ./index_file_go_src.csv -i gopackage -s .go
 $ hanap client destroy index -i solidity
 $ hanap client reindex file -f ./index_file_solidity.csv -i solidity -s .sol
 $ hanap client reindex file -f ./index_file_solidity_js.csv -i solidity -s .js
 $ hanap client destroy index -i rust
 $ hanap client reindex file -f ./index_file_rust.csv -i rust -s .rs
 $ hanap client destroy index -i pdf
 $ hanap client reindex file -f ./index_file_pdf.csv -i pdf -s .pdf
 $ hanap client destroy index -i web
 $ hanap client reindex file -f ./index_file_web.csv -i web -s web
 $ hanap client destroy index -i note
 $ hanap client reindex file -f ./index_file_note.csv -i note -s .txt
 $ hanap client reindex file -f ./index_file_note.csv -i note -s .md
 $ hanap client destroy index -i kubernetes
 $ hanap client reindex file -f ./index_file_kubernetes.csv -i kubernetes -s .yml
 $ hanap client reindex file -f ./index_file_kubernetes.csv -i kubernetes -s .yaml
 $
*/

// Cobra is both a library for creating powerful modern CLI
// applications as well as a program to generate applications
// and command files

var rootCmd *cobra.Command

func main() {
	rootCmd.Execute()
}

func init() {

	rootCmd = &cobra.Command{
		Use:   "hanap",
		Short: "Search application for finding application source code files",
	}
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(clientCmd)

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
