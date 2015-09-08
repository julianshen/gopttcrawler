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
