/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package sql

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/ortuman/jackal/xmpp"
)
func (s *Storage) InsertOrUpdateRoomcfgItem(iq xmpp.XElement, roomname string) error{
	q := sq.Insert("roomcfg").
		Columns("roomname", "cfgcontent").
		Values(roomname, iq.String())
	_, err := q.RunWith(s.db).Exec()
	return err
}
// FetchRoomItem retrieves from storage a roster item entity.
func (s *Storage) FetchRoomcfgItem(iq *xmpp.IQ, roomname string) (*xmpp.XElement, error) {
	var cfgcontent string
	q := sq.Select("cfgcontent").
		From("roomcfg").
		Where(sq.Eq{"roomname": roomname})

	err := q.RunWith(s.db).QueryRow().Scan(&cfgcontent)//单记录查询
	if err != nil {
		return nil, err
	}

	buf := s.pool.Get()
	defer s.pool.Put(buf)//释放缓冲到缓冲池

	buf.WriteString(cfgcontent)

	parser := xmpp.NewParser(buf, xmpp.DefaultMode, 0)
	elem, err := parser.ParseElement()
	if err != nil {
		return nil, err
	}

/*	var msgs []*xmpp.Message
	for _, el := range elems {
		fromJID, _ := jid.NewWithString(el.From(), true)
		toJID, _ := jid.NewWithString(el.To(), true)
		msg, err := xmpp.NewMessageFromElement(el, fromJID, toJID)  //TODO(lxf) 重要！！！！！
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}*/
	return &elem, nil
}