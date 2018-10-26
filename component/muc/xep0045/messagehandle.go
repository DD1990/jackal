package xep0045

import (
	"errors"
	"github.com/ortuman/jackal/component/muc/model/roommodel"
	"github.com/ortuman/jackal/router"
	"github.com/ortuman/jackal/xmpp"
	"strings"
)

//发送群聊信息
//TODO(lxf) 发送放在manager中，此处制作message的封装和解析
//TODO(lxf) 邀请成员进群也是message消息，此处需判断，不然会造成空指针
func sendGroupChatMessage(roomusersitms []roommodel.RoommemberItem, message *xmpp.Message) error {
	for _, user := range roomusersitms {
		if &user != nil && user.Membername != "" {
			pre := xmpp.NewElementName("message")
			pre.SetAttribute("from", message.ToJID().String()+"/"+message.FromJID().Node())
			pre.SetAttribute("to", user.Membername+"@"+message.FromJID().Domain()+"/"+message.FromJID().Resource())
			pre.SetAttribute("type", "groupchat")
			item := xmpp.NewElementName("body")
			// TODO(lxf) 改为敏感词汇配置表或配置文件;也可开发敏感词汇处理组件扩展服务器
			if message.Elements().Child("body") == nil {
				return errors.New("空消息")
			}
			if text := message.Elements().Child("body").Text(); strings.Contains(text, "你个混蛋") {
				text = strings.Replace(text, "你个混蛋", "****", -1)
				item.SetText(text)
			} else {
				item.SetText(text)
			}
			pre.AppendElement(item)

			//判断如果成员在线，则发送
			//TODO(lxf) 如果成员不在线，则发送离线消息
			if stms := router.UserStreams(user.Membername); stms != nil {
				stm := stms[0]
				stm.SendElement(pre)
			}
		}
	}
	return nil
}

//无权限错误
func sendNoPrivilegesErr(rm *roommodel.RoommemberItem, message *xmpp.Message) error {

	/*
		<message type="error" id="Nr1X2-137" xmlns="jabber:client" to="tig_test2@test/Spark" from="roomtest1@muc.test">
			<body>22</body>
			<x xmlns="jabber:x:event">
				<offline/>
				<delivered/>
				<displayed/>
				<composing/>
			</x>
			<error type="auth" code="403">
				<forbidden xmlns="urn:ietf:params:xml:ns:xmpp-stanzas"/>
				<text xmlns="urn:ietf:params:xml:ns:xmpp-stanzas" xml:lang="en">Insufficient privileges to send groupchat message.</text>
			</error>
		</message>
	*/

	mes := xmpp.NewElementNamespace("message", "jabber:client")
	mes.SetAttribute("id", message.ID())
	mes.SetAttribute("from", message.ToJID().String())
	mes.SetAttribute("to", message.FromJID().String())
	mes.SetAttribute("type", "error")
	body := xmpp.NewElementName("body")
	body.SetText(message.Elements().Child("body").Text())
	x := xmpp.NewElementNamespace("x", "jabber:x:event")
	offline := xmpp.NewElementName("offline")
	delivered := xmpp.NewElementName("delivered")
	displayed := xmpp.NewElementName("displayed")
	composing := xmpp.NewElementName("composing")
	error := xmpp.NewElementName("error")
	error.SetAttribute("type", "auth")
	error.SetAttribute("code", "403")
	forbidden := xmpp.NewElementNamespace("forbidden", "urn:ietf:params:xml:ns:xmpp-stanzas")
	text := xmpp.NewElementNamespace("text", "urn:ietf:params:xml:ns:xmpp-stanzas")
	text.SetAttribute("xml:lang", "en")
	text.SetText("Insufficient privileges to send groupchat message.")

	x.AppendElement(offline)
	x.AppendElement(delivered)
	x.AppendElement(displayed)
	x.AppendElement(composing)
	error.AppendElement(forbidden)
	error.AppendElement(text)

	mes.AppendElement(body)
	mes.AppendElement(x)
	mes.AppendElement(error)

	//判断如果成员在线，则发送
	//TODO(lxf) 如果成员不在线，则发送离线消息
	if stms := router.UserStreams(rm.Membername); stms != nil {
		stm := stms[0]
		stm.SendElement(mes)
	}

	return nil
}

//发送群聊历史记录
func sendGroupChatHistoryMessage() error {
	return nil
}

//查看历史记录
func viewHistoryMessage() error {
	return nil
}
