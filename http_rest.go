package nimo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	HttpGet    = "GET"
	HttpPost   = "POST"
	HttpUpdate = "UPDATE"
)

type HttpRestProvider struct {
	controller map[URL]*HandlerList
	port       int
}

type URL struct {
	uri    string
	method string
}

type HandlerList struct {
	handers []func(body []byte) interface{}
}

func NewHttpRestProvdier(port int) *HttpRestProvider {
	return &HttpRestProvider{controller: make(map[URL]*HandlerList, 512), port: port}
}

func (rest *HttpRestProvider) Listen() error {
	for web, handlerList := range rest.controller {
		register(web, handlerList)
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", rest.port), nil)
}

func (rest *HttpRestProvider) RegisterAPI(url string, method string, handler func(body []byte) interface{}) {
	if c, exist := rest.controller[URL{uri: url, method: method}]; exist {
		c.handers = append(c.handers, handler)
	} else {
		rest.controller[URL{uri: url, method: method}] = &HandlerList{
			handers: []func([]byte) interface{}{handler},
		}
	}
}

func register(web URL, handlerList *HandlerList) {
	http.HandleFunc(web.uri, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != web.method {
			return
		}
		// read full body content
		body, _ := ioutil.ReadAll(r.Body)

		var v []byte
		var stuff interface{}
		if len(handlerList.handers) > 1 {
			var results []interface{}
			// aggregate all results if multi-controller register
			for _, handler := range handlerList.handers {
				results = append(results, handler(body))
			}
			stuff = results
		} else {
			stuff = handlerList.handers[0](body)
		}
		v, _ = json.Marshal(stuff)
		w.Write(v)
	})
}
