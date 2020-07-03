package main

import "github.com/tufin/blog/go-proxy/common"

func main() {

	_, _ = common.NewHTTPClient().Get("https://api.wind.io/forecast")
}
