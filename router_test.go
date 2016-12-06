package brouter

import (
	"bytes"
	"github.com/bysir-zl/bygo/util"
	"github.com/bysir-zl/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"log"
	"regexp"
	"testing"
)

func TestAppend(t *testing.T) {
	type X struct {
		a string
	}
	x := X{}
	xs := []X{}
	xs = append(xs, x)

	log.Printf("%p", &x)
	log.Printf("%p", &xs[0])
}

func TestMatch(t *testing.T) {
	r := New()
	r.Use(func(c *Context) {
		log.Print("in root")
	})
	r.Any("/13", func(c *Context) {
		log.Print("in root b")
	})
	r.Get("/13", func(c *Context) {
		log.Print("in root b")
	})
	r.Group("/v1", func(a *Node) {
		a.Any("/a/123", func(c *Context) {
			log.Print("in a")
		})
		a.Any("/b/*", func(c *Context) {
			log.Print("in b")
		})
		a.Any("/c/(name)/(x)", func(c *Context) {
			log.Print("in c")
		})
		a.Any("", func(c *Context) {
			log.Print("in black")
		})
		a.Any(`/user/(id:\d+)`, func(c *Context) {
			log.Print("in user")
		}).UseToChild(func(c *Context) {
			log.Print("in end2")
		})
	}).UseToChild(func(c *Context) {
		log.Print("in end")
	}).Use(func(c *Context) {
		log.Print("in start")
	})

	r.Init()

	for _,p:=range r.paths{
		log.Print(string(p.path),"--",string(p.method))
	}

	// init

	status,nodes, keys, values := match(util.S2B("/v1/c/1/2"), util.S2B("GET"), r)

	log.Print("---------")
	log.Print(status)
	log.Print(keys, values)
	run(nodes, new(Context))
}

func TestHand(t *testing.T) {
	r := New()
	r.Use(func(c *Context) {
		log.Print("in root")
	})
	r.Any("/", func(c *Context) {
		log.Print("in root b")
	})
	a := r.Group(`/v1`, func(a *Node) {
		//a.Any(`/a/(\d+?)`, func(c *Context) {
		//	log.Print("in a")
		//	log.Print(c.routerParams)
		//})
		//a.Any("/b/*", func(c *Context) {
		//	log.Print("in b")
		//	log.Print(c.routerParams)
		//})
		//a.Any(`/c/(id:\d+?)`, func(c *Context) {
		//	log.Print("in c")
		//	log.Print(c.routerParams)
		//})
		//a.Any("/", func(c *Context) {
		//	log.Print("in black")
		//})
		//a.Any(`/user/(id:.+?)/(name:.+?)`, func(c *Context) {
		//	c.Write(c.routerParams["id"])
		//	c.Write(c.routerParams["name"])
		//	log.Print(c.routerParams)
		//
		//}).UseToChild(func(c *Context) {
		//	log.Print("in end2")
		//})
	}).UseToChild(func(c *Context) {
		log.Print("in end")
	}).Use(func(c *Context) {
		log.Print("in start")
	})

	a.Any("/a", func(c *Context) {
		log.Print("in a")
	})
	a.Any("/b/*", func(c *Context) {
		log.Print("in b")
	})
	a.Any(`/user/(id:\d+?)/`, func(c *Context) {
		log.Print("in user")
	})

	f := r.Init()
	x := &fasthttp.RequestCtx{}
	x.URI().SetPath("v1/user/1")
	f(x)

	//fasthttp.ListenAndServe(":8081", r.Init())
}

func ma2(sb, ub []byte) (params map[string][]byte) {
	c := regexp.MustCompile(`\(.+?:(.+?)\)`)

	cname := regexp.MustCompile(`\((.+?):.+?\)`)
	names := cname.FindAllSubmatch(sb, -1)
	nas := [][]byte{}
	for _, v := range names {
		nas = append(nas, v[1])
	}

	sb = c.ReplaceAll(sb, []byte("($1)"))
	cc := regexp.MustCompile(util.B2S(sb))
	vs := cc.FindAllSubmatch(ub, -1)[0][1:]
	kvs := map[string][]byte{}
	for i, v := range vs {
		kvs[string(nas[i])] = v
	}
	return kvs
}

