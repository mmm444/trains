trains
======

Track generator for IKEA LILLABO and LEGO DUPLO train sets.

My son got [Lego Duplo Deluxe Train Set](http://shop.lego.com/en-US/Deluxe-Train-Set-10508)
for Christmas and I was thinking what
tracks could be built from the parts in the box. This program finds them
all by using a brute force search with some primitive branch pruning.

It was built as faster version of the program presented by John Graham-Cumming
in his [blog post](http://blog.jgc.org/2010/01/more-fun-with-toys-ikea-lillabo-train.html).


Install
-------
```
go get -u github.com/mmm444/trains
```

Run
---
To generate all the tracks for Lego train set run
```
trains
``` 
and it will output some HTML and SVG to the current directory. On my machine it takes about 20s
to find all 2036 tracks.

If you want to genrate all the [IKEA LILLABO](http://www.ikea.com/us/en/catalog/products/30064359/)
tracks run
```
trains -ikea -c 12 -s 2 -b 1
```

TODO
----
- [ ] eliminate more symmetrical tracks from the result
- [ ] support switches from the [Duplo Train Acccessory Set](http://shop.lego.com/en-US/Train-Accessory-Set-10506)
- [ ] beautify the code
- [ ] beautify the output
