/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package httpupload

import (
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

type uploadInfoProvider struct{}

func (ip *uploadInfoProvider) Identities(toJID, fromJID *jid.JID, node string) []xep0030.Identity {
	return []xep0030.Identity{
		{Category: "store", Type: "file", Name: httpUploadServiceName},
	}
}

func (ip *uploadInfoProvider) Items(toJID, fromJID *jid.JID, node string) ([]xep0030.Item, *xmpp.StanzaError) {
	return nil, nil
}

func (ip *uploadInfoProvider) Features(toJID, fromJID *jid.JID, node string) ([]xep0030.Feature, *xmpp.StanzaError) {
	return []xep0030.Feature{httpUploadFeature}, nil
}
