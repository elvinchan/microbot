package utils

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

func RenderJson(w http.ResponseWriter, v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(bs)
}

type Resp struct {
	Result  interface{} `json:"result"`
	Success bool        `json:"success"`
	Error   RespError   `json:"error"`
}

type RespError struct {
	Message string `json:"message"`
}

func RenderDataJson(w http.ResponseWriter, data interface{}) {
	RenderJson(w, Resp{Result: data, Success: true})
}

func RenderErrorJson(w http.ResponseWriter, err error) {
	RenderJson(w, Resp{Success: false, Error: RespError{
		Message: err.Error(),
	}})
}

func Render(w http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		RenderErrorJson(w, err)
		return
	}
	RenderDataJson(w, data)
}

func IsPublicIP(ip string) bool {
	netIP := net.ParseIP(ip)
	if netIP.IsLoopback() || netIP.IsLinkLocalMulticast() || netIP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := netIP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

func RealIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ", ")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ra
}
