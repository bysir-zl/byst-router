package byrouter

import "github.com/valyala/fasthttp"

type Context struct {
	*fasthttp.RequestCtx
	routerParams map[string][]byte      // params on url, eg: user\(id:\d+)
	items        map[string]interface{} // get and set data on one context
}

func (c *Context)Get(key string) interface{} {
	return c.items[key]
}
func (c *Context)Set(key string, obj interface{}) {
	c.items[key] = obj
}
