package main

import "fmt"
import "net/http"
import "net/http/httputil"
import "net/url"
import "time"
import "log"
import "text/template"
import "io/ioutil"
import "io"
import "flag"
import "os"
import "github.com/elazarl/goproxy"


func Log(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
        handler.ServeHTTP(w, r)
    })
}

func IndexPage(wr io.Writer) {
    template_bytes, err := ioutil.ReadFile("templates/index.html")
    if err != nil {
        log.Fatal(err)
    }
    template_string := string(template_bytes[:])

    template_renderer, err := template.New("index").Parse(template_string)
    if err != nil {
        log.Fatal(err)
    }
    template_renderer.Execute(wr, nil)
}

var record = flag.Bool("record", false, "run a http proxy that will record requests")
var playback = flag.String("playback", "", "playback a recoreded session")
var target = flag.String("target", "", "the target host to record")

func getTimeMilis() int64 {
    return time.Now().UnixNano()%1e6/1e3
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

            r.RequestURI = ""
            r.URL        = request_url
            r.URL.Scheme = "https"
            r.Host       = *target
            r.URL.Host   = *target
            r.URL.User   = nil

            var requestBytes []byte
            requestBytes, err = httputil.DumpRequest(r, true)
            currentTimeMilis := getTimeMilis()
            fo.WriteString(fmt.Sprint(currentTimeMilis-startTimeMilis))
            fo.Write([]byte{0xff,0xff})
            fo.Write(requestBytes)
            fo.Write([]byte{0xff,0xff,0xff})
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
}
