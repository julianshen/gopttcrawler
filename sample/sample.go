package main

import (
	"github.com/julianshen/gopttcrawler"
	"log"
)

func main() {
	alist, _ := gopttcrawler.GetArticles("Beauty", 0)

	for _, a := range alist.Articles {
		log.Println(a.Title)
	}

	nextpage, err := alist.GetFromPreviousPage()
	if err != nil {
		panic(err)
	}

	for _, a := range nextpage.Articles {
		a.Load()
		log.Println(a.Content)
		log.Println(a.GetImageUrls())
	}
}
