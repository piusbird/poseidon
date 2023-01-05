package main

var version = "Old Age Pius"
var UserAgents = map[string]string{
	"Desktop":          "Mozilla/5.0 (X11; Linux x86_64; rv:108.0) Gecko/20100101 Firefox/108.0",
	"Googlebot Mobile": "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html",
	"Twitter":          "Mozilla/5.0 (compatible; Twitterbot/1.0)as",
}
var default_agent = UserAgents["Googlebot Mobile"]
var ourProxy = "http://localhost:8090"
