package mygeecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_mygeecache/"

type HTTPPool struct {
	// 用于记录自己的地址，包括主机名/IP 端口
	self string
	// 作为通讯地址的前缀，默认为_mygeecache
	// 那么http://example.com/_mygeecache/开头的请求，就用于节点之间的访问
	basePath string
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// v是一个可变参数，将接受到的任意数量的参数打包成一个切片；之后又通过v...将切片打散成多个参数
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// 实现ServeHTTP方法，处理请求r,通过w进行回应,包括正常情况回复值的内容，以及非正常情况通过http.Error回复错误信息
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		// 如果请求的url的前缀不对，直接panic
		panic("unexpected path" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// r.URL.Path约定为 <basePath>/<groupName>/<key>
	// 从其中提取出groupName和key
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// application/octet-stream表示相应的数据是未知的二进制数据
	w.Header().Set("Content-type", "application/octet-stream")
	w.Write(view.ByteSlice())

}
