package common

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

type Resp struct {
	Result  interface{} `json:"result"`
	Success bool        `json:"success"`
	Error   RespError   `json:"error"`
}

type RespError struct {
	Message string `json:"message"`
}

func renderJson(w http.ResponseWriter, status int, v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(bs)
}

func RenderDataJson(w http.ResponseWriter, status int, v interface{}) {
	renderJson(w, status, Resp{Result: v, Success: true})
}

func RenderErrorJson(w http.ResponseWriter, status int, msg string) {
	renderJson(w, status, Resp{Error: RespError{
		Message: msg,
	}})
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