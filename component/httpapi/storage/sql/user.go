/*
 * 用户持久化操作.
 * See the LICENSE file for more information.
 */

package sql

import (
	"database/sql"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/ortuman/jackal/model"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
)

// 增加或更新用户.
func (s *Storage) InsertOrUpdateUser(u *model.User) error {
	var presenceXML string
	if u.LastPresence != nil {
		buf := s.pool.Get()
		u.LastPresence.ToXML(buf, true)
		presenceXML = buf.String()
		s.pool.Put(buf)
	}
	columns := []string{"username", "password", "domain", "updated_at", "created_at"}
	values := []interface{}{u.Username, u.Password, u.Domain, nowExpr, nowExpr}

	if len(presenceXML) > 0 {
		columns = append(columns, []string{"last_presence", "last_presence_at"}...)
		values = append(values, []interface{}{presenceXML, nowExpr}...)
	}
	var suffix string
	var suffixArgs []interface{}
	if len(presenceXML) > 0 {
		suffix = "ON DUPLICATE KEY UPDATE password = ?, last_presence = ?, last_presence_at = NOW(), updated_at = NOW()"
		suffixArgs = []interface{}{u.Password, presenceXML}
	} else {
		suffix = "ON DUPLICATE KEY UPDATE password = ?, updated_at = NOW()"
		suffixArgs = []interface{}{u.Password}
	}
	q := sq.Insert("users").
		Columns(columns...).
		Values(values...).
		Suffix(suffix, suffixArgs...)
	_, err := q.RunWith(s.db).Exec()
	return err
}

// 查询用户.
func (s *Storage) FetchUser(username string) (*model.User, error) {
	q := sq.Select("username", "password", "domain", "last_presence", "last_presence_at").
		From("users").
		Where(sq.Eq{"username": username})

	var presenceXML string
	var presenceAt time.Time
	var usr model.User

	err := q.RunWith(s.db).QueryRow().Scan(&usr.Username, &usr.Password, &usr.Domain, &presenceXML, &presenceAt)
	switch err {
	case nil:
		if len(presenceXML) > 0 {
			parser := xmpp.NewParser(strings.NewReader(presenceXML), xmpp.DefaultMode, 0)
			if lastPresence, err := parser.ParseElement(); err != nil {
				return nil, err
			} else {
				fromJID, _ := jid.NewWithString(lastPresence.From(), true)
				toJID, _ := jid.NewWithString(lastPresence.To(), true)
				usr.LastPresence, _ = xmpp.NewPresenceFromElement(lastPresence, fromJID, toJID)
				usr.LastPresenceAt = presenceAt
			}
		}
		return &usr, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

// 删除用户.
func (s *Storage) DeleteUser(username string) error {
	return s.inTransaction(func(tx *sql.Tx) error {
		var err error
		_, err = sq.Delete("offline_messages").Where(sq.Eq{"username": username}).RunWith(tx).Exec()
		if err != nil {
			return err
		}
		_, err = sq.Delete("roster_items").Where(sq.Eq{"username": username}).RunWith(tx).Exec()
		if err != nil {
			return err
		}
		_, err = sq.Delete("roster_versions").Where(sq.Eq{"username": username}).RunWith(tx).Exec()
		if err != nil {
			return err
		}
		_, err = sq.Delete("private_storage").Where(sq.Eq{"username": username}).RunWith(tx).Exec()
		if err != nil {
			return err
		}
		_, err = sq.Delete("vcards").Where(sq.Eq{"username": username}).RunWith(tx).Exec()
		if err != nil {
			return err
		}
		_, err = sq.Delete("users").Where(sq.Eq{"username": username}).RunWith(tx).Exec()
		if err != nil {
			return err
		}
		return nil
	})
}

// 判断用户是否存在.
func (s *Storage) UserExists(username string) (bool, error) {
	q := sq.Select("COUNT(*)").From("users").Where(sq.Eq{"username": username})
	var count int
	err := q.RunWith(s.db).QueryRow().Scan(&count)
	switch err {
	case nil:
		return count > 0, nil
	default:
		return false, err
	}
}
