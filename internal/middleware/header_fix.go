package middleware

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// bufferWriter 用于缓存响应体并设置 Content-Length
type bufferWriter struct {
	gin.ResponseWriter
	buf    *bytes.Buffer
	status int
}

func (w *bufferWriter) Write(b []byte) (int, error) {
	if w.buf == nil {
		w.buf = &bytes.Buffer{}
	}
	return w.buf.Write(b)
}

func (w *bufferWriter) WriteHeader(code int) {
	w.status = code
}

// flush 将缓存的内容写入真正的 ResponseWriter
func (w *bufferWriter) flush() {
	if w.buf == nil {
		return
	}
	// 在写入前设置 Content-Length
	w.Header().Set("Content-Length", strconv.Itoa(w.buf.Len()))
	// 修正 Content-Type 为小写（可选，大小写不敏感）
	if ct := w.Header().Get("Content-Type"); strings.Contains(ct, "application/json") {
		w.Header().Set("content-type", "application/json")
	}
	// 尝试移除 Date 头（可能无效，但保留）
	w.Header().Del("Date")
	// 写入状态码（如果未设置则默认为 200）
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.ResponseWriter.WriteHeader(w.status)
	// 写入缓存的响应体
	w.ResponseWriter.Write(w.buf.Bytes())
}

// FixResponseHeaders 中间件：确保响应头符合 Kards 客户端要求
func FixResponseHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		bw := &bufferWriter{ResponseWriter: c.Writer}
		c.Writer = bw
		c.Next()
		// 请求结束后，刷新缓存并写入真实响应
		bw.flush()
	}
}
