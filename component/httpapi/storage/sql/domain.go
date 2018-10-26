/*
 * 用户持久化操作.
 * See the LICENSE file for more information.
 */

package sql

import (
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	. "github.com/ortuman/jackal/component/httpapi/handlerImpl/model"
)

// 增加或更新虚拟域.
func (s *Storage) InsertOrUpdateDomain(u *Domain) error {
	columns := []string{"service_jid", "service_jid_sha1"}
	values := []interface{}{u.ServiceJid, u.ServiceJidSha}

	suffix := "ON DUPLICATE KEY UPDATE service_jid = ?, service_jid_sha1 = ?"
	suffixArgs := []interface{}{u.ServiceJid, u.ServiceJidSha}

	q := sq.Insert("pubsub_service_jids").
		Columns(columns...).
		Values(values...).
		Suffix(suffix, suffixArgs...)
	_, err := q.RunWith(s.db).Exec()
	return err
}

// 查询虚拟域.
func (s *Storage) FetchDomainAll() (*[]Domain, error) {
	q := sq.Select("service_id", "service_jid", "service_jid_sha1").
		From("pubsub_service_jids")

	//var domItem []Domain
	rows, err := q.RunWith(s.db).Query()
	domItem, err := s.scanDomainEntities(rows)
	switch err {
	case nil:
		return &domItem, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

// 查询虚拟域.
func (s *Storage) FetchDomain(name string) (*Domain, error) {
	q := sq.Select("service_id", "service_jid", "service_jid_sha1").
		From("pubsub_service_jids").
		Where(sq.Eq{"service_jid": name})

	var domItem Domain
	err := s.scanDomainEntity(&domItem, q.RunWith(s.db).QueryRow())
	switch err {
	case nil:
		return &domItem, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

// 删除用户.
func (s *Storage) DeleteDomain(username string) error {
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
func (s *Storage) DomainExists(username string) (bool, error) {
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

func (s *Storage) scanDomainEntity(domItem *Domain, scanner rowScanner) error {
	if err := scanner.Scan(&domItem.ServiceId, &domItem.ServiceJid, &domItem.ServiceJidSha); err != nil {
		return err
	}
	return nil
}

func (s *Storage) scanDomainEntities(scanner rowsScanner) ([]Domain, error) {
	var ret []Domain
	for scanner.Next() {
		var domItem Domain
		if err := s.scanDomainEntity(&domItem, scanner); err != nil {
			return nil, err
		}
		ret = append(ret, domItem)
	}
	return ret, nil
}
