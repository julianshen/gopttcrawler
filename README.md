gopttcrawler - 用來爬ptt文章的go library
========

使用方法請參考sample.go
#### 資料結構
```go
type Article struct {
	ID       string //Article ID
	Board    string //Board name
	Title    string
	Content  string
	Author   string //Author ID
	DateTime string
	Nrec     int //推文數(推-噓)
}

type ArticleList struct {
	Articles     []*Article //Articles
	Board        string //Board
	PreviousPage int //Previous page id
	NextPage     int //Next page id
}
```

#### 載入文章列表
1. 載入最新一頁表特版文章
```go
    articleList, _ := gopttcrawler.GetArticles("Beauty", 0)
    // the 1st parameter is the board name
    // the 2nd parameter is the page id. 0 indicates the latest page
```
2. 載入前一頁文章列表
```go
    prevArticleList, _ := articleList.GetFromPreviousPage()
```

#### 載入文章內容
1. 載入單篇文章詳細內容
```go
    article := articleList.Articles[0]
    article.Load()
    fmt.Println(article.Content) //印出內文(html)
```
2. 取得文章中所有圖片連結
```go
    images := article.GetImageUrls()
```
3. 取得文章中的連結
```go
    links := article.GetLinks()
```

#### Iterator

新增Iterator功能:

```go
	n := 100

	articles, e := gopttcrawler.GetArticles("movie", 0)
	
	if e != nil {
		....
	}

	iterator := articles.Iterator()

	i := 0
	for {
		if article, e := iterator.Next(); e == nil {
			if i >= n {
				break
			}
			i++

			log.Printf("%v %v", i, article)
		}
	}
```

上面這範例是抓取最新的100篇文章, 不用管第幾頁, 或是上一頁下一頁, 反正就一直抓

#### Go routine版本的GetArticles

```go
	ch, done := gopttcrawler.GetArticlesGo("Beauty", 0)
	n := 100
	i := 0
	for article := range ch {
		if i >= n {
			done <- true
			break
		}
		i++
		log.Printf("%v %v", i, article)
	}
```

這範例一樣也是抓一百篇, 只是抓文章的部分被放到go routine去了, 會立即回傳兩個channel,
第一個是receive only channel, 跟Iterator類似, 一次拿一篇文章, 可以用range, 第二個是一個bool channel, 拿夠了送個訊息通知他終止go routine,
如果把receive部分放到select去, 就是non blocking了, 不會被讀上一頁下一頁的IO給卡住