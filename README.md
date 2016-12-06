# byst-router
a light router for fast-http

## preview
router.go
```go
r := brouter.New()
r.Use(func(c *brouter.Context) {
    log.Print("in root")
})
r.Any("/", func(c *brouter.Context) {
    log.Print("black")
})
r.Post("/api",func(c *brouter.Context){
    log.Print("/api")
})
r.Group("/v1",func(r *brouter.Node){
    r.Get("/user",func(c *brouter.Context){
        log.Print("/v1/user")
    })
    r.Post("/user/(id)",func(c *brouter.Context){
        log.Print("/v1/user/",c.Param("id"))
    })
    r.Get(`/article/(id:\d+?)`,func(c *brouter.Context){
        log.Print("/v1/article/",c.Param("id"))
    })
    r.Controller("/news",NewsController{})    
}).Use(func (c *brouter.Context){
    log.Print("group /v1 handler")
}).UseToChild(func (c *brouter.Context){
    log.Print("end handler")
})


fasthttp.ListenAndServe(":8081", r.Init())
```
news_controller.go
```go
type NewsController struct{}
// match url '/v1/news/id' , method is GET 
func (p *NewsController) GETid(c *brouter.Context) {}
// match url '/v1/news/getLast' , method is POST
func (p *NewsController) POST_GetLast(c *brouter.Context) {}
// match url '/v1/news/find' ,method is ANY if func name is not start with POST,GET,PUT or DELETE
func (p *NewsController) Find(c *brouter.Context) {}
// match url '/v1/news/set' ,method is ANY
func (p *NewsController) Any_Set(c *brouter.Context) {}
```

## About Use
you can use too many handle to dispose a request, like middleware.
```go
r.Use(m1).Get("/user",h).Use(m2)
// m1 -> h -> m2
```
handle be run order by depth 
```go
r.Use(m1)
a:=r.Group("/api")
a.Use(m2)
a.Get("/user",h).Use(m3)
a.Use(m4)
// h is a child of a, a's handler always run first
// m1 -> m2 -> h -> m4 -> m3
```
you can use UerToChild set handle to he's child's finally
```go
r.Use(m1)
a:=r.Group("/api")
a.Use(m2)
a.Get("/user",h).Use(m3)
a.UseToChild(m4)
// m1 -> m2 -> h -> m3 -> m4
```

