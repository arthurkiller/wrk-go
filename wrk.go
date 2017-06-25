package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/http2"

	"github.com/arthurkiller/perfm"
)

func ErrLog(args ...interface{}) {
	var errargs []interface{} = []interface{}{"[E]"}
	errargs = append(errargs, args...)
	log.Println(errargs...)
	os.Exit(0)
}

func main() {
	var header, host, method, payload string
	var thread, connection, delay int
	var ssl, sessionResume, falseStart, keepAlive, compression, h2 bool
	var err error

	flag.StringVar(&header, "H", "", "set for the request header")
	flag.StringVar(&host, "h", "", "set the server host")
	flag.StringVar(&method, "X", "GET", "request method")
	flag.StringVar(&payload, "D", "", "request payload")
	flag.IntVar(&thread, "t", 1, "set for the concurrence when benchmark run")
	flag.IntVar(&connection, "c", 1, "set for the connection for each worker")
	flag.IntVar(&delay, "d", 10, "set for the benchmark delay")

	flag.BoolVar(&keepAlive, "keepalive", false, "enable the ssl for http")
	flag.BoolVar(&compression, "compress", true, "enable the ssl for http")
	flag.BoolVar(&h2, "http2", false, "enable the ssl for http")
	flag.BoolVar(&ssl, "https", true, "enable the ssl for http")
	flag.BoolVar(&sessionResume, "session", true, "enable the session resumption in ssl handshake")
	flag.BoolVar(&falseStart, "alpn", true, "enable ssl false start via alpn")
	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		os.Exit(0)
	}

	if connection < 0 || thread < 0 || delay < 0 || host == "" {
		ErrLog("argument invalid")
	}

	if !strings.Contains(host, "://") {
		ErrLog("host contains no protocol")
	}

	var req *http.Request
	req, err = http.NewRequest(method, host, strings.NewReader(payload))
	if err != nil {
		ErrLog("request error", err)
	}

	if header != "" {
		headers := strings.Split(header, ",")
		for _, v := range headers {
			kvs := strings.Split(v, ":")
			if len(kvs) == 2 {
				req.Header.Add(kvs[0], kvs[1])
				continue
			}
			ErrLog("header format invalid", kvs)
			return
		}
	}

	var tlscfg *tls.Config
	if ssl {
		tlscfg = &tls.Config{
			ServerName:         strings.Split(host, "://")[1],
			ClientSessionCache: tls.NewLRUClientSessionCache(4096),
			Renegotiation:      tls.RenegotiateFreelyAsClient,
		}
		if falseStart {
			tlscfg.NextProtos = []string{"h2", "http/1.1"} // http 2 support for alpn
		}
		if !sessionResume {
			tlscfg.SessionTicketsDisabled = true
		}
	}

	var tr *http.Transport = &http.Transport{
		TLSClientConfig: tlscfg,
	}
	if !keepAlive {
		tr.DisableKeepAlives = true
	}
	if !compression {
		tr.DisableCompression = true
	}

	tr2 := &http2.Transport{
		TLSClientConfig: tlscfg,
	}

	var client *http.Client
	if h2 {
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{Transport: tr2}
	}

	perf := perfm.New(perfm.WithBinsNumber(15))

	perf.Registe(func() (err error) {
		_, err = client.Do(req)
		return
	})

	perf.Start()
	perf.Wait()
}
