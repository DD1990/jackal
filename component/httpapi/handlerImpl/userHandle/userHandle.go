package userHandle

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/ortuman/jackal/component/httpapi/checkAuth"
	aer "github.com/ortuman/jackal/component/httpapi/errors"
	md "github.com/ortuman/jackal/component/httpapi/handlerImpl/model"
	"github.com/ortuman/jackal/component/httpapi/storage"
	"github.com/ortuman/jackal/log"
	"github.com/ortuman/jackal/model"
	"github.com/ortuman/jackal/xmpp"
	"github.com/ortuman/jackal/xmpp/jid"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type UserHandlerface struct {
	aer.ApiErr
	Password       string `xml:"password"`
	Email          string `xml:"email"`
	LastPresence   *xmpp.Presence
	LastPresenceAt time.Time
}

//用户管理
func (user *UserHandlerface) UserManage(w http.ResponseWriter, r *http.Request) {

	log.Infof("进入 UserManage()")
	//认证失败，返回
	if !checkAuth.CheckBasicAuth(w, r) {
		user.HandleErr(&w, aer.AUTH_ERR_CODE, nil)
		return
	}

	defer r.Body.Close()
	con, _ := ioutil.ReadAll(r.Body) //获取body体数据
	bodyContent := string(con)

	log.Infof("请求包体：", strings.TrimSpace(bodyContent))
	//解析xml
	//type User struct {
	//	Password string `xml:"password"`
	//	Email    string `xml:"email"`
	//}
	//user1 := User{}
	xml.Unmarshal(con, user)

	//user = &user1
	var err error
	if &r.RequestURI != nil && r.RequestURI != "" {

		if uris := strings.Split(r.RequestURI, "/"); len(uris) < 2 {
			user.HandleErr(&w, aer.URI_ERR_CODE, nil)
		} else {
			switch r.Method {
			case md.GET:
				err = user.getUser(uris[len(uris)-1])
				if err != nil {
					user.HandleErr(&w, aer.SEARCH_USER_ERR_CODE, err)
					return
				} else {
					goto RESPONSE
				}
			case md.PUT:
				//user.StringsToUser(uris)
				err = user.addUser()
				if err != nil {
					user.HandleErr(&w, aer.ADD_USER_ERR_CODE, err)
					return
				} else {
					goto RESPONSE
				}
			case md.DELETE:
				err = user.deleteUser(uris[len(uris)-1])

			}
		}

	}

RESPONSE:
	w.Write([]byte("处理成功\n"))
	w.Write([]byte("处理结果：" + user.String()))
}

//查询单个用户
func (user *UserHandlerface) getUser(email string) error {
	return nil
}

//新增用户
func (user *UserHandlerface) addUser() error {
	if user.Password == "" {
		return errors.New(aer.PASSWORD_NULL_ERR)
	}
	if user.Email == "" {
		return errors.New(aer.EMAIL_NULL_ERR)
	}

	type UserRes struct {
		Jid    string
		Domain string
		Uid    uint64
	}
	userRes := UserRes{
		Jid:    user.Email,
		Domain: strings.Split(user.Email, "@")[1],
		//Uid:    111, //数据库插入时自动生成
	}
	log.Infof(userRes.Jid, userRes.Uid, userRes.Domain)
	jid, _ := jid.NewWithString("localhost", true)
	user2 := model.User{
		Username:     strings.Split(user.Email, "@")[0],
		Password:     user.Password,
		Domain:       strings.Split(user.Email, "@")[1],
		LastPresence: xmpp.NewPresence(jid, jid, xmpp.UnavailableType),
	}
	//新增用户到数据库
	if err := storage.Instance().InsertOrUpdateUser(&user2); err != nil {
		log.Error(err)
	}

	return nil
}

//删除用户
func (user *UserHandlerface) deleteUser(email string) error {
	return nil
}

func (user *UserHandlerface) String() string {
	userStr := fmt.Sprintf("%s%s", "用户名称", user.Email)
	return userStr
}

//func (user *UserHandlerface) StringsToUser(userStrs []string) {
//	user.Email = userStrs[0]
//	user.Password = userStrs[1]
//}
