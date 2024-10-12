package proxy

import (
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/auth"
)

func authenticate(user string, passwordHash string) bool {
	return true
}

func ListenAndServe() {
	proxy := goproxy.NewProxyHttpServer()
	auth.ProxyBasic(proxy, "Auth", authenticate)
	log.Fatal(http.ListenAndServe(":8888", proxy))
}
