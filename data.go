package main

var version = "10-CURRENT"
var UserAgents = map[string]string{
	"Desktop":          "Mozilla/5.0 (X11; Linux x86_64; rv:108.0) Gecko/20100101 Firefox/108.0",
	"Googlebot Mobile": "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html",
	"Twitter":          "Mozilla/5.0 (compatible; Twitterbot/1.0)as",
	"IPhone":           "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148",
}

var maxBodySize int64 = 2 * 1024 * 1024

type ParserType bool

const (
	NUParser  ParserType = false
	ArcParser ParserType = true
)

var default_agent = UserAgents["IPhone"]

type OurCookie struct {
	UserAgent string
	Parser    ParserType
}
type ImgData struct {
	Name  string
	Image string
}

var rateMax = 6
var rateBurst = 3

var defaultCookie = OurCookie{default_agent, ArcParser}
var Header = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="ie=edge" />
    <title>{{article.Title}}</title>
    <link rel="stylesheet" href="/assets/style.css" />
  </head>
  <body id="top">
  <main>
  <h1>{{article.Title}} </h1> <br/>
  <h3> {{article.Byline}} </h3> <br/>
  <a href={{switchurl}}> Switch Engines </a>
  
  <img src="{{article.Image}}">  Article Images </img>
  <hr/>
  {{article.Content | safe }}
  <hr/>
  <footer> <b> <a href="{{url}}"> Original Source </a> -  </footer>
  </main> 
  </body>
  </html>`

var ImageProlog = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="ie=edge" />
    <title>{{imgdata.Name}}</title>
    <link rel="stylesheet" href="/assets/style.css" />
  </head>
  <body id="top">
  <main>
  <img src="{{imgdata.Image}}">  {{imgdata.Name}} </img>
  </main> 
  </body>
  </html>
  
  `
var SupportedImagesTypes = [...]string{"image/jpeg", "image/png", "image/gif"}
var YTProxy = "vid.puffyan.us"

type GenaricArticle struct {
	Title   string
	Byline  string
	Content string
	Image   string
	Length  int
}
