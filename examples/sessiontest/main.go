package main

import (
	"fmt"
	e "github.com/frank418/entropy"
	"log"
	"math/rand"
	"strconv"
)

type CookieHandler struct {
	e.Handler
}

func (c *CookieHandler) Get() {
	rndId := strconv.Itoa(rand.Intn(100))
	uname := "User" + rndId
	upwd := "password" + rndId
	log.Println(uname, upwd)
	c.Session.SetSession("uid", uname)
	c.Session.SetSession("upwd", upwd)
	c.Session.Flush()
	c.RenderText("complete")
}

type CookieTestHandler struct {
	e.Handler
}

func (c *CookieTestHandler) Get() {
	ret := fmt.Sprintf("uname%v upwd%v", c.Session.GetSession("uid"), c.Session.GetSession("upwd"))
	//log.Printf("uname%v upwd%v", c.Session.GetSession("uid"), c.Session.GetSession("upwd"))
	c.RenderText(ret)
}

func main() {
	app := e.NewApplication("appconf.json")
	app.AddHandler("/", "eName", "cName", &CookieHandler{})
	app.AddHandler("/r", "eName1", "cName1", &CookieTestHandler{})
	app.Go("", 4000)
}
