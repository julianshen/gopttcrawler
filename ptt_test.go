package gopttcrawler

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
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

func TestIterator(t *testing.T) {
	assert := assert.New(t)
	n := 100

	articles, e := GetArticles("movie", 0)
	assert.Nil(e)
	iterator := articles.Iterator()

	i := 0
	for {
		if article, e := iterator.Next();e == nil {
			if i >= n {
				break
			}
			i++

			log.Printf("%v %v", i, article)
		}
	}
}
