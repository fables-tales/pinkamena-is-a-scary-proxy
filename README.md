Pinkamena is a scary proxy
==========================

![Pinkamena is a scary proxy](http://images5.fanpop.com/image/photos/30900000/Pinkamena-diane-pie-cupcakes-the-movie-30989323-2296-2560.jpg)

##Example usage:

```
$ pinkamena --record --target http://localhost:5443
Serving on http://localhost:8080
```

Then browse to http://localhost:8080, which will proxy http://localhost:5443.
Do some things, then hit ctrl-c in your terminal and run:


```
$ pinkamena --playback .requests
Playing back requests from file `.requests`
```
