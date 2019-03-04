// file name: metadata.go
package main

import (
	// import standard libraries

	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	// import third party libraries
)

func metaScrape() {
	// doc, err := goquery.NewDocument("https://raywangblog.wordpress.com/2017/07/16/golang-profiling/")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	res, err := http.Get("https://yos.io/2018/02/08/getting-started-with-serverless-go/") //"https://raywangblog.wordpress.com/2017/07/16/golang-profiling/")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// ok := doc.Find("div").IsFunction(func(i int, s *Selection) bool {
	// 	return s.HasClass("container-fluid")
	// })

	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		band := s.Text()
		//title := s.Find("i").Text()
		fmt.Printf("Review %d: %s \n", i, band)
	})

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		band := s.Text()
		//title := s.Find("i").Text()
		fmt.Printf("Review %d: %s \n", i, band)
	})

	// doc.Find("div").Each(func(i int, s *goquery.Selection) {
	// 	// For each item found, get the band and title
	// 	band := s.Find("p").Text()
	// 	//title := s.Find("i").Text()
	// 	fmt.Printf("Review %d: %s \n", i, band)
	// })

	// Find the review items
	// doc.Find(".sidebar-reviews article .content-block").Each(func(i int, s *goquery.Selection) {
	// 	// For each item found, get the band and title
	// 	band := s.Find("a").Text()
	// 	title := s.Find("i").Text()
	// 	fmt.Printf("Review %d: %s - %s\n", i, band, title)
	// })

	// var metaDescription string
	// var pageTitle string

	// // use CSS selector found with the browser inspector
	// // for each, use index and item
	// pageTitle = doc.Find("title").Contents().Text()

	// doc.Find("meta").Each(func(index int, item *goquery.Selection) {
	// 	if item.AttrOr("name", "") == "description" {
	// 		metaDescription = item.AttrOr("content", "")
	// 	}
	// })
	// fmt.Printf("Page Title: '%s'\n", pageTitle)
	// fmt.Printf("Meta Description: '%s'\n", metaDescription)
}

func main() {
	metaScrape()
}
