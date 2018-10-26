/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package httpupload

import (
	"github.com/ortuman/jackal/module/xep0030"
	"github.com/ortuman/jackal/stream"
	"github.com/ortuman/jackal/xmpp"
)

const mailboxSize = 2048

const httpUploadServiceName = "HTTP File Upload"

const httpUploadFeature = "urn:xmpp:http:upload:0"

type HttpUpload struct {
	cfg        *Config
	discoInfo  *xep0030.DiscoInfo
	actorCh    chan func()
	shutdownCh <-chan struct{}
}

func New(cfg *Config, discoInfo *xep0030.DiscoInfo, shutdownCh <-chan struct{}) *HttpUpload {
	c := &HttpUpload{
		cfg:        cfg,
		discoInfo:  discoInfo,
		actorCh:    make(chan func(), mailboxSize),
		shutdownCh: shutdownCh,
	}
	c.registerDiscoInfo()
	go c.loop()
	return c
}

func (c *HttpUpload) Host() string {
	return c.cfg.Host
}

func (c *HttpUpload) ProcessStanza(stanza xmpp.Stanza, stm stream.C2S) {
	c.actorCh <- func() {
		c.processStanza(stanza, stm)
	}
}

func (c *HttpUpload) loop() {
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

func (c *HttpUpload) processStanza(stanza xmpp.XElement, stm stream.C2S) {
	switch stanza := stanza.(type) {
	case *xmpp.IQ:
		c.processIQ(stanza, stm)
	}
}

func (c *HttpUpload) processIQ(iq *xmpp.IQ, stm stream.C2S) {
	if c.discoInfo.MatchesIQ(iq) {
		c.discoInfo.ProcessIQ(iq, stm)
		return
	}
}

func (c *HttpUpload) registerDiscoInfo() {
	c.discoInfo.RegisterServerItem(xep0030.Item{Jid: c.Host(), Name: httpUploadServiceName})
	c.discoInfo.RegisterProvider(c.Host(), &uploadInfoProvider{})
}

func (c *HttpUpload) unregisterDiscoInfo() {
	c.discoInfo.UnregisterServerItem(xep0030.Item{Jid: c.Host(), Name: httpUploadServiceName})
	c.discoInfo.UnregisterProvider(c.Host())
}
