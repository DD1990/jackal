package sql

import (
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/ortuman/jackal/component/muc/model/roommodel"
)

//房间成员插入或更新
func (s *Storage) InsertOrUpdateRoommemberItem(rm *roommodel.RoommemberItem) error{
	err := s.inTransaction(func(tx *sql.Tx) error {
		q := sq.Insert("roommembers").
			Columns("membername","roomname", "jid", "isblack",  "role", "update_at", "create_at").
			Values(rm.Membername,rm.Roomname,rm.Jid,rm.Isblack,rm.Role, nowExpr, nowExpr).
			Suffix("ON DUPLICATE KEY UPDATE membername = ?, roomname = ?, jid = ?, isblack = ?, role = ?, update_at = NOW(), create_at = NOW()",
			rm.Membername,rm.Roomname,rm.Jid,rm.Isblack,rm.Role)

		_, err := q.RunWith(tx).Exec()
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

//房间成员批量插入或更新
func (s *Storage) InsertOrUpdateRoommemberItems(rm *[]roommodel.RoommemberItem) error{
	for _,item := range *rm{
		if &item != nil{
			s.InsertOrUpdateRoommemberItem(&item)
		}
	}
	return nil
}

//查询某房间所有房间成员
func (s *Storage) FetchRoommemberItems(arg interface{}) ([]roommodel.RoommemberItem, error){
	q := sq.Select("membername", "roomname","jid","isblack","role").From("roommembers").Where(sq.Eq{"roomname": arg}).OrderBy("create_at DESC")
/*	switch arg {
		case "membername":  //成员名称
			q.Where(sq.Eq{"membername": arg}).OrderBy("create_at DESC")
		case "roomname":    //房间名称
			q.Where(sq.Eq{"roomname": arg}).OrderBy("create_at DESC")
		case "jid": 		//成员jid
			q.Where(sq.Eq{"jid": arg}).OrderBy("create_at DESC")
		case "isblack":     //成员是否在黑名单
			q.Where(sq.Eq{"isblack": arg}).OrderBy("create_at DESC")
		case "role":        //成员角色
			q.Where(sq.Eq{"role": arg}).OrderBy("create_at DESC")
		case "",nil:        //参数为空
			q.OrderBy("create_at DESC")
		default:
			q.OrderBy("create_at DESC")
	}*/

	rows, err := q.RunWith(s.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := s.scanRoomMemberItemEntities(rows)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return items,  nil
}

// FetchRoomItem retrieves from storage a roster item entity.
func (s *Storage) FetchRoommemberItem(membername, roomname string) (*roommodel.RoommemberItem, error) {
	q := sq.Select("membername", "roomname","jid","isblack","role").
		From("roommembers").
		Where(sq.And{sq.Eq{"membername": membername},sq.Eq{"roomname": roomname}})


	var rm roommodel.RoommemberItem
	err := s.scanRoomMemberItemEntity(&rm, q.RunWith(s.db).QueryRow())
	switch err {
	case nil:
		return &rm, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}
//删除符合条件(未知类型参数)的房间成员
func (s *Storage) DeleteRoommemberItem(arg interface{}) error{
	err := s.inTransaction(func(tx *sql.Tx) error {
		_, err := sq.Delete("roommembers").
			Where(sq.And{sq.Eq{"membername": arg}, sq.Eq{"roomname": arg}}).     //TODO(lxf) 条件需明确为唯一索引，防止数据操作不准确
			RunWith(tx).Exec()
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) scanRoomMemberItemEntity(ri *roommodel.RoommemberItem, scanner rowScanner) error {
	if err := scanner.Scan(&ri.Membername, &ri.Roomname, &ri.Jid, &ri.Isblack, &ri.Role); err != nil {
		return err
	}
	return nil
}

func (s *Storage) scanRoomMemberItemEntities(scanner rowsScanner) ([]roommodel.RoommemberItem, error) {
	var ret []roommodel.RoommemberItem
	for scanner.Next() {
		var ri roommodel.RoommemberItem
		if err := s.scanRoomMemberItemEntity(&ri, scanner); err != nil {
			return nil, err
		}
		ret = append(ret, ri)
	}
	return ret, nil
}
