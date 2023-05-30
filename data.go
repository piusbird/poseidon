package main

var version = "cf30f1f"
var UserAgents = map[string]string{
	"Desktop":          "Mozilla/5.0 (X11; Linux x86_64; rv:108.0) Gecko/20100101 Firefox/108.0",
	"Googlebot Mobile": "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html",
	"Twitter":          "Mozilla/5.0 (compatible; Twitterbot/1.0)as",
	"IPhone":           "Mozilla/5.0 (iPhone; CPU iPhone OS 12_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148",
}
var default_agent = UserAgents["IPhone"]
var ourProxy = "http://localhost:8090"

type OurCookie struct {
	UserAgent   string
	Readability bool
}

var rateMax = 6
var rateBurst = 3

var defaultCookie = OurCookie{default_agent, true}
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
  <h3> {{article.Byline}} </h3>
  
  <img src="{{article.Image}}">  Article Images </img>
  <hr/>
  {{article.Content | safe }}
  <hr/>
  <footer> <b> <a href="{{url}}"> Original Source </a>  </footer>
  </main> 
  </body>
  </html>`
