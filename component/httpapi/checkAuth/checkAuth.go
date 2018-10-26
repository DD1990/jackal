/*
 *请求认证
 */
package checkAuth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

//Basic Auth
func CheckBasicAuth(w http.ResponseWriter, r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}

	fmt.Println(pair)
	return pair[0] == "user" && pair[1] == "pass"
}

//Digest Auth
func CheckDigestAuth(w http.ResponseWriter, r *http.Request) bool { return false }

//AuthOne
func CheckAuthOne(w http.ResponseWriter, r *http.Request) bool { return false }

//AuthTwo
func CheckAuthTwo(w http.ResponseWriter, r *http.Request) bool { return false }
