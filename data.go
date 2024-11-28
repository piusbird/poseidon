package main

var version = "[GITREV]: Netscape Wizardry "
var UserAgents = map[string]string{
	"Desktop": "Mozilla/5.0 (X11; Linux x86_64; rv:108.0) Gecko/20100101 Firefox/108.0",
	"IPhone":  "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148",
	"XBox":    "Mozilla/5.0 (Windows NT 10.0; Win64; x64; Xbox; Xbox One) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edge/44.18363.8131",
}

var maxBodySize int64 = 2 * 1024 * 1024

type ParserType bool

const (
	NUParser  ParserType = true
	ArcParser ParserType = false
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
    <base href="/">
    <title>{{article.Title}}</title>
    <link rel="stylesheet" href="/assets/style.css" />
    <script type="module" crossorigin="anonymous" src="/assets/js/index.js"></script> 
  </head>
  <body id="top">
  <main>
  <div class="content">
  <h1>{{article.Title}} </h1> <br/>
  <h3> {{article.Byline}} </h3> <br/>
  
  
  <img src="{{article.Image}}">  Article Images </img>
  <hr/>
  {{article.Content | safe }}
  
  <hr/>
  <footer> <b> <a href="{{url}}"> Original Source </a> - 
  <a href={{switchurl}}> Switch Engines </a> </footer>
  </div>
  <div class="nav-bar">
        <button class="nav-button" onclick="lineHeightAdust(0.1)">LineSP +</button>
        <button class="nav-button" onclick="lineHeightAdust(-0.1)">LineSp -</button>
        <button class="nav-button"  onclick="fontSizeAdjust(2)">Font Size +</button>
        <button class="nav-button" onclick="fontSizeAdjust(-2)">Font Size -</button>
    </div>

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
	Text    string
}
