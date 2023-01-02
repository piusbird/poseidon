#!/bin/sh
set -e

[ -f intercept.key ] ||
	openssl genrsa -out intercept.key 2048

[ -f intercept.csr ] ||
	openssl req -new -key intercept.key -out intercept.csr -subj /CN=intercept.miniweb

openssl x509 -sha256 -req -days 365 -in intercept.csr -out intercept.crt -signkey intercept.key
