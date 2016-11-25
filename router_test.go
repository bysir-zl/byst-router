package byrouter

import (
	"testing"
	"github.com/bysir-zl/bygo/util"
	"log"
	"regexp"
	"bytes"
	"github.com/valyala/fasthttp"
	"github.com/bysir-zl/fasthttp-routing"
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
	//s := map[string]string{
	//	//`user/1(id:\d+)/2(name:.+)/`:    "user/12345678/2zl/",
	//	//`user/1(x:\d+)/2(name:.+)/`:    "user/12345678/2zl/",
	//	//`user/(1:\d+)/(name:.+)/`:    "user/12345678/zl/",
	//	`user/a(id:\d+)/n(name:.+)/`:    "user/a9987/nzl/",
	//}
	//
	//for s, u := range s {
	//	sb := util.S2B(s)
	//	ub := util.S2B(u)
	//
	//	kvs := ma(sb, ub)
	//
	//	log.Print(kvs)
	//}

	r := NewRouter()
	r.Use(func(c *Context) {
		log.Print("in root")
	})
	r.Any("/", func(c *Context) {
		log.Print("in root b")
	})
	r.Group("v1/", func(a *Node) {
		a.Any("a/123", func(c *Context) {
			log.Print("in a")
		})
		a.Any("b/*", func(c *Context) {
			log.Print("in b")
		})
		a.Any("", func(c *Context) {
			log.Print("in black")
		})
		a.Any(`user/(id:\d+)`, func(c *Context) {
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

	// init

	nodes, params := match(util.S2B("v1/"), r)

	log.Print("---------")
	log.Print(params)
	run(nodes, new(Context))
}

func TestHand(t *testing.T) {
	r := NewRouter()
	r.Use(func(c *Context) {
		log.Print("in root")
	})
	r.Any("/", func(c *Context) {
		log.Print("in root b")
	})
	r.Group(`/v1`, func(a *Node) {
		a.Any(`/a/(\d+?)`, func(c *Context) {
			log.Print("in a")
			log.Print(c.routerParams)
		})
		a.Any("/b/*", func(c *Context) {
			log.Print("in b")
			log.Print(c.routerParams)
		})
		a.Any(`/c/(id:\d+?)`, func(c *Context) {
			log.Print("in c")
			log.Print(c.routerParams)
		})
		a.Any("/", func(c *Context) {
			log.Print("in black")
		})
		a.Any(`/user/(id:.+?)/(name:.+?)`, func(c *Context) {
			c.Write(c.routerParams["id"])
			c.Write(c.routerParams["name"])
			log.Print(c.routerParams)

		}).UseToChild(func(c *Context) {
			log.Print("in end2")
		})
	}).UseToChild(func(c *Context) {
		log.Print("in end")
	}).Use(func(c *Context) {
		log.Print("in start")
	})

	fasthttp.ListenAndServe(":8081", r.Init())
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
				if regString[len(regString) - 1] != '$' {
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
					if (len(urlV) == len(ub)) {
						params[string(name)] = urlV
						break
					}
					ub = ub[len(urlV) + len(ks[1]):]
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
	r := NewRouter()

	a := r.Group("v1/")
	a.Any("a", func(c *Context) {
		log.Print("in a")
	})
	a.Any("b/*", func(c *Context) {
		log.Print("in b")
	})
	a.Any(`user/(id:\d+)/`, func(c *Context) {
		log.Print("in user")
	})

	r.Init()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		// 1512 ns/op
		match(util.S2B("v1/user/1"), r)
		// 71.4 ns/op
		match(util.S2B("v1/a"), r)
	}

	routing.Route{}
}

