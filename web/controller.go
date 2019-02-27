package web

import (
	"elasticsearch-olivere/cmd/hanap/util"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type content struct {
	ItemNum int
	Phrase  string
	Index   string
	Source  string
	Content string
}

func RegisterRoutes() *gin.Engine {

	r := gin.Default()

	r.LoadHTMLGlob("templates/**/*.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/", func(c *gin.Context) {
		//fmt.Println("+++++ Search Now ++++")

		index := c.PostForm("index")
		phrase := c.PostForm("phrase")

		fmt.Printf(" index: %s  phrase: %s\n", index, phrase)

		searcher, err := util.NewSearcher()
		if err != nil {
			c.HTML(http.StatusOK, "error.html", nil)
		}
		files, err := searcher.Search(index, phrase)

		//files, err := util.SearchWithReturn(index, phrase)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", nil)
		}

		//fmt.Println(files)

		contents := []content{}

		for idx, filePath := range files {

			byteContents, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Printf("%v\n", err)
				panic(err)
			}

			_content := content{
				ItemNum: idx + 1,
				Phrase:  phrase,
				Index:   index,
				Source:  filePath,
				Content: string(byteContents),
			}

			// fmt.Printf(" content.content: %s\n", _content.Content)
			contents = append(contents, _content)
		}

		c.HTML(http.StatusOK, "index.html",
			gin.H{"SearchContents": contents})

	})
	return r
}
