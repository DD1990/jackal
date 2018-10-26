/*
 * Copyright (c) 2018 Miguel Ángel Ortuño.
 * See the LICENSE file for more information.
 */

package sql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/go-sql-driver/mysql" // SQL driver
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/pool"
	// 引入数据库持久化框架gorm（建表用）
	"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/mysql"      //重复引用，上面已经有mysql，实际上他们是同等的
)

var (
	nowExpr = sq.Expr("NOW()")
)

type rowScanner interface {
	Scan(...interface{}) error
}

type rowsScanner interface {
	rowScanner
	Next() bool
}

// Config represents SQL storage configuration.
type Config struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	PoolSize int    `yaml:"pool_size"`
}

// Storage represents a SQL storage sub system.
type Storage struct {
	db     *sql.DB
	gormdb *gorm.DB //建表使用
	pool   *pool.BufferPool
	doneCh chan chan bool
}

// New returns a SQL storage instance.
func New(cfg *Config) *Storage {
	var err error
	s := &Storage{
		pool:   pool.NewBufferPool(),
		doneCh: make(chan chan bool),
	}
	host := cfg.Host
	user := cfg.User
	pass := cfg.Password
	db := cfg.Database
	poolSize := cfg.PoolSize

	//初始化数据库原生db
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, pass, host, db)
	s.db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("%v", err)
	}
	s.db.SetMaxOpenConns(poolSize) // set max opened connection count

	if err := s.db.Ping(); err != nil {
		log.Fatalf("%v", err)
	}

	s.gormdb, err = gorm.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("%v", err)
		panic(err)
	}

	//初始化数据库表 initBaseTable()
	err = s.initBaseTable()
	if err != nil {
		log.Fatalf("%v", err)
	}
	go s.loop()

	return s
}

// NewMock returns a mocked SQL storage instance.
func NewMock() (*Storage, sqlmock.Sqlmock) {
	var err error
	var sqlMock sqlmock.Sqlmock
	s := &Storage{
		pool: pool.NewBufferPool(),
	}
	s.db, sqlMock, err = sqlmock.New()
	if err != nil {
		log.Fatalf("%v", err)
	}
	return s, sqlMock
}

// Shutdown shuts down SQL storage sub system.
func (s *Storage) Shutdown() {
	ch := make(chan bool)
	s.doneCh <- ch
	<-ch
}

func (s *Storage) loop() {
	tc := time.NewTicker(time.Second * 15)
	defer tc.Stop()
	for {
		select {
		case <-tc.C:
			err := s.db.Ping()
			if err != nil {
				log.Error(err)
			}
		case ch := <-s.doneCh:
			s.db.Close()
			close(ch)
			return
		}
	}
}

