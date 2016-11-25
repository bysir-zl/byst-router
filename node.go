package byrouter

import (
	"github.com/bysir-zl/bygo/util"
)

type Handler func(c *Context)

// use a tree struct to save all url node
type Node struct {
	handlers []Handler // run handler when the node is matched
	name     []byte    // alias
	child    []*Node   // he's child
	path     []byte    // path
	method   []byte
}

var METHOD_GET = util.S2B("GET")
var METHOD_POST = util.S2B("POST")
var METHOD_PUT = util.S2B("PUT")
var METHOD_DELETE = util.S2B("DELETE")
var METHOD_ANY = util.S2B("ANY")

func (p *Node)Any(path  string, h ...Handler) *Node {
	return p.add(METHOD_ANY, path, h)
}

func (p *Node)Get(path  string, h ...Handler) *Node {
	return p.add(METHOD_GET, path, h)
}

func (p *Node)Post(path  string, h ...Handler) *Node {
	return p.add(METHOD_POST, path, h)
}

func (p *Node)Put(path  string, h ...Handler) *Node {
	return p.add(METHOD_PUT, path, h)
}

func (p *Node)Delete(path  string, h ...Handler) *Node {
	return p.add(METHOD_DELETE, path, h)
}

// add a handle to handle list
func (p *Node)Use(h ...Handler) *Node {
	p.handlers = append(p.handlers, h...)
	return p
}
func (p *Node)UseToChild(h ...Handler) *Node {
	if len(p.child) == 0 {
		p.handlers = append(p.handlers, h...)
	} else {
		for i := range p.child {
			p.child[i].UseToChild(h...)
		}
	}
	return p
}

// if u want use index format to code router
// u can set funs param and code in it
func (p *Node) Group(path  string, funs ...func(node *Node)) *Node {
	//log.Printf("g : %p",p.child[0])
	n := &Node{
		child:[]*Node{},
		name:util.S2B("group"),
		handlers:[]Handler{},
		path:util.S2B(path),
	}

	p.child = append(p.child, n)
	if funs != nil && len(funs) != 0 {
		funs[0](n)
	}
	return n
}

func (p *Node) add(method []byte, path  string, h []Handler) *Node {
	if h == nil {
		h = []Handler{}
	}
	n := &Node{
		child:[]*Node{},
		handlers:h,
		method:method,
		path:util.S2B(path),
	}

	p.child = append(p.child, n)
	return n
}