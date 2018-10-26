/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package sql

import (
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/ortuman/jackal/component/muc/model/roommodel"
)

// 增加或更新房间信息.
func (s *Storage) InsertOrUpdateRoomItem(ri *roommodel.RoomItem) error {
	err := s.inTransaction(func(tx *sql.Tx) error {
		q := sq.Insert("rooms").
			Columns("roomname", "roomserver", "password", "ispravite", "roomtype", "update_at", "create_at").
			Values(ri.Roomname, ri.Roomserver, ri.Password, ri.Ispravite, ri.Roomtype, nowExpr, nowExpr).
			Suffix("ON DUPLICATE KEY UPDATE roomname = ?, roomserver = ?, password = ?, ispravite = ?, roomtype = ?, update_at = NOW(), create_at = NOW()", ri.Roomname, ri.Roomserver, ri.Password, ri.Ispravite, ri.Roomtype)

		_, err := q.RunWith(tx).Exec()
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

// 删除一个房间
func (s *Storage) DeleteRoomItem(username, jid string) error {
	err := s.inTransaction(func(tx *sql.Tx) error {
		q := sq.Insert("roster_versions").
			Columns("username", "created_at", "updated_at").
			Values(username, nowExpr, nowExpr).
			Suffix("ON DUPLICATE KEY UPDATE ver = ver + 1, last_deletion_ver = ver, updated_at = NOW()")

		if _, err := q.RunWith(tx).Exec(); err != nil {
			return err
		}
		_, err := sq.Delete("roster_items").
			Where(sq.And{sq.Eq{"username": username}, sq.Eq{"jid": jid}}).
			RunWith(tx).Exec()
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

// 查询所有房间.
func (s *Storage) FetchRoomItems(domainname string) ([]roommodel.RoomItem, error) {
	q := sq.Select("roomname", "roomserver").
		From("rooms").
		Where(sq.Eq{"roomserver": domainname}).
		OrderBy("create_at DESC")

	rows, err := q.RunWith(s.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := s.scanRoomItemEntities(rows)
	if err != nil {
		return nil, err
	}
	//ver, err := "",nil/*s.fetchRoomVer(username)*/
	if err != nil {
		return nil, err
	}
	return items, nil
}

// FetchRoomItem retrieves from storage a roster item entity.
/*func (s *Storage) FetchRoomItem(username, jid string) (*roommodel.RoomItem, error) {
	q := sq.Select("username", "jid", "name", "subscription", "`groups`", "ask", "ver").
		From("roster_items").
		Where(sq.And{sq.Eq{"username": username}, sq.Eq{"jid": jid}})

	var ri roommodel.RoomItem
	err := s.scanRoomItemEntity(&ri, q.RunWith(s.db).QueryRow())
	switch err {
	case nil:
		return &ri, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}
*/

func (s *Storage) scanRoomItemEntity(ri *roommodel.RoomItem, scanner rowScanner) error {
	//	var groups string
	if err := scanner.Scan(&ri.Roomname, &ri.Roomserver); err != nil {
		return err
	}
	//	ri.Groups = strings.Split(groups, ";")
	return nil
}

func (s *Storage) scanRoomItemEntities(scanner rowsScanner) ([]roommodel.RoomItem, error) {
	var ret []roommodel.RoomItem
	for scanner.Next() {
		var ri roommodel.RoomItem
		if err := s.scanRoomItemEntity(&ri, scanner); err != nil {
			return nil, err
		}
		ret = append(ret, ri)
	}
	return ret, nil
}
