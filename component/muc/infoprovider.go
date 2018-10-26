/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package muc

import (
	"github.com/ortuman/jackal/component/muc/storage"
	"github.com/ortuman/jackal/component/muc/xep0045"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

type mucInfoProvider struct{}

// muc服务的身份定义
func (ip *mucInfoProvider) Identities(toJID, fromJID *jid.JID, node string) []xep0030.Identity {
	//Bare类型查询房间特性
	if toJID.IsBare() {
		var roomInfoProvider xep0045.RoomInfoProvider
		return roomInfoProvider.Identities(toJID, fromJID, node)
	}
	return []xep0030.Identity{
		{Category: "conference", Type: "XS Chat Service", Name: mucServiceName},
	}
}

// muc服务的条目集合（即：房间集合）
func (ip *mucInfoProvider) Items(toJID, fromJID *jid.JID, node string) ([]xep0030.Item, *xmpp.StanzaError) {

	var itms []xep0030.Item

	roomitms, err := storage.Instance().FetchRoomItems(toJID.Domain())
	if err == nil {
		for _, roomitem := range roomitms {
			itms = append(itms, xep0030.Item{roomitem.Roomname + "@" + roomitem.Roomserver, roomitem.Roomname, ""})
		}
	}
	return itms, nil
}

// muc服务的特性集合
func (ip *mucInfoProvider) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	//Bare类型查询房间特性
	if toJID.IsBare() {
		var roomInfoProvider xep0045.RoomInfoProvider
		return roomInfoProvider.Features(toJID, fromJID, node)
	}
	return []xep0030.Feature{mucFeature}, nil
}