func BenchmarkMatch(b *testing.B) {

	b.StopTimer()
	s := `user/(id:\d+)/(name:.+)/`
	u := "user/9987/zl/"
	sb := util.S2B(s)
	ub := util.S2B(u)

	b.StartTimer()
	params := map[string][]byte{}
	//params := map[string][]byte{}
	for i2 := 0; i2 < b.N; i2++ {
		//  20631 ns/op
		//ma(sb, ub)
		//continue
		// 243 ns/op
		ss := bytes.Split(sb, []byte{'('})

		for i, v := range ss {
			if i == 0 {
				if bytes.Index(ub, v) != 0 {
					break
				}
				params = map[string][]byte{}
				// 23456/2zl/
				ub = ub[len(v):]
			} else {
				// [id:\d+ , /2]
				ks := bytes.Split(v, []byte{')'})
				// 23456
				urlV := bytes.Split(ub, ks[1])[0]
				nameReg := bytes.Split(ks[0], []byte{':'})
				name := nameReg[0]
				reg := nameReg[1]
				regString := util.B2S(reg)
				if regString[len(regString)-1] != '$' {
					regString = regString + "$"
				}
				if regString[0] != '^' {
					regString = "^" + regString
				}
				ok, _ := regexp.Match(regString, urlV)
				if !ok {
					break
				} else {
					// 已经匹配到结束
					if len(urlV) == len(ub) {
						params[string(name)] = urlV
						break
					}
					ub = ub[len(urlV)+len(ks[1]):]
					params[string(name)] = urlV
				}
			}
		}
	}
	log.Print(params)

}

// 1981 ns/op
func BenchmarkMatch2(b *testing.B) {

	s := `user/(id:\d+)/(name:.+)/`
	u := "user/9987/zl/"
	url := util.S2B(s)
	ub := util.S2B(u)
	c := regexp.MustCompile(`\(.+?:(.+?)\)`)

	cname := regexp.MustCompile(`\((.+?):.+?\)`)
	names := cname.FindAllSubmatch(url, -1)
	regNames := [][]byte{}
	for _, v := range names {
		regNames = append(regNames, v[1])
	}

	url = c.ReplaceAll(url, []byte("($1)"))
	reg := regexp.MustCompile(util.B2S(url))

	for i := 0; i < b.N; i++ {
		// 20331 ns/op
		//ma(sb, ub)

		vs := reg.FindAllSubmatch(ub, -1)
		if len(vs) == 1 {
			k := vs[0][1:]
			kvs := map[string][]byte{}
			for i, v := range k {
				kvs[string(regNames[i])] = v
			}

		}
	}
}

func BenchmarkRouterMy(b *testing.B) {
	b.StopTimer()
	r := New()
	a := r.Group("/v1")
	//a.Any("/a", func(c *Context) {
	//log.Print("in a")
	//})
	//a.Any("/b/*", func(c *Context) {
	//log.Print("in b")
	//})
	a.Any("/c/(id)/(a)/(b)/(c)/(d)/(e)/(g)", func(c *Context) {
		//	//log.Print("in c")
	})
	a.Any(`/user/(id:\d+?)/(a:.*?)/(a:.*?)/(a:.*?)/(a:.*?)/(a:.*?)`, func(c *Context) {
		//log.Print("in user")
	})

	f := r.Init()

	x := &fasthttp.RequestCtx{}
	// 5110 ns/op
	x.URI().SetPath("/v1/user/1/a/b/c/d/e/g")
	// 136 ns/op
	//x.URI().SetPath("/v1/a/")
	// 565 ns/op | 776 ns/op
	//x.URI().SetPath("v1/b/132/123")
	// 5768 ns/op
	x.URI().SetPath("v1/c/132/a/b/c/d/e/g")
	//f(x)

	b.StartTimer()

	//nodes, params := match(util.S2B("/v1/c/1"), r)

	//log.Print("---------")
	//log.Print(params)
	//run(nodes, new(Context))

	for i := 0; i < b.N; i++ {
		f(x)
	}
}

func BenchmarkRouterRouting(b *testing.B) {
	b.StopTimer()
	r := routing.New()
	a := r.Group("/v1")
	a.Any("/a", func(c *routing.Context) (e error) {
		//log.Print("in a")
		return
	})
	a.Any("/b/*", func(c *routing.Context) (e error) {
		//log.Print("in b")
		return
	})
	a.Any(`/user/<id>/<name>`, func(c *routing.Context) (e error) {
		//log.Print("in user",c.Param("id"))
		return
	})

	f := r.HandleRequest

	b.StartTimer()
	x := &fasthttp.RequestCtx{}
	// 422 ns/op
	x.URI().SetPath("v1/user/13456/1")
	// 306 ns/op
	//x.URI().SetPath("v1/a")

	for i := 0; i < b.N; i++ {
		f(x)
	}

}
