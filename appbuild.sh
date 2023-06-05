#!/bin/sh
if [ ! -n $appRoot ]
then
	export appRoot="/app"
fi
mkdir -pv $appRoot/src
cd $appRoot/src
echo $appRoot
git clone https://git.piusbird.space/poseidon.git/
cd poseidon 
go build
cp *.html $appRoot
cp -r assets $appRoot
cp poseidon $appRoot
