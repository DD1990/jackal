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

// 增加或更新节点.
func (s *Storage) InsertOrUpdateNode(u *Node) error {
	columns := []string{"service_id", "name", "name_sha1", "type", "title", "description", "creator_id", "creation_date", "configuration" /*, "collection_id"*/}
	values := []interface{}{u.ServiceId, u.NodeName, u.NodeNameSha, u.Type, u.Title, u.Description, u.CreatorId, nowExpr, u.Configuration /*, u.CollectionId*/}

	suffix := "ON DUPLICATE KEY UPDATE service_id = ?, name = ?, name_sha1 = ?, type = ?, title = ?, description = ?, creator_id = ?, creation_date = NOW(), configuration = ?/*, collection_id = ?*/"
	suffixArgs := []interface{}{u.ServiceId, u.NodeName, u.NodeNameSha, u.Type, u.Title, u.Description, u.CreatorId, u.Configuration /*, u.CollectionId*/}

	q := sq.Insert("pubsub_nodes").
		Columns(columns...).
		Values(values...).
		Suffix(suffix, suffixArgs...)
	_, err := q.RunWith(s.db).Exec()
	return err
}

// 查询节点.
func (s *Storage) FetchNode(nodename string) (*Node, error) {
	q := sq.Select("node_id", "service_id", "name", "creator_id").
		From("pubsub_nodes").
		Where(sq.Eq{"name": nodename})

	var node Node
	err := s.scanNodeEntity(&node, q.RunWith(s.db).QueryRow())
	switch err {
	case nil:
		return &node, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

// 删除节点.
func (s *Storage) DeleteNode(username string) error {
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

// 判断节点是否存在.
func (s *Storage) NodeExists(username string) (bool, error) {
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

func (s *Storage) scanNodeEntity(nodeItem *Node, scanner rowScanner) error {
	if err := scanner.Scan(&nodeItem.NodeId, &nodeItem.ServiceId, &nodeItem.NodeName, &nodeItem.CreatorId /*, &nodeItem.CollectionId, &nodeItem.Configuration, &nodeItem.CreationDate, &nodeItem.CreatorId, &nodeItem.Description, &nodeItem.Title, &nodeItem.Type, &nodeItem.NodeNameSha, &nodeItem.NodeName*/); err != nil {
		return err
	}
	return nil
}

func (s *Storage) scanNodeEntities(scanner rowsScanner) ([]Node, error) {
	var ret []Node
	for scanner.Next() {
		var nodeItem Node
		if err := s.scanNodeEntity(&nodeItem, scanner); err != nil {
			return nil, err
		}
		ret = append(ret, nodeItem)
	}
	return ret, nil
}

/*************************************************Jid操作**************************************************************/
// 增加或更新Jid.
func (s *Storage) InsertOrUpdateJid(u *Jid) error {
	columns := []string{"jid", "jid_sha1"}
	values := []interface{}{u.Jid, u.JidSha}

	suffix := "ON DUPLICATE KEY UPDATE jid = ?, jid_sha1 = ?"
	suffixArgs := []interface{}{u.Jid, u.JidSha}

	q := sq.Insert("pubsub_jids").
		Columns(columns...).
		Values(values...).
		Suffix(suffix, suffixArgs...)
	_, err := q.RunWith(s.db).Exec()
	return err
}
func (s *Storage) FetchJid(jid string) (*Jid, error) {
	q := sq.Select("jid_id", "jid", "jid_sha1").
		From("pubsub_jids").
		Where(sq.Eq{"jid": jid})

	var jidItem Jid
	err := s.scanJidEntity(&jidItem, q.RunWith(s.db).QueryRow())
	switch err {
	case nil:
		return &jidItem, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}
func (s *Storage) FetchJidById(id uint64) (*Jid, error) {
	q := sq.Select("jid_id", "jid", "jid_sha1").
		From("pubsub_jids").
		Where(sq.Eq{"jid_id": id})

	var jidItem Jid
	err := s.scanJidEntity(&jidItem, q.RunWith(s.db).QueryRow())
	switch err {
	case nil:
		return &jidItem, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

func (s *Storage) scanJidEntity(jidItem *Jid, scanner rowScanner) error {
	if err := scanner.Scan(&jidItem.JidId, &jidItem.Jid, &jidItem.JidSha); err != nil {
		return err
	}
	return nil
}

func (s *Storage) scanJidEntities(scanner rowsScanner) ([]Jid, error) {
	var ret []Jid
	for scanner.Next() {
		var jidItem Jid
		if err := s.scanJidEntity(&jidItem, scanner); err != nil {
			return nil, err
		}
		ret = append(ret, jidItem)
	}
	return ret, nil
}

/*************************************************Subscribe操作**************************************************************/

// 增加或更新订阅.
func (s *Storage) InsertOrUpdateSubscribe(u *Subscribe) error {
	columns := []string{"node_id", "jid_id", "subscription", "subscription_id"}
	values := []interface{}{u.NodeId, u.JidId, u.Subscription, u.SubscriptionId}

	suffix := "ON DUPLICATE KEY UPDATE node_id = ?, jid_id = ?, subscription = ?, subscription_id = ?"
	suffixArgs := []interface{}{u.NodeId, u.JidId, u.Subscription, u.SubscriptionId}

	q := sq.Insert("pubsub_subscriptions").
		Columns(columns...).
		Values(values...).
		Suffix(suffix, suffixArgs...)
	_, err := q.RunWith(s.db).Exec()
	return err
}

//查询订阅数据
func (s *Storage) FetchSubscribe(nodeid uint64) (*[]Subscribe, error) {
	q := sq.Select("node_id", "jid_id", "subscription", "subscription_id").
		From("pubsub_subscriptions").
		Where(sq.Eq{"node_id": nodeid})

	//var subscribe []Subscribe
	rows, err := q.RunWith(s.db).Query()
	subscribe, err := s.scanSubscribeEntities(rows)
	switch err {
	case nil:
		return &subscribe, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

func (s *Storage) scanSubscribeEntity(subscribe *Subscribe, scanner rowScanner) error {
	if err := scanner.Scan(&subscribe.NodeId, &subscribe.JidId, &subscribe.Subscription, &subscribe.SubscriptionId); err != nil {
		return err
	}
	return nil
}

func (s *Storage) scanSubscribeEntities(scanner rowsScanner) ([]Subscribe, error) {
	var ret []Subscribe
	for scanner.Next() {
		var subscribe Subscribe
		if err := s.scanSubscribeEntity(&subscribe, scanner); err != nil {
			return nil, err
		}
		ret = append(ret, subscribe)
	}
	return ret, nil
}

/*************************************************Subscribe操作**************************************************************/

// 增加或更新节点.
func (s *Storage) InsertOrUpdateItem(u *Item) error {
	columns := []string{"node_id", "id", "id_sha1", "creation_date", "publisher_id", "update_date", "data"}
	values := []interface{}{u.NodeId, u.Id, u.IdSha1, nowExpr, u.PublisherId, nowExpr, u.Data}

	suffix := "ON DUPLICATE KEY UPDATE node_id = ?, id = ?, id_sha1 = ?, creation_date = NOW(), publisher_id = ?, update_date = NOW(), data = ?"
	suffixArgs := []interface{}{u.NodeId, u.Id, u.IdSha1, u.PublisherId, u.Data}

	q := sq.Insert("pubsub_items").
		Columns(columns...).
		Values(values...).
		Suffix(suffix, suffixArgs...)
	_, err := q.RunWith(s.db).Exec()
	return err
}
