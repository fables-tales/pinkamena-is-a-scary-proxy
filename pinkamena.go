package main

import "fmt"
import "net/http"
import "net/http/httputil"
import "net/url"
import "time"
import "log"
import "encoding/base64"
import "flag"
import "os"
import "github.com/elazarl/goproxy"


func Log(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
        handler.ServeHTTP(w, r)
    })
}

var record = flag.Bool("record", false, "run a http proxy that will record requests")
var playback = flag.String("playback", "", "playback a recoreded session")
var target = flag.String("target", "", "the target host to record")

func getTimeMilis() int64 {
    return time.Now().UnixNano()/1e6
}

func runProxy() {
    proxy := goproxy.NewProxyHttpServer()
    proxy.Verbose = true
    client := http.Client{}
    fo, err := os.Create(".requests")
    if err != nil { panic(err) }
    // close fo on exit and check for its returned error
    defer func() {
        if err := fo.Close(); err != nil {
            panic(err)
        }
    }()

    startTimeMilis := getTimeMilis()

    proxy.OnRequest().DoFunc(func(r *http.Request,ctx *goproxy.ProxyCtx)(*http.Request,*http.Response) {
        fmt.Println(r.Host)
        request_url, err := url.Parse(r.RequestURI)
        if err != nil {
            panic(err)
        }

        log.Print("Rewriting request on ", r.Host, " to ", *target)

        r.RequestURI = ""
        r.URL        = request_url
        r.URL.Scheme = "https"
        r.Host       = *target
        r.URL.Host   = *target
        r.URL.User   = nil

        currentTimeMilis := getTimeMilis()
        fo.WriteString(fmt.Sprint(currentTimeMilis-startTimeMilis))
        fo.WriteString("\nLOLPONIES\n")

        var requestBytes []byte
        requestBytes, err = httputil.DumpRequest(r, true)
        fo.WriteString(base64.StdEncoding.EncodeToString(requestBytes))

        fo.WriteString("\nLOLPONIES\n")
        fo.Sync()

        resp, err := client.Do(r)
        if err != nil {
            panic(err)
        }

        return nil, resp
    })
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
    fmt.Println("Serving on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", proxy))
}

func main() {
    flag.Parse()
    if (*playback == "" && *record == false) {
        fmt.Println("One of --playback <filename> or --record is required")
        flag.Usage()
        os.Exit(1)
    } else if (*record && *target == "") {
        fmt.Println("--target <url> is required with --record")
        flag.Usage()
        os.Exit(1)
    } else if (*record && *target != "") {
        runProxy()
    }
}
