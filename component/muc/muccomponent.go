/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package muc

import (
	"github.com/ortuman/jackal/component/muc/xep0045"
	"github.com/ortuman/jackal/module/xep0030"
	"strings"

	//TODO(lxf) 作为组件来讲，都应该使用组件自身的依赖，因为组件本身单独开发，是不知道也不需要依赖主服务本身的实现
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
)

const mailboxSize = 2048

const mucServiceName = "mucServer"

const mucFeature = "http://jabber.org/protocol/muc"

type Muc struct {
	cfg        Config
	discoInfo  *xep0030.DiscoInfo
	chatRoom   *xep0045.ChatRoom
	actorCh    chan func()
	shutdownCh <-chan struct{}
}

func New(cfg Config, discoInfo *xep0030.DiscoInfo, shutdownCh <-chan struct{}) *Muc {
	c := &Muc{
		cfg:       cfg,
		discoInfo: discoInfo,
		// TODO(lxf) 初始化群聊charRoom？
		chatRoom:   xep0045.New(shutdownCh),
		actorCh:    make(chan func(), mailboxSize),
		shutdownCh: shutdownCh,
	}
	c.registerDiscoInfo()
	go c.loop()

	return c
}

func (c *Muc) Host() string {
	return c.cfg.Host
}

func (c *Muc) ProcessStanza(stanza xmpp.Stanza, stm stream.C2S) {
	c.actorCh <- func() {
		c.processStanza(stanza, stm)
	}
}

func (c *Muc) loop() {
	for {
		select {
		case f := <-c.actorCh:
			f()
		case <-c.shutdownCh:
			c.unregisterDiscoInfo()
			return
		}
	}
}

func (c *Muc) processStanza(stanza xmpp.XElement, stm stream.C2S) {
	switch stanza := stanza.(type) {
	case *xmpp.IQ:
		c.processIQ(stanza, stm)
	case *xmpp.Presence:
		c.processPresence(stanza, stm)
	case *xmpp.Message:
		c.processMessage(stanza, stm)
	}
}

func (c *Muc) processIQ(iq *xmpp.IQ, stm stream.C2S) {
	// TODO(lxf) 如果是disco的请求服务，交由disco处理；如果不是，再路由一次。（TODO：此判断可放在上一层in.go中）
	if c.discoInfo.MatchesIQ(iq) {

		//start - add - lxf - 20181025
		//判断用户所在虚拟域和muc所在虚拟域是否相同
		//TODO(lxf) 当domain名称中间有.时，会出问题，此处需要修改，例如去掉muc.
		if strings.Split(iq.ToJID().Domain(), ".")[1] != iq.FromJID().Domain() {
			stm.SendElement(iq.RemoteServerNotFoundError())
			return
		}
		//start - add - lxf - 20181025
		c.discoInfo.ProcessIQ(iq, stm)
		return
	} else {
		//TODO(lxf):muc相关的IQ请求处理
		c.chatRoom.ProcessIQ(iq, stm)
	}
}

func (c *Muc) processPresence(presence *xmpp.Presence, stm stream.C2S) {
	//TODO(lxf) 此处增加路由或判断
	toJID := presence.ToJID()
	if toJID.IsFullWithUser() {
		c.chatRoom.ProcessPresence(presence, stm)
	}
	return
}

func (c *Muc) processMessage(message *xmpp.Message, stm stream.C2S) {
	//TODO(lxf) 此处增加路由或判断
	c.chatRoom.ProcessMessage(message, stm)
	return
}

func (c *Muc) registerDiscoInfo() {
	c.discoInfo.RegisterServerItem(xep0030.Item{Jid: c.Host(), Name: mucServiceName + c.Host()})
	c.discoInfo.RegisterProvider(c.Host(), &mucInfoProvider{})
	/*
		c.discoInfo.RegisterServerItem(xep0030.Item{Jid: c.Host() + ".localhost2", Name: mucServiceName + ".localhost2"})
		c.discoInfo.RegisterProvider(c.Host()+".localhost2", &mucInfoProvider{})

		c.discoInfo.RegisterServerItem(xep0030.Item{Jid: c.Host() + ".localhost3", Name: mucServiceName + ".localhost3"})
		c.discoInfo.RegisterProvider(c.Host()+".localhost3", &mucInfoProvider{})

		c.discoInfo.RegisterServerItem(xep0030.Item{Jid: c.Host() + ".localhost4", Name: mucServiceName + ".localhost4"})
		c.discoInfo.RegisterProvider(c.Host()+".localhost4", &mucInfoProvider{})*/
}

func (c *Muc) unregisterDiscoInfo() {
	c.discoInfo.UnregisterServerItem(xep0030.Item{Jid: c.Host(), Name: mucServiceName})
	c.discoInfo.UnregisterProvider(c.Host())
	/*
		c.discoInfo.UnregisterServerItem(xep0030.Item{Jid: c.Host() + ".localhost2", Name: mucServiceName + ".localhost2"})
		c.discoInfo.UnregisterProvider(c.Host() + ".localhost2")

		c.discoInfo.UnregisterServerItem(xep0030.Item{Jid: c.Host() + ".localhost3", Name: mucServiceName + ".localhost3"})
		c.discoInfo.UnregisterProvider(c.Host() + ".localhost3")

		c.discoInfo.UnregisterServerItem(xep0030.Item{Jid: c.Host() + ".localhost4", Name: mucServiceName + ".localhost4"})
		c.discoInfo.UnregisterProvider(c.Host() + ".localhost4")*/
}
