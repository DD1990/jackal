package errors

import (
	"errors"
	"fmt"
	"github.com/ortuman/jackal/log"
	"net/http"
)

//API错误代码及对应解析
const (
	PASSWORD_NULL_ERR, EMAIL_NULL_ERR     = "密码不能为空", "邮箱不能为空"
	ADD_USER_ERR_CODE, ADD_USER_ERR       = 201, "注册用户失败"
	SEARCH_USER_ERR_CODE, SEARCH_USER_ERR = 202, "查询用户失败"

	ADD_DOMAIN_ERR_CODE, ADD_DOMAIN_ERR       = 301, "创建虚拟域失败"
	UPDATE_DOMAIN_ERR_CODE, UPDATE_DOMAIN_ERR = 302, "更新虚拟域失败"

	ADD_NODE_ERR_CODE, ADD_NODE_ERR             = 401, "创建节点失败"
	UPDATE_NODE_ERR_CODE, UPDATE_NODE_ERR       = 402, "更新节点失败"
	DELETE_NODE_ERR_CODE, DALETE_NODE_ERR       = 403, "删除节点失败"
	SUBSCRIBE_NODE_ERR_CODE, SUBSCRIBE_NODE_ERR = 404, "用户订阅失败"
	PUBLISH_ITEM_ERR_CODE, PUBLISH_ITEM_ERR     = 405, "群发消息失败"

	AUTH_ERR_CODE, AUTH_ERR       = 997, "请求认证失败"
	URI_ERR_CODE, URI_ERR         = 998, "请求地址错误"
	UNKNOWN_ERR_CODE, UNKNOWN_ERR = 999, "未知错误"
)

type ApiErr struct {
	w http.ResponseWriter
}

func (ar *ApiErr) getErr(errCode int) string {
	switch errCode {
	case ADD_USER_ERR_CODE:
		return ADD_USER_ERR
	case SEARCH_USER_ERR_CODE:
		return SEARCH_USER_ERR

	case ADD_DOMAIN_ERR_CODE:
		return ADD_DOMAIN_ERR
	case UPDATE_DOMAIN_ERR_CODE:
		return UPDATE_DOMAIN_ERR

	case ADD_NODE_ERR_CODE:
		return ADD_NODE_ERR
	case UPDATE_NODE_ERR_CODE:
		return UPDATE_NODE_ERR
	case DELETE_NODE_ERR_CODE:
		return DALETE_NODE_ERR
	case SUBSCRIBE_NODE_ERR_CODE:
		return SUBSCRIBE_NODE_ERR
	case PUBLISH_ITEM_ERR_CODE:
		return PUBLISH_ITEM_ERR

	case AUTH_ERR_CODE:
		return AUTH_ERR
	case URI_ERR_CODE:
		return URI_ERR
	case UNKNOWN_ERR_CODE:
		return UNKNOWN_ERR
	}
	return UNKNOWN_ERR
}

func (ar *ApiErr) HandleErr(w *http.ResponseWriter, errCode int, err error) {
	ar.w = *w
	log.Error(errors.New(ar.getErr(errCode)))
	//w.Header().Set("WWW-Authenticate", `Basic realm="MY REALM"`)  //请求认证失败
	ar.w.WriteHeader(errCode)
	if err == nil {
		ar.w.Write([]byte(fmt.Sprintf("%s%s", ar.getErr(errCode), "，原因：【未知】\n")))
		return
	}
	ar.w.Write([]byte(fmt.Sprintf("%s%s%s%s", ar.getErr(errCode), "，原因：【", err.Error(), "】\n")))
}
