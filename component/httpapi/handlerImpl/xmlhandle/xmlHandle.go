package xmlhandle

import (
	"fmt"
	"github.com/ortuman/jackal/component/httpapi/checkAuth"
	aer "github.com/ortuman/jackal/component/httpapi/errors"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/module"
	"github.com/ortuman/jackal/pool"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"io/ioutil"
	"net/http"
)

type XmlHandlerface struct {
	aer.ApiErr
}

//func Init() *XmlHandlerface {
//	return &XmlHandlerface{pool: pool.NewBufferPool()}
//}
func (c *XmlHandlerface) StreamHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("进入 StreamHandle()")
	defer r.Body.Close()
	con, _ := ioutil.ReadAll(r.Body) //获取post的数据
	bodyContent := string(con)

	if checkAuth.CheckBasicAuth(w, r) {
		/*解析xml，缺点：无法解析有换行的xml，容错率差*/
		pool := pool.NewBufferPool()
		buf := pool.Get()
		defer pool.Put(buf) //释放缓冲到缓冲池
		buf.WriteString(bodyContent)
		parser := xmpp.NewParser(buf, xmpp.DefaultMode, 0)
		elem, err := parser.ParseElement()

		if err != nil {
			log.Error(err)
			w.Write([]byte("failed : 请求成功，处理失败")) //请求成功，处理失败
			return
		} else {
			c.processXElement(elem)
			w.Write([]byte("success : 请求成功，处理成功")) //请求成功，处理成功
		}
		return
	}

	c.HandleErr(&w, aer.AUTH_ERR_CODE, nil)
}

//func (c *XmlHandlerface) ProcessXElement(stanza xmpp.XElement) {
//	c.processXElement(stanza)
//}

func (c *XmlHandlerface) processXElement(stanza xmpp.XElement) {
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

func (c *XmlHandlerface) processIQ(iq *xmpp.IQ) {
	return
}

func (c *XmlHandlerface) processPresence(presence *xmpp.Presence) {
	//toJID := presence.ToJID()
	//if toJID.IsFullWithUser() {
	//	c.chatRoom.ProcessPresence(presence, stm)
	//}
	return
}

func (c *XmlHandlerface) processMessage(message *xmpp.Message) {

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
