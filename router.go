package byrouter

import "github.com/valyala/fasthttp"

type Handler func(c *Context)

type Context struct {
	*fasthttp.RequestCtx
	routerParams map[string][]byte      // params on url, eg: user\(id:\d+)
	items        map[string]interface{} // get and set data on one context
	node         Node                   // the matched node, use to get some params
}

// use a tree struct to save all url node
type Node struct {
	handlers []Handler // run handler by order when the node is matched
	name     []byte    // alias
	child    []Node    // he's child
	path     []byte    // path
}

type Router struct {
	root     Node
	rootPath []byte
}

func NewRouter() *Node {
	r := Router{}
	node := Node{
		handlers:[]Handler{},
		name:"root",
		child:[]Node{},
		path:"",
	}
	r.root = node
	return &r
}

func match(url []byte, ) {

}