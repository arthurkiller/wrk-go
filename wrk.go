package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/arthurkiller/perfm"

	"golang.org/x/net/http2"
)

func ErrLog(args ...interface{}) {
	var errargs []interface{} = []interface{}{"\033[31m[E]\033[0m"}
	errargs = append(errargs, args...)
	log.Println(errargs...)
	os.Exit(0)
}

func main() {
	var header, host, method, payload string
	var thread, delay, timeout int
	var ssl, sessionResume, falseStart, keepAlive, compression, h2 bool
	var err error

	flag.StringVar(&header, "H", "", "set for the request header")
	flag.StringVar(&host, "host", "", "set the server host")
	flag.StringVar(&method, "X", "GET", "request method")
	flag.StringVar(&payload, "D", "", "request payload")
	flag.IntVar(&thread, "N", 1, "set for the concurrence worker when benchmark run the job")
	flag.IntVar(&delay, "duration", 10, "set for the benchmark delay")
	flag.IntVar(&timeout, "timeout", 0, "set timeout for each request with second. Default is no timeout")

	flag.BoolVar(&keepAlive, "keepalive", false, "keep tcp connection alive")
	flag.BoolVar(&compression, "compress", false, "enable haeder compress")
	flag.BoolVar(&h2, "http2", false, "use http/2 for benchmark")
	flag.BoolVar(&ssl, "ssl", true, "enable the ssl")
	flag.BoolVar(&sessionResume, "session", true, "enable the session resumption in ssl handshake")
	flag.BoolVar(&falseStart, "alpn", false, "enable ssl false start via alpn")
	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		os.Exit(0)
	}

	if thread < 0 || delay < 0 || host == "" {
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
			ServerName: strings.Split(host, "://")[1],
			//Renegotiation: tls.RenegotiateNever,
		}
		if falseStart {
			tlscfg.NextProtos = []string{"http/1.1", "h2"} // http 2 support for alpn
		}
		if sessionResume {
			tlscfg.ClientSessionCache = tls.NewLRUClientSessionCache(128)
		} else {
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
	if timeout > 0 {
		client.Timeout = time.Second * time.Duration(timeout)
	}

	if h2 {
		client = &http.Client{
			Transport: tr2,
		}
	} else {
		client = &http.Client{
			Transport: tr,
		}
	}

	perf := perfm.New(perfm.WithBinsNumber(15))

	var errCode error = errors.New("!")

	perf.Registe(func() (err error) {
		resp, err := client.Do(req)
		if resp.StatusCode > 400 {
			return errCode
		}
		return
	})

	perf.Start()
	perf.Wait()
}
