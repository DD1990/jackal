/*
 * HTTP-API组件.
 * See the LICENSE file for more information.
 */

package httpapi

import (
	"fmt"
	"github.com/ortuman/jackal/component/httpapi/checkAuth"
	"github.com/ortuman/jackal/component/httpapi/handler"
	"github.com/ortuman/jackal/component/httpapi/handlerImpl/model"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/module"
	"github.com/ortuman/jackal/pool"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"io/ioutil"
	"net/http"
	"strings"
)

const mailboxSize = 2048

//const mucServiceName = "httpApi"

//HttpApi 结构体
type HttpApi struct {
	cfg        *Config
	pool       *pool.BufferPool
	doneCh     chan chan bool
	actorCh    chan func()
	shutdownCh <-chan struct{}
}

//生成 HttpApi 引用
func New(cfg *Config, shutdownCh <-chan struct{}) *HttpApi {
	c := &HttpApi{
		cfg:        cfg,
		pool:       pool.NewBufferPool(),
		doneCh:     make(chan chan bool),
		actorCh:    make(chan func(), mailboxSize),
		shutdownCh: shutdownCh,
	}
	//新启go程，监听http请求
	go c.SrtartListen()
	go c.loop()

	return c
}

func (c *HttpApi) loop() {
	for {
		select {
		case f := <-c.actorCh:
			f()
		case <-c.shutdownCh:
			//TODO(lxf)  请求结束处理
			return
		}
	}
}

func (c *HttpApi) Host() string {
	return c.cfg.Host
}

//新启go程，监听http请求
func (c *HttpApi) SrtartListen() {

	handler.Intli()
	//服务端
	//方式一：原生http

	//用户管理
	http.HandleFunc(model.RESTUSER, handler.Instance().UserManage)

	//消息群发
	http.HandleFunc(model.RESTSTREAM, handler.Instance().StreamHandle)
	//http.HandleFunc(model.RESTSTREAM, c.ProcessXMPP)

	//多域管理
	http.HandleFunc(model.RESTADHOC, handler.Instance().DomainManage)

	//节点管理
	http.HandleFunc(model.RESTPUBSUB, handler.Instance().PubSubManage)

	addrStr := fmt.Sprintf("%s%s%s", c.cfg.Host, ":", c.cfg.Port)
	http.ListenAndServe(addrStr, nil)

	//方式二：mux框架  - "github.com/gorilla/mux" 包
	/*	router := mux.NewRouter().StrictSlash(true)
		router.HandleFunc("/", Index)
		router.HandleFunc("/todos", TodoIndex)
		router.HandleFunc("/todos/{todoId}", TodoShow)

		log.Fatal(http.ListenAndServe(":8080", router))
	*/

	//方式三：Gin框架 - "github.com/gin-gonic/gin" 包
	/*	router := gin.Default()
		router.GET("/welcome", func(c *gin.Context) {
		   firstname := c.DefaultQuery("firstname", "Guest")
		   lastname := c.Query("lastname")

		   c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
		})
		router.Run()
	*/
}

func (c *HttpApi) ProcessXMPP(w http.ResponseWriter, r *http.Request) {

	fmt.Println("进入 ProcessXMPP()")
	defer r.Body.Close()
	con, _ := ioutil.ReadAll(r.Body) //获取post的数据
	bodyContent := string(con)

	fmt.Println(strings.TrimSpace(bodyContent))

	if checkAuth.CheckBasicAuth(w, r) {
		buf := c.pool.Get()
		defer c.pool.Put(buf) //释放缓冲到缓冲池
		buf.WriteString(bodyContent)

		parser := xmpp.NewParser(buf, xmpp.DefaultMode, 0)
		elem, err := parser.ParseElement()
		if err != nil {
			log.Error(err)
			w.Write([]byte("failed")) //请求成功，处理失败
			return
		} else {
			c.ProcessXElement(elem, nil)
			w.Write([]byte("success")) //请求成功，处理成功
		}
		return
	}

	w.Header().Set("WWW-Authenticate", `Basic realm="MY REALM"`)
	w.WriteHeader(401)
	w.Write([]byte("401 Unauthorized\n"))
}

func (c *HttpApi) ProcessStanza(stanza xmpp.Stanza, stm stream.C2S) {
	return
}

func (c *HttpApi) ProcessXElement(stanza xmpp.XElement, stm stream.C2S) {
	c.actorCh <- func() {
		c.processXElement(stanza)
	}
}

func (c *HttpApi) processXElement(stanza xmpp.XElement) {
	fromJid, _ := jid.NewWithString(stanza.From(), false)
	toJid, _ := jid.NewWithString(stanza.To(), false)

	switch stanza.Name() {
	case "iq":
		iq, err := xmpp.NewIQFromElement(stanza, fromJid, toJid)
		if err == nil {
			c.processIQ(iq)
		}
	case "presence":
		presence, err := xmpp.NewPresenceFromElement(stanza, fromJid, toJid)
		if err == nil {
			c.processPresence(presence)
		}
	case "message":
		message, err := xmpp.NewMessageFromElement(stanza, fromJid, toJid)
		if err == nil {
			c.processMessage(message)
		}
	}
}

func (c *HttpApi) processIQ(iq *xmpp.IQ) {
	return
}

func (c *HttpApi) processPresence(presence *xmpp.Presence) {
	//toJID := presence.ToJID()
	//if toJID.IsFullWithUser() {
	//	c.chatRoom.ProcessPresence(presence, stm)
	//}
	return
}

func (c *HttpApi) processMessage(message *xmpp.Message) {

	//判断如果成员在线，则发送
	if stms := router.UserStreams(message.ToJID().Node()); stms != nil {
		stm := stms[0]
		stm.SendElement(message)
	} else { //如果成员不在线，则发送离线消息
		//TODO(lxf) 一对一聊天可以，广播类型的离线消息未处理
		offlineServer := module.Modules().Offline
		offlineServer.ArchiveMessage(message)
	}
	return
}
