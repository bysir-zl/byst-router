package brouter

import (
	"bytes"
	"github.com/bysir-zl/bygo/util"
	"github.com/valyala/fasthttp"
)

type Context struct {
	*fasthttp.RequestCtx
	rKeys   [][]byte               // params on url, eg: user\(id:\d+)
	rValues [][]byte               // params on url, eg: user\(id:\d+)
	items   map[string]interface{} // get and set data on one context
	abort   bool
}

// get data in the context
func (c *Context) Get(key string) interface{} {
	return c.items[key]
}

// set data to the context
func (c *Context) Set(key string, obj interface{}) {
	if c.items == nil {
		c.items = map[string]interface{}{}
	}
	c.items[key] = obj
}

// get param from url
func (c *Context) Param(name string) string {
	nbs := util.S2B(name)
	for i, key := range c.rKeys {
		if bytes.Equal(nbs, key) {
			return util.B2S(c.rValues[i])
		}
	}
	return ""
}

func (c *Context) Abort() {
	c.abort = true
}
