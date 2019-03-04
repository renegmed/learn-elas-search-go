package web

import (
	"bytes"
	"elasticsearch-olivere/cmd/hanap/util"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	pdf "github.com/ledongthuc/pdf"
)

type content struct {
	ItemNum int
	Phrase  string
	Index   string
	Source  string
	Content string
	IsWeb   bool
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
			var isWeb = false
			var fileContent string

			if strings.HasSuffix(filePath, ".pdf") {
				fileContent, err = readPdf(filePath)
			} else if strings.HasPrefix(filePath, "http") {
				isWeb = true
				fileContent, err = readHtml(filePath)
			} else {
				b, err := ioutil.ReadFile(filePath)
				if err != nil {
					fmt.Printf("%v\n", err)
					panic(err)
				}
				fileContent = string(b)
			}

			if err != nil {
				fmt.Printf("%v\n", err)
				panic(err)
			}
			_content := content{
				ItemNum: idx + 1,
				Phrase:  phrase,
				Index:   index,
				Source:  filePath,
				Content: fileContent,
				IsWeb:   isWeb,
			}

			// fmt.Printf(" content.content: %s\n", _content.Content)
			contents = append(contents, _content)
		}

		c.HTML(http.StatusOK, "index.html",
			gin.H{"SearchContents": contents})

	})
	return r
}

func readHtml(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		//log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		return "", fmt.Errorf("status code error: %d %s %s", res.StatusCode, res.Status, url)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var textBuilder bytes.Buffer

	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		content := s.Text()
		//title := s.Find("i").Text()
		//fmt.Printf("Title %d: %s \n", i, content)
		textBuilder.WriteString(content + "\n")
	})

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		content := s.Text()
		//title := s.Find("i").Text()
		//fmt.Printf("Line %d: %s \n", i, content)
		textBuilder.WriteString(content + "\n")
	})

	return textBuilder.String(), nil
}

func readPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	defer f.Close()

	if err != nil {
		return "", err
	}
	totalPage := r.NumPage()

	var textBuilder bytes.Buffer

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		var fonts = map[string]*pdf.Font{}
		for _, font := range p.Fonts() {
			f := p.Font(font)
			//fmt.Printf("%s - %v", font, f)
			fonts[font] = &f
		}

		pdfPageContent, err := p.GetPlainText(fonts)
		if err != nil {
			continue
		}
		fmt.Printf("-------------------------------------------\n")
		fmt.Printf("%s\n", pdfPageContent)

		textBuilder.WriteString("-------------------------------------------\n")
		textBuilder.WriteString(pdfPageContent + "\n")

		// texts := p.Content().Text

		// for _, text := range texts {
		// 	// fmt.Printf("x=%06.2f y=%06.2f w=%06.2f %q %s %.1fpt\n",
		// 	// 	text.X, text.Y, text.W, text.S, text.Font,
		// 	// 	text.FontSize)

		// 	textBuilder.WriteString(text.S)
		// }

		// textBuilder.WriteString(p.GetPlainText(""))
	}
	return textBuilder.String(), nil
	//return "", nil
}
