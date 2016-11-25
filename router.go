package byrouter

import (
	"github.com/bysir-zl/bygo/util"
	"bytes"
	"regexp"
	"strconv"
	"github.com/valyala/fasthttp"
)

type path struct {
	path     []byte         // the full path
	reg      *regexp.Regexp // compiled path reg
	regNames [][]byte       // name with $1 $2
	nodes    []*Node
}

type Router struct {
	root     *Node
	rootPath []byte
	paths    []path
}

// return self
func (p *Router)Any(path  string, h ...Handler) *Node {
	return p.root.Any(path, h...)
}

func (p *Router)Get(path  string, h ...Handler) *Node {
	return p.root.Get(path, h...)
}

func (p *Router)Post(path  string, h ...Handler) *Node {
	return p.root.Post(path, h...)
}

func (p *Router)Put(path  string, h ...Handler) *Node {
	return p.root.Put(path, h...)
}

func (p *Router)Delete(path  string, h ...Handler) *Node {
	return p.root.Delete(path, h...)
}

func (p *Router)Use(h ...Handler) *Node {
	return p.root.Use(h...)
}

func (p *Router)UseToChild(h ...Handler) *Node {
	return p.root.UseToChild(h...)
}

// return child
func (p *Router)Group(path  string, funs ...func(node *Node)) *Node {

	return p.root.Group(path, funs...)
}

func (p *Router)Init() func(ctx *fasthttp.RequestCtx) {
	p.paths = parse(p.root)
	f := func(ctx *fasthttp.RequestCtx) {
		url := ctx.Path()
		c := &Context{RequestCtx:ctx}
		nodes, ps := match(url, p)
		c.routerParams = ps
		run(nodes, c)
	}
	return f
}

func NewRouter() *Router {
	r := Router{}
	node := &Node{
		name:util.S2B("root"),
		child:[]*Node{},
		path:[]byte{},
	}
	r.root = node
	return &r
}

// parse paths from router's node
func parse(root *Node) []path {
	p([]byte{}, root, []*Node{})
	pps := []path{}
	for urlString, nodes := range paths {
		url := util.S2B(urlString)
		var regNames [][]byte
		var reg *regexp.Regexp
		if bytes.Contains(url, []byte{'('}) {

			// 可以不写名字 (id:\d+?) => (\d+?)
			// 那么将默认使用 0 1 2 3 4... 去命名
			cname := regexp.MustCompile(`\((.*?:?.+?\))`)
			names := cname.FindAllSubmatch(url, -1)
			regNames = [][]byte{}
			for i, v := range names {
				n := v[1]
				if bytes.Contains(n, []byte{':'}) {
					n = bytes.Split(n, []byte{':'})[0]
				} else {
					n = util.S2B(strconv.Itoa(i))
				}
				regNames = append(regNames, n)
			}

			// 找出真正的正则表达式
			c := regexp.MustCompile(`\(.*?:(.+?)\)`)
			url2 := c.ReplaceAll(url, []byte("($1)"))
			// 加上前后界限
			// 加上? 因为 ^/v1/c/(.*?)/?$ 当无参数时(url: /v1/c/)也能匹配到
			s := "^" + util.B2S(url2) + "?$"

			//log.Print(s)
			reg = regexp.MustCompile(s)
		}
		p := path{
			path:url,
			nodes:nodes,
			reg:reg,
			regNames:regNames,
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
		paths[string(url)] = path
	}
}

func run(nodes []*Node, c *Context) {
	for _, node := range nodes {
		for _, handle := range node.handlers {
			handle(c)
		}
	}
}


// match router
func match(url []byte, router *Router) (nodes []*Node, params map[string][]byte) {
	if len(url) == 0 {
		url = []byte{'/'}
	} else if url[len(url) - 1] != '/' {
		url = append(url, '/')
	}
	for _, path := range router.paths {
		ok, params := isMatched(path, url)
		if ok {
			return path.nodes, params
		}
	}
	return
}


// match one url and return url params
func isMatched(path path, url []byte) (ok bool, params map[string][]byte) {
	urlBase := path.path
	if bytes.Equal(urlBase, url) {
		ok = true
		return
	}
	// 正则匹配
	if path.reg != nil {
		vs := path.reg.FindAllSubmatch(url, -1)
		if len(vs) == 1 {
			ok = true
			k := vs[0][1:]
			params = map[string][]byte{}
			for i, v := range k {
				params[string(path.regNames[i])] = v
			}
		}
		return
	}

	// 语法糖
	// 匹配 user/*/  => user/a/b/c/  => [0:a,1:b,2:c]
	lbase := len(urlBase)
	if lbase > 2 && urlBase[lbase - 2] == '*' &&len(url) >= lbase - 2 && bytes.Equal(url[:lbase - 2], urlBase[:lbase - 2]) {
		params = map[string][]byte{}
		ok = true
		if len(url) == lbase - 2 {
			return
		}
		for i, v := range bytes.Split(url[lbase - 2:len(url) - 1], []byte{'/'}) {
			params[strconv.Itoa(i)] = v
		}
		return
	}

	return
}



