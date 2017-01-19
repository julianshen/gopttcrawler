package gopttcrawler

import (
	"errors"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	BASE_URL = "https://www.ptt.cc/bbs/"
)

var (
	ERROR_EMPTY_LIST = errors.New("list is empty")
)

type Article struct {
	ID       string `json:"id"`    //Article ID
	Board    string `json:"board"` //Board name
	Title    string `json:"title"`
	Content  string `json:"content"`
	Author   string `json:"author"` //Author ID
	DateTime string `json:"date_time"`
	Nrec     int    `json:"nrec"` //推文數(推-噓)
	Url      string `json:"url"`
	doc      *goquery.Document
}

type ArticleList struct {
	Articles     []*Article //Articles
	Board        string     //Board
	PreviousPage int        //Previous page id
	NextPage     int        //Next page id
}

type Iterator func() (*Article, error)

func (i Iterator) Next() (*Article, error) {
	return i()
}

func newDocument(url string) (*goquery.Document, error) {
	// Load the URL
	req, e := http.NewRequest("GET", url, nil)
	if e != nil {
		return nil, e
	}

	cookie := http.Cookie{
		Name:  "over18",
		Value: "1",
	}
	req.AddCookie(&cookie)

	res, e := http.DefaultClient.Do(req)

	if e != nil {
		return nil, e
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}

	return goquery.NewDocumentFromResponse(res)
}

func getPage(prefix, text string) int {
	re := regexp.MustCompile(prefix + "/index(\\d+).html$")
	matched := re.FindStringSubmatch(text)

	if len(matched) > 1 {
		ret, _ := strconv.Atoi(matched[1])
		return ret
	}

	return 0
}

func GetArticles(board string, page int) (*ArticleList, error) {
	index := "/index.html"
	if page != 0 {
		index = "/index" + strconv.Itoa(page) + ".html"
	}

	url := BASE_URL + board + index
	doc, err := newDocument(url)

	if err != nil {
		return nil, err
	}

	articleList := &ArticleList{PreviousPage: 0, NextPage: 0, Board: board}

	prevPageSel := doc.Find(".action-bar").Find("a:contains('上頁')")
	if len(prevPageSel.Nodes) > 0 {
		href, _ := prevPageSel.Attr("href")
		articleList.PreviousPage = getPage("/bbs/"+board, href)
	}
	nextPageSel := doc.Find(".action-bar").Find("a:contains('下頁')")
	if len(nextPageSel.Nodes) > 0 {
		href, _ := nextPageSel.Attr("href")
		articleList.NextPage = getPage("/bbs/"+board, href)
	}

	articles := make([]*Article, 0)
	stop := false
	doc.Find(".r-ent").Each(func(i int, s *goquery.Selection) {
		//過濾掉置底文章
		if class, found := s.Prev().Attr("class"); found && class == "r-list-sep" {
			stop = true
		}

		article := &Article{Board: board}
		//Nrec
		nrecSel := s.Find(".nrec")
		if len(nrecSel.Nodes) > 0 {
			nrecStr := nrecSel.Text()

			if nrecStr == "爆" {
				article.Nrec = math.MaxInt32
			} else {
				article.Nrec, _ = strconv.Atoi(nrecStr)
			}
		}
		//DateTime
		DateTimeSel := s.Find(".date")
		if len(DateTimeSel.Nodes) > 0 {
			article.DateTime = strings.TrimSpace(DateTimeSel.Text())
		}
		//Author
		authorSel := s.Find(".author")
		if len(authorSel.Nodes) > 0 {
			article.Author = authorSel.Text()
		}
		//Title
		linkSel := s.Find(".title > a")
		if linkSel.Size() != 0 {
			href, existed := linkSel.Attr("href")
			if existed {
				re := regexp.MustCompile("/bbs/" + board + "/(.*).html$")
				matchedID := re.FindStringSubmatch(href)
				if matchedID != nil && len(matchedID) > 1 {
					article.ID = matchedID[1]
					article.Title = strings.TrimSpace(linkSel.Text())
					article.Url = BASE_URL + article.Board + "/" + article.ID + ".html"

					if !stop {
						articles = append(articles, article)
					}
				}
			}
		}
	})

	articleList.Articles = articles

	return articleList, nil
}