//事物控制管理
func (s *Storage) inTransaction(f func(tx *sql.Tx) error) error {
	tx, txErr := s.db.Begin()
	if txErr != nil {
		return txErr
	}
	if err := f(tx); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

/******************************************TODO(lxf) 单独放在createbasetable.go包中********************************************************/
//房间配置缺省值
const (
	defaultRoomeName  = "default"
	defaultCfgcontent = `<x xmlns="jabber:x:data" type="form"><title>Configuration for "darkcave" Room</title><instructions>Your room darkcave@macbeth has been created!The default configuration is as follows:- No logging- No moderation- Up to 20 occupants- No password required- No invitation required- Room is not persistent- Only admins may change the subject- Presence broadcasted for all usersTo accept the default configuration, click OK. Toselect a different configuration, please complete this form.</instructions><field label="Natural-Language Room Name" type="text-single" var="muc#roomconfig_roomname"><value/></field><field label="Short Description of Room" type="text-single" var="muc#roomconfig_roomdesc"><value/></field><field label="Make Room Persistent?" type="boolean" var="muc#roomconfig_persistentroom"><value>true</value></field><field label="Make Room Publicly Searchable?" type="boolean" var="muc#roomconfig_publicroom"><value>1</value></field><field label="Make Room Moderated?" type="boolean" var="muc#roomconfig_moderatedroom"><value>0</value></field><field label="Make Room Members Only?" type="boolean" var="muc#roomconfig_membersonly"><value>0</value></field><field label="Allow Occupants to Invite Others?" type="boolean" var="muc#roomconfig_allowinvites"><value>1</value></field><field label="Password Required to Enter?" type="boolean" var="muc#roomconfig_passwordprotectedroom"><value>0</value></field><field label="Password" type="text-single" var="muc#roomconfig_roomsecret"><value/></field><field type="list-single" var="muc#roomconfig_anonymity" label="Room anonymity level:"><value>semianonymous</value><option label="Non-Anonymous Room"><value>nonanonymous</value></option><option label="Semi-Anonymous Room"><value>semianonymous</value></option><option label="Fully-Anonymous Room"><value>fullanonymous</value></option></field><field label="Allow Occupants to Change Subject?" type="boolean" var="muc#roomconfig_changesubject"><value>0</value></field><field label="Enable Public Logging?" type="boolean" var="muc#roomconfig_enablelogging"><value>true</value></field><field type="list-single" var="logging_format" label="Logging format:"><value>html</value><option label="HTML"><value>html</value></option><option label="Plain text"><value>plain</value></option></field><field label="Maximum Number of History Messages Returned by Room" type="text-single" var="muc#maxhistoryfetch"><value>50</value></field><field label="Maximum Number of Occupants" type="list-single" var="muc#roomconfig_maxusers"><value>20</value><option label="10"><value>10</value></option><option label="20"><value>20</value></option><option label="30"><value>30</value></option><option label="50"><value>50</value></option><option label="100"><value>100</value></option><option label="None"><value/></option></field><field type="list-single" var="tigase#presence_delivery_logic" label="Presence delivery logic"><value>PREFERE_PRIORITY</value><option label="PREFERE_LAST"><value>PREFERE_LAST</value></option><option label="PREFERE_PRIORITY"><value>PREFERE_PRIORITY</value></option></field><field type="boolean" var="tigase#presence_filtering" label="Enable filtering of presence (broadcasting presence only between selected groups"><value>0</value></field><field type="list-multi" var="tigase#presence_filtered_affiliations" label="Affiliations for which presence should be delivered"><option label="admin"><value>admin</value></option><option label="member"><value>member</value></option><option label="none"><value>none</value></option><option label="outcast"><value>outcast</value></option><option label="owner"><value>owner</value></option></field><field type="boolean" var="tigase#welcome_messages" label="Send welcome messages on room creation"><value>1</value></field><field type="jid-multi" var="muc#roomconfig_roomadmins" label="Full List of Room Admins"/></x>`
)

//初始化muc组件的基本表
func (s *Storage) initBaseTable() error {

	//房间配置表结构(字段首字母必须大写，不然报错1064)
	type roomcfg_test struct {
		Roomname   string `gorm:"type:varchar(255);not null;primary_key"`
		Cfgcontent string `gorm:"type:mediumtext;not null;"`
	}
	//房间表结构
	type rooms_test struct {
		Roomname   string `gorm:"type:varchar(255);not null;primary_key"`
		Roomserver string `gorm:"type:varchar(255);not null;"`
		Password   string `gorm:"type:varchar(255);NULL DEFAULT NULL;"`
		Ispravite  bool   `gorm:"type:tinyint(1);not null;"`
		Roomtype   uint8  `gorm:"type:int(1);not null;"`
		Update_at  time.Time
		Create_at  time.Time
	}
	//房间成员表结构
	type roommembers_test struct {
		Membername string `gorm:"type:varchar(255);not null;unique_index:uniquename_mr"`
		Roomname   string `gorm:"type:varchar(255);not null;unique_index:uniquename_mr"`
		Jid        string `gorm:"type:varchar(255);not null;"`
		Isblack    bool   `gorm:"type:tinyint(1);not null;"`
		Role       uint8  `gorm:"type:int(1);not null;"`
		Update_at  time.Time
		Create_at  time.Time
	}

	err := s.inTransaction(func(tx *sql.Tx) error {
		defer s.gormdb.Close()
		//建表（房间表，房间成员表，房间配置表）
		//全局设置表名不可以为复数形式(表明后多一个字符s)。
		// 如果想自己根据不同的条件来创建对应的表名，可以重写模型的TableName方法,如:https://blog.csdn.net/wongcony/article/details/79063407
		s.gormdb.SingularTable(true)
		if !s.gormdb.HasTable(&rooms_test{}) && !s.gormdb.HasTable(&roommembers_test{}) && !s.gormdb.HasTable(&roomcfg_test{}) {
			if err := s.gormdb.Set("gorm:options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(&rooms_test{}, &roommembers_test{}, &roomcfg_test{}).Error; err != nil {
				panic(err)
				return err
			} else {
				//初始化房间缺省配置数据
				roomcfgparams := &roomcfg_test{
					Roomname:   defaultRoomeName,
					Cfgcontent: defaultCfgcontent,
				}
				if err := s.gormdb.Create(roomcfgparams).Error; err != nil {
					return err
				}
			}
		}

		return nil

		/*		原生db建表方式
				//房间配置表
				roomcfgsql := "create table roomcfg_test (roomname varchar(255) primary key ,cfgcontent mediumtext not null);"
				smt, err = s.db.Prepare(roomcfgsql)
				smt.Exec()
				return err*/
	})

	if err != nil {
		return err
	}
	return nil
}
