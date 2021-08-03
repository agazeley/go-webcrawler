package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	baseUrl = "http://www.example.com"

	noAnchorsHtml = `<html>
						<body>
							Hello world!
						</body>
					</html>`
	oneAnchorHtml = fmt.Sprintf(`
					<html>
						<body>
						<a href="%v%v">Link</a>
						</body>
					</html>`, baseUrl, "/best")
	twoAnchorHtml = fmt.Sprintf(`
					<html>
						<body>
						<a href="%v%v">Link1</a>
						<a href="%v%v">Link2</a>
						</body>
					</html>`, baseUrl, "/best", baseUrl, "/test")
)

func NewTestServer(address string, routes map[string]string) *httptest.Server {
	mux := http.NewServeMux()

	for path, html := range routes {
		mux.HandleFunc(path, func(rw http.ResponseWriter, r *http.Request) {
			rw.Write([]byte(html))
		})
	}

	ts := httptest.NewServer(mux)
	return ts
}

func checkLen(t *testing.T, items interface{}, expected int) {
	listVal := reflect.ValueOf(items)
	if listVal.Len() != expected {
		t.Log("Expected", expected, "items. Got", listVal.Len())
		t.Fail()
	}
}

func Test_IsValidUrl(t *testing.T) {
	res := isValidUrl("")
	if res {
		t.Fail()
	}

	res = isValidUrl("adbcd")
	if res {
		t.Fail()
	}

	res = isValidUrl("https://")
	if res {
		t.Fail()
	}

	res = isValidUrl("www.google.com")
	if res {
		t.Fail()
	}

	res = isValidUrl(baseUrl)
	if !res {
		t.Fail()
	}
}

func Test_ParseArgs(t *testing.T) {
	_, err := parseArgs([]string{})
	if err == nil {
		t.Fail()
	}

	_, err = parseArgs([]string{"program", "google.com"})
	if err == nil {
		t.Fail()
	}

	opts, _ := parseArgs([]string{"program", "http://www.google.com"})
	if opts == nil {
		t.Fail()
	}

	opts, _ = parseArgs([]string{"program", "http://www.google.com", "10"})
	if opts.maxCrawls != 10 {
		fmt.Println(opts)
		t.Fail()
	}
}

func Test_ParseResponse(t *testing.T) {
	resp := &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(noAnchorsHtml)),
	}

	urls := parseResponse(resp)

	checkLen(t, urls, 0)

	resp = &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(oneAnchorHtml)),
	}

	urls = parseResponse(resp)
	checkLen(t, urls, 1)

	resp = &http.Response{
		Body: ioutil.NopCloser(bytes.NewBufferString(twoAnchorHtml)),
	}
	urls = parseResponse(resp)
	checkLen(t, urls, 2)
}

func Test_ScrapeUrl(t *testing.T) {
	ts1 := NewTestServer(baseUrl, map[string]string{"/": noAnchorsHtml})
	urls := scrapeUrl(ts1.URL)
	checkLen(t, urls, 0)
	defer ts1.Close()

	ts2 := NewTestServer(baseUrl, map[string]string{"/": twoAnchorHtml})
	defer ts2.Close()
	urls = scrapeUrl(ts2.URL)
	checkLen(t, urls, 2)
}
