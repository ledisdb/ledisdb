package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestHttp(t *testing.T) {
	startTestApp()

	r, err := http.Get(fmt.Sprintf("http://%s/SET/http_hello/world", testApp.cfg.HttpAddr))
	if err != nil {
		t.Fatal(err)
	}

	ioutil.ReadAll(r.Body)
	r.Body.Close()

	r, err = http.Get(fmt.Sprintf("http://%s/GET/http_hello?type=json", testApp.cfg.HttpAddr))
	if err != nil {
		t.Fatal(err)
	}

	b, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	var v struct {
		Data string `json:"GET"`
	}

	if err = json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	} else if v.Data != "world" {
		t.Fatal("not equal")
	}

	// XSCAN should not give BASE64 keys
	r, err = http.Get(fmt.Sprintf("http://%s/XSCAN/KV/", testApp.cfg.HttpAddr))
	if err != nil {
		t.Fatal(err)
	}

	b, _ = ioutil.ReadAll(r.Body)
	r.Body.Close()
	if string(b) != `{"XSCAN":["",["http_hello"]]}` {
		t.Fatal("XSCAN result not correct")
	}
}