func GetArticlesGo(board string, page int) (<-chan *Article, chan bool) {
	ch := make(chan *Article)
	notifyCh := make(chan bool)

	go func() {
		defer close(ch)
		defer close(notifyCh)

		articles, err := GetArticles(board, page)

		if err != nil {
			log.Println(err)
			close(ch)
		}

		i := articles.Iterator()
		for {
			if article, e := i.Next(); e != nil {
				log.Println(err)
				close(ch)
				break
			} else {
				select {
				case ch <- article:
				case <-notifyCh:
					log.Println("get notified")
					break
				}
			}
		}
	}()

	return ch, notifyCh
}

func LoadArticle(board, id string) (*Article, error) {
	url := BASE_URL + board + "/" + id + ".html"
	doc, err := newDocument(url)

	if err != nil {
		return nil, err
	}

	return loadArticle(doc, board, id)
}

func loadArticle(doc *goquery.Document, board, id string) (*Article, error) {
	article := &Article{ID: id, Board: board}

	//Get title
	article.Title = strings.TrimSpace(doc.Find("title").Text())
	//Get Content
	meta := doc.Find(".article-metaline")
	meta.Find(".article-meta-value").Each(func(i int, s *goquery.Selection) {
		switch i {
		case 0: //Author
			name := s.Text()
			re := regexp.MustCompile("^(.*)\\s+\\(.*\\)")
			matched := re.FindStringSubmatch(name)

			if matched != nil && len(matched) > 1 {
				name = matched[1]
			}
			article.Author = name
		case 2: //Time
			article.DateTime = strings.TrimSpace(s.Text())
		}
	})

	meta.Remove() //Remove header

	//Remove board name
	metaRight := doc.Find(".article-metaline-right")
	metaRight.Remove()

	push := doc.Find(".push")
	//Count push
	pushCnt := push.Find(".push-tag:contains('推')").Size()
	booCnt := push.Find(".push-tag:contains('噓')").Size()
	article.Nrec = pushCnt - booCnt

	if article.Nrec < 0 {
		article.Nrec = 0
	}
	push.Remove()

	sel := doc.Find("#main-content")
	article.Content, _ = sel.Html()

	article.Url = BASE_URL + article.Board + "/" + article.ID + ".html"

	return article, nil
}

func (aList *ArticleList) GetFromPreviousPage() (*ArticleList, error) {
	newList, err := GetArticles(aList.Board, aList.PreviousPage)

	if err != nil {
		return aList, nil
	}

	*aList = *newList
	return aList, nil
}

func (aList *ArticleList) GetFromNextPage() (*ArticleList, error) {
	newList, err := GetArticles(aList.Board, aList.NextPage)

	if err != nil {
		return aList, nil
	}

	*aList = *newList
	return aList, nil
}

func (aList *ArticleList) Iterator() Iterator {
	index := 0

	return func() (*Article, error) {
		if index >= len(aList.Articles) && index != 0 {
			if _, err := aList.GetFromPreviousPage(); err != nil {
				return nil, err
			}
			index = 0
		}

		if aList.Articles == nil || len(aList.Articles) == 0 {
			return nil, ERROR_EMPTY_LIST
		}

		article := aList.Articles[index]
		index++
		return article, nil
	}
}

func (a *Article) Load() *Article {
	url := a.Url
	doc, err := newDocument(url)

	if err != nil {
		return a
	}

	a.doc = doc

	newA, err := loadArticle(doc, a.Board, a.ID)
	if err == nil {
		*a = *newA
	}
	return a
}

func (a *Article) GetImageUrls() ([]string, error) {
	doc := a.doc
	if doc == nil {
		var err error
		doc, err = newDocument(a.Url)

		if err != nil {
			return nil, err
		}
		a.doc = doc
	}

	result := make([]string, 0)
	imgs := doc.Find("#main-content").Find("img")
	imgs.Each(func(i int, s *goquery.Selection) {
		src := s.AttrOr("src", "")
		if src != "" {
			result = append(result, src)
		}
	})
	return result, nil
}

func (a *Article) GetLinks() ([]string, error) {
	doc := a.doc

	if doc == nil {
		doc, err := newDocument(a.Url)

		if err != nil {
			return nil, err
		}
		a.doc = doc
	}

	result := make([]string, 0)
	links := doc.Find("#main-content").Find("a")
	links.Each(func(i int, s *goquery.Selection) {
		src := s.AttrOr("href", "")
		if src != "" {
			result = append(result, src)
		}
	})
	return result, nil
}
