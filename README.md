# Poseidon 

This is a web proxy which grabs the text of an article on the web,
and refines it to something, more conducive to actually reading 
whatever it is that you're trying to read. Also it'll be 
a whole lot smaller at least on the client end.

This also has some paywall busting features, but this is not 
the primary goal of the project. Use Responsibly 

# Why?

Because the web is a bloated ugly mess, with accessibility issues and 
just ARGH. I get so ANGY trying to explain this. Basically rare disability. 
Locks me into Firefox i no like Firefox. Therefore i make accessibility 
features exist in not Firefox so I can yeet Firefox in near future.

Also this works better on my older/low memory systems

# This is not a security tool

This proxies a **single** request, things like images style sheets and custom
JavaScript and other gunk will be fetched locally. 
While we do a pretty good job of filtering most of that stuff out before it
gets to you. We don't make any security promises. 

Do not use this as a privacy or anonymizing service you have been warned 

# Technical stuff
Poseidon, has two rendering engines Miniweb and MozReader. 
Miniweb will go away at some point as it's a pain, 
but for now this has a hard runtime dependency on [Miniweb Proxy](https://humungus.tedunangst.com/r/miniwebproxy)

I'm sure tedu would be horrified at the abuse of his poor code, but when i 
was knocking this out last week, I knew of no better solution. 
So Sorry I guess. 

The MozReader engine is based on a not particularly up to date port of 
Mozilla's Readability.js to golang. I am working to make it better
 
