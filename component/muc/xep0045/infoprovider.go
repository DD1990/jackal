/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package xep0045

import (
	"github.com/ortuman/jackal/component/muc/storage"
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

type RoomInfoProvider struct{}

// room服务的身份定义
func (ip *RoomInfoProvider) Identities(toJID, fromJID *jid.JID, node string) []xep0030.Identity {
	// TODO(lxf) Name: "A Dark Cave" 改为去数据或内存查询的变量
	return []xep0030.Identity{
		{Category: "conference", Type: "test", Name: "A Dark Cave"},
	}
}

// room服务的条目集合（即：房间成员集合）
func (ip *RoomInfoProvider) Items(toJID, fromJID *jid.JID, node string) ([]xep0030.Item, *xmpp.StanzaError) {

	var itms []xep0030.Item
	roomitms, err := storage.Instance().FetchRoomItems(toJID.Domain())
	if err == nil {
		for _, roomitem := range roomitms {
			itms = append(itms, xep0030.Item{roomitem.Roomname + "@" + roomitem.Roomserver, roomitem.Roomname, ""})
		}
	}
	return itms, nil
}

// room服务的特性集合
func (ip *RoomInfoProvider) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	var features []xep0030.Feature
	roomitms, err := storage.Instance().FetchRoomItems(toJID.Domain())
	if err == nil {
		for _, roomitem := range roomitms {
			if roomitem.Ispravite {
				features = []xep0030.Feature{passwordprotectedFeature, hiddenFeature, temporaryFeature, openFeature, unmoderatedFeature, nonanonymousFeature}
			} else {
				features = []xep0030.Feature{hiddenFeature, temporaryFeature, openFeature, unmoderatedFeature, nonanonymousFeature}
			}
		}
	}
	return features, nil
}
