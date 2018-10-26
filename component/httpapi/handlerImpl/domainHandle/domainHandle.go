package domainHandle

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"github.com/ortuman/jackal/component/commponentInter"
	"github.com/ortuman/jackal/component/httpapi/checkAuth"
	aer "github.com/ortuman/jackal/component/httpapi/errors"
	"github.com/ortuman/jackal/component/httpapi/handlerImpl/model"
	"github.com/ortuman/jackal/component/httpapi/storage"
	"github.com/ortuman/jackal/host"
	"github.com/ortuman/jackal/log"
	"io/ioutil"
	"net/http"
	"strings"
)

type DoaminHandlerface struct {
	Command
	aer.ApiErr
}
type Command struct {
	Jid    string `xml:"jid"`
	Node   string `xml:"node"`
	Fields Field  `xml:"fields"`
}
type Field struct {
	Items []Item `xml:"item"`
}
type Item struct {
	Var     string   `xml:"var"`
	Type    string   `xml:"type"`
	Value   string   `xml:"value"`
	Label   string   `xml:"label"`
	Options []Option `xml:"options"`
}
type Option struct {
	Items []Item `xml:"item"`
}

func (domain *DoaminHandlerface) DomainManage(w http.ResponseWriter, r *http.Request) {

	log.Infof("进入 DomainManage()")
	//认证失败，返回
	if !checkAuth.CheckBasicAuth(w, r) {
		domain.HandleErr(&w, aer.AUTH_ERR_CODE, nil)
		return
	}

	defer r.Body.Close()
	con, _ := ioutil.ReadAll(r.Body) //获取body体数据
	bodyContent := string(con)

	log.Infof("请求包体：", strings.TrimSpace(bodyContent))
	//解析xml
	/*原生态方式*/
	dom := DoaminHandlerface{}
	xml.Unmarshal(con, &dom)
	domain = &dom

	/*jackal封装xep0004,缺点：无法解析换行的xml*/
	//parser := xmpp.NewParser(strings.NewReader(bodyContent), xmpp.DefaultMode, 0)
	//elem, elemerr := parser.ParseElement()
	//if elemerr != nil {
	//	log.Error(elemerr)
	//	return
	//}
	//domain.Name = elem.Elements().Child("fields").Elements().Children("item")[0].Attributes().Get("var")
	//user = &user1
	var err error
	if &r.RequestURI != nil && r.RequestURI != "" {

		if uris := strings.Split(r.RequestURI, "/"); len(uris) < 2 {
			dom.HandleErr(&w, aer.URI_ERR_CODE, nil)
			return
		} else {
			switch domain.Node {
			case model.ITEMUPDAT:
				err = domain.getDomain()
				if err != nil {
					dom.HandleErr(&w, aer.UPDATE_DOMAIN_ERR_CODE, err)
					return
				} else {
					goto RESPONSE
				}
			case model.ITEMADD:
				err = domain.addDomain()
				if err != nil {
					dom.HandleErr(&w, aer.ADD_DOMAIN_ERR_CODE, err)
					return
				} else {
					goto RESPONSE
				}
			case model.ITEMREMOVE:
				err = domain.deleteDomain()

			}
		}

	}

RESPONSE:
	w.Write([]byte("处理成功\n"))
	w.Write([]byte("处理结果：" + domain.Name()))
}

//新增虚拟域
func (domain *DoaminHandlerface) addDomain() error {
	domainName := domain.Name()
	r := sha1.Sum([]byte(domainName)) //sha1加密

	dom := model.Domain{
		ServiceJid:    domainName,
		ServiceJidSha: hex.EncodeToString(r[:]),
	}
	//copy(sha1([]byte(domainName), &dom.ServiceJidSha)
	//数据库持久化
	err := storage.Instance().InsertOrUpdateDomain(&dom)
	if err != nil {
		log.Error(err)
		return err
	}
	//TODO(lxf)生成并加载新建域的证书
	host.LoadCertificateForVirtualDomain(dom.ServiceJid)

	//TODO(lxf)生成并加载新建域对应的群聊服务
	//domainMuc := commponentInter.DomainMuc{}
	commponentInter.Initialize(dom.ServiceJid)
	return nil
}

//查询虚拟域
func (domain *DoaminHandlerface) getDomain() error {
	return nil
}

//删除虚拟域
func (domain *DoaminHandlerface) deleteDomain() error {
	return nil
}

//获取虚拟域名称
func (domain *DoaminHandlerface) Name() string {
	for _, itm := range domain.Fields.Items {
		if itm.Var == "Domain name" {
			return itm.Value
		}
	}
	return ""
}
