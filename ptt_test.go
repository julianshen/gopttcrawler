package gopttcrawler

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"log"
)

func TestGetArticles(t *testing.T) {
	assert := assert.New(t)

	articles, e := GetArticles("Beauty", 0)
	assert.Nil(e)
	log.Println(len(articles.Articles))

	articles, e = GetArticles("Gossiping", 2)
	assert.Nil(e)
	log.Println(len(articles.Articles))

	nextpage, e := articles.GetFromPreviousPage()
	assert.Nil(e)
	log.Println(len(nextpage.Articles))

	articles, e = GetArticles("NotExisted", 1)
	assert.NotNil(e)
}

func TestLoadArticle(t *testing.T) {
	assert := assert.New(t)
	articles, e := GetArticles("Beauty", 0)
	assert.Nil(e)

	for _, a := range articles.Articles {
		oa := *a
		log.Println(a.DateTime)
		log.Println(a.Nrec)
		a.Load()
		log.Println(a.DateTime)
		log.Println(a.Nrec)
		
		assert.NotEqual(a.Content, oa.Content)
		_, e = a.GetImageUrls()
		assert.Nil(e)
	}
}