package shardhttp

import (
	"encoding/json"
	"hash/crc32"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var (
	enableLog = os.Getenv("SHARD_ENABLE_LOG") == "true"
)

func jlog(i interface{}) {
	bin, _ := json.MarshalIndent(i, " ", " ")
	log.Print(string(bin))
}

func flog(data ...interface{}) {
	if !enableLog {
		return
	}
	log.Print(data...)
}

const (
	default_scheme = "http"
)

// calc extract address from list
func GetShardAddressFromShardKey(skey string, addrs []string) (string, int) {
	index := int(crc32.ChecksumIEEE([]byte(skey))) % len(addrs)
	flog(skey, " ", index)
	host := addrs[index]
	return host, index
}

func proxy(path string, req *http.Request, resp http.ResponseWriter, headerExtend http.Header) error {
	// log.Println("path:", path)
	remote, err := url.Parse(path)
	if err != nil {
		return err
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	// add more data to header
	if len(headerExtend) > 0 {
		for key, val := range headerExtend {
			req.Header[key] = val
		}
	}
	if remote.Scheme == "" {
		remote.Scheme = default_scheme
	}

	//Define the director func
	//This is a good place to log, for example
	proxy.Director = func(req *http.Request) {
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
	}

	proxy.ServeHTTP(resp, req)
	return nil
}
