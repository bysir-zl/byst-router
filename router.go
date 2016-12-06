package brouter

import (
	"bytes"
	"github.com/bysir-zl/bygo/util"
	"github.com/valyala/fasthttp"
	"regexp"
	"sync"
)

const (
	type_nomal = iota // 普通的无任何参数 直接判断是否相等
	type_reg          // 正则 如 id/(id:\d+?)/
	type_tang         // 语法糖 如 bong/*/
	type_param        // 有参数 但不是正则 如 /id/(id)/
)
const (
	status_notfound = iota
	status_methodnotallow
	status_matched
)

type path struct {
	path     []byte         // the full path
	reg      *regexp.Regexp // compiled path reg
	regNames [][]byte       // name with $1 $2
	nodes    []*Node
	types    int
	method   []byte
}

type Router struct {
	root       *Node
	paths      []path
	handler404 Handler
	handler405 Handler
}

// return self
func (p *Router) Any(path string, h ...Handler) *Node {
	return p.root.Any(path, h...)
}

func (p *Router) Get(path string, h ...Handler) *Node {
	return p.root.Get(path, h...)
}

func (p *Router) Post(path string, h ...Handler) *Node {
	return p.root.Post(path, h...)
}

func (p *Router) Put(path string, h ...Handler) *Node {
	return p.root.Put(path, h...)
}

func (p *Router) Delete(path string, h ...Handler) *Node {
	return p.root.Delete(path, h...)
}
func (p *Router) Option(path string, h ...Handler) *Node {
	return p.root.Option(path, h...)
}
func (p *Router) Head(path string, h ...Handler) *Node {
	return p.root.Head(path, h...)
}
func (p *Router) Controller(path string, c interface{}) *Node {
	return p.root.Controller(path, c)
}
// set 404 handler
func (p *Router) When404(h Handler) {
	p.handler404 = h
}
// set 404 handler
func (p *Router) When405(h Handler) {
	p.handler405 = h
}

func (p *Router) Use(h ...Handler) *Node {
	return p.root.Use(h...)
}

func (p *Router) UseToChild(h ...Handler) *Node {
	return p.root.UseToChild(h...)
}

// return child
func (p *Router) Group(path string, funs ...func(node *Node)) *Node {
	return p.root.Group(path, funs...)
}

var pool sync.Pool

func (p *Router) Init() func(ctx *fasthttp.RequestCtx) {
	pool.New = func() interface{} {
		return &Context{}
	}

	p.paths = parse(p.root)
	f := func(ctx *fasthttp.RequestCtx) {
		c := pool.Get().(*Context)
		c.RequestCtx = ctx
		url := ctx.Path()
		status, nodes, keys, values := match(url, ctx.Method(), p)
		switch status {
		case status_notfound:
			when404(c)
		case status_methodnotallow:
			when405(c)
		case status_matched:
			c.rKeys = keys
			c.rValues = values
			run(nodes, c)
		}
		pool.Put(c)
	}
	return f
}

func New() *Router {
	r := Router{}
	node := &Node{
		name:  util.S2B("root"),
		child: []*Node{},
		path:  []byte{},
	}
	r.root = node
	r.handler404 = when404
	r.handler405 = when405
	return &r
}

// init
// parse full paths from router's node
func parse(root *Node) []path {
	p([]byte{}, root, []*Node{})
	pps := []path{}
	for urlString, nodes := range paths {
		url := util.S2B(urlString)
		methodAurl := bytes.Split(url, []byte("::"))
		url = methodAurl[1]
		method := methodAurl[0]
		var regNames [][]byte
		var reg *regexp.Regexp
		types := type_nomal
		// 有 : 就表示是正则路由
		if bytes.Contains(url, []byte{':'}) {
			// 找出名字 (id:\d+?)
			cname := regexp.MustCompile(`\((.+?):.+?\)`)
			names := cname.FindAllSubmatch(url, -1)
			regNames = [][]byte{}
			for _, v := range names {
				n := v[1]
				regNames = append(regNames, n)
			}

			// 找出真正的正则表达式
			c := regexp.MustCompile(`\(.+?:(.+?)\)`)
			url2 := c.ReplaceAll(url, []byte("($1)"))
			// 加上前后界限
			// 加上? 因为 ^/v1/c/(.*?)/?$ 当无参数时(url: /v1/c/)也能匹配到
			s := "^" + util.B2S(url2) + "?$"

			//log.Print(s)
			reg = regexp.MustCompile(s)
			types = type_reg
		} else if bytes.Contains(url, []byte{'('}) {
			types = type_param
		} else if lurl := len(url); lurl > 2 && url[lurl - 2] == '*' {
			types = type_tang
		}
		p := path{
			path:     url,
			types:    types,
			nodes:    nodes,
			method:   method,
			reg:      reg,
			regNames: regNames,
		}
		pps = append(pps, p)
	}
	return pps
}

