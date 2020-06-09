package web

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"github.com/renegmed/learn-elas-search-go/cmd/hanap/searcher"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	pdf "github.com/ledongthuc/pdf"
)

type content struct {
	ItemNum      int
	Phrase       string
	Index        string
	Source       string
	Score        float32
	Content      string
	SearchMethod string
	IsWeb        bool
	IsPdf        bool
}

type header struct {
	Phrase       string
	IsWeb        bool
	IsPdf        bool
	IsGolang     bool
	IsRust       bool
	IsJavascript bool
	IsSolidity   bool
	IsNote       bool
	IsKubernetes bool
	IsGoPackage  bool
}

func RegisterRoutes() *gin.Engine {

	r := gin.Default()

	r.LoadHTMLGlob("templates/**/*.html")

	r.GET("/", func(c *gin.Context) {
		pdfFiles, _ := c.GetPostFormMap("pdfFile")

		fmt.Printf("pdfFile: %v \n", pdfFiles)

		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/", func(c *gin.Context) {

		var IsGolang = false
		var IsWeb = false
		var IsPdf = false
		var IsRust = false
		var IsJavascript = false
		var IsSolidity = false
		var IsNote = false
		var IsKubernetes = false
		var IsGoPackage = false

		index := c.PostForm("index")
		phrase := c.PostForm("phrase")
		searchMethod := c.PostForm("btn-search")

		fmt.Printf(" index: %s  phrase: %s sorted: %s\n", index, phrase, searchMethod)

		switch index {
		case "golang":
			IsGolang = true
		case "web":
			IsWeb = true
		case "pdf":
			IsPdf = true
		case "rust":
			IsRust = true
		case "javascript":
			IsJavascript = true
		case "solidity":
			IsSolidity = true
		case "note":
			IsSolidity = true
		case "kubernetes":
			IsKubernetes = true
		case "goPackage":
			IsGoPackage = true
		}

		searcher, err := searcher.NewSearcher()
		if err != nil {
			c.HTML(http.StatusOK, "error.html", nil)
		}
		files, err := searcher.Search(index, phrase, searchMethod)

		if err != nil {
			c.HTML(http.StatusOK, "error.html", nil)
		}

		//fmt.Println(files)

		contents := []content{}

		for idx, filePath := range files {
			var isWeb = false
			var isPdf = false
			var fileContent string

			if strings.HasSuffix(filePath, ".pdf") {
				fileContent, err = readPdf(filePath)
				isPdf = true
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
				ItemNum:      idx + 1,
				Phrase:       phrase,
				Index:        index,
				Source:       filePath,
				Content:      fileContent,
				SearchMethod: searchMethod,
				IsWeb:        isWeb,
				IsPdf:        isPdf,
			}
			contents = append(contents, _content)
		}

		// sort contents
		if strings.Contains(searchMethod, "sorted") {
			sort.Slice(contents, func(i, j int) bool {
				return contents[i].Source < contents[j].Source
			})
		}

		numberedContents := []content{}
		for i, content := range contents {
			content.ItemNum = i + 1
			numberedContents = append(numberedContents, content)
		}

		pdfFile := c.PostForm("pdfFile")
		if len(pdfFile) > 0 {
			var waitGroup sync.WaitGroup
			waitGroup.Add(1)
			go func() {
				waitGroup.Done()
				if len(pdfFile) > 0 {
					err := openPdfFile(pdfFile)
					if err != nil {
						fmt.Printf("%v\n", err)
					}
				}
			}()
			waitGroup.Wait()
		}

		//fmt.Printf("+++ calling c.HTML() len(numberedContents): %d\n", len(numberedContents))

		header := header{
			Phrase:       phrase,
			IsWeb:        IsWeb,
			IsPdf:        IsPdf,
			IsGolang:     IsGolang,
			IsRust:       IsRust,
			IsJavascript: IsJavascript,
			IsSolidity:   IsSolidity,
			IsNote:       IsNote,
			IsKubernetes: IsKubernetes,
			IsGoPackage:  IsGoPackage,
		}
		c.HTML(http.StatusOK, "index.html",
			gin.H{"SearchContents": numberedContents, "Header": header})

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
		// fmt.Printf("-------------------------------------------\n")
		// fmt.Printf("%s\n", pdfPageContent)

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

func openPdfFile(file string) error {
	cmd := exec.Command("open", "-a", "Preview", file)
	return cmd.Run()
}