var paths = map[string][]*Node{}

func p(url []byte, node *Node, path []*Node) {
	url = append(url, node.path...)
	path = append(path, node)
	if node.child != nil && len(node.child) != 0 {
		for i := range node.child {
			p(url, node.child[i], path)
		}
	} else {
		if url[len(url) - 1] != '/' {
			url = append(url, '/')
		}
		u := string(node.method) + "::" + string(url)
		paths[u] = path
	}
}

func when404(c *Context) {
	c.WriteString("404 not found")
	c.SetStatusCode(404)
}

func when405(c *Context) {
	c.WriteString("405 method not allowed")
	c.SetStatusCode(405)
}

func run(nodes []*Node, c *Context) {
	abort := false
	for _, node := range nodes {
		for _, handle := range node.handlers {
			if c.abort {
				abort = true
				break
			}
			handle(c)
		}
		if abort {
			break
		}
	}
}

// match router
func match(url []byte, method []byte, router *Router) (status int, nodes []*Node, keys [][]byte, values [][]byte) {
	if len(url) == 0 {
		url = []byte{'/'}
	} else if url[len(url) - 1] != '/' {
		url = append(url, '/')
	}
	for _, path := range router.paths {
		status, keys, values := isMatched(path, url, method)
		if status != 0 {
			return status, path.nodes, keys, values
		}
	}
	return
}

var rk = []byte{')'}
var lk = []byte{'('}

// match one url and return url params
func isMatched(path path, url []byte, method []byte) (status int, keys, values [][]byte) {
	urlBase := path.path
	switch path.types {
	case type_tang:
		// 语法糖
		// 匹配 user/*/  => user/a/b/c/  => [0:a,1:b,2:c]
		lbase := len(urlBase)
		if bytes.Equal(url[:lbase - 2], urlBase[:lbase - 2]) {
			status = status_methodnotallow
			if path.method[0] == 'A' || bytes.Equal(method, path.method) {
				status = status_matched
			}
			if len(url) == lbase - 2 {
				return
			}
			kvs := bytes.Split(url[lbase - 2:len(url) - 1], []byte{'/'})
			lkvs := len(kvs)
			keys = make([][]byte, lkvs)
			values = make([][]byte, lkvs)

			for i := 0; i < lkvs; i++ {
				//params[strconv.Itoa(i)] = v
				keys[i] = []byte{byte(i + 49)}
				values[i] = kvs[i]
			}
			return
		}
	case type_nomal:
		if bytes.Equal(urlBase, url) {
			status = status_methodnotallow
			if path.method[0] == 'A' || bytes.Equal(method, path.method) {
				status = status_matched
			}
			return
		}
	case type_param:
		// user/(id)/
		// user/123/
		us := bytes.Split(urlBase, lk)
		lus := len(us)
		//ok = true
		for i := 0; i < lus; i++ {
			u := us[i]
			if i == 0 {
				// user/
				if bytes.Index(url, u) != 0 {
					status = 0
					return
				}
				keys = make([][]byte, lus - 1)
				values = make([][]byte, lus - 1)
				// 132/
				url = url[len(u):]
			} else {
				nameAsp := bytes.Split(u, rk)
				// id
				//name := nameAsp[0]
				// /
				//sp := nameAsp[1]
				value := bytes.Split(url, nameAsp[1])
				keys[i - 1] = nameAsp[0]
				values[i - 1] = value[0]
				if len(value[1]) == 0 {
					// is end of url
					if i == lus - 1 {
						status = status_methodnotallow
						if path.method[0] == 'A' || bytes.Equal(method, path.method) {
							status = status_matched
						}
						return
					} else {
						// router url is not end, but url is end .
						// has not enough params
						status = 0
						return
					}
				}
				url = url[len(value[0]) + 1:]
			}
		}
		return
	case type_reg:
		vs := path.reg.FindAllSubmatch(url, -1)
		if len(vs) == 1 {
			status = status_methodnotallow
			if path.method[0] == 'A' || bytes.Equal(method, path.method) {
				status = status_matched
			}

			k := vs[0][1:]
			lk := len(k)
			keys = make([][]byte, lk)
			values = make([][]byte, lk)

			for i := 0; i < lk; i++ {
				keys[i] = path.regNames[i]
				values[i] = k[i]
			}
		}
		return
	}
	return
}
