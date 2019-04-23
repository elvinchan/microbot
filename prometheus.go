package microbot

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Options
type Option struct {
	Path string
}

func MetricsController() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Routing
		t := r.FormValue("t")
		switch t {
		case "pprof":
			ProfController().ServeHTTP(w, r)
		case "db":
			// TODO
		default:
			promhttp.Handler().ServeHTTP(w, r)
		}
	})
}

func ProfController() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		name := r.FormValue("name")
		if name == "profile" {
			sec, err := strconv.ParseInt(r.FormValue("seconds"), 10, 64)
			if sec <= 0 || err != nil {
				sec = 30
			}

			if durationExceedsWriteTimeout(r, float64(sec)) {
				serveError(w, http.StatusBadRequest, "profile duration exceeds server's WriteTimeout")
				return
			}

			// Set Content Type assuming StartCPUProfile will work,
			// because if it does it starts writing.
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", `attachment; filename="profile"`)
			if err := pprof.StartCPUProfile(w); err != nil {
				// StartCPUProfile failed, so no writes yet.
				serveError(w, http.StatusInternalServerError,
					fmt.Sprintf("Could not enable CPU profiling: %s", err))
				return
			}
			sleep(w, time.Duration(sec)*time.Second)
			pprof.StopCPUProfile()
		} else {
			p := pprof.Lookup(name)
			if p == nil {
				serveError(w, http.StatusNotFound, "Unknown profile")
				return
			}
			gc, _ := strconv.Atoi(r.FormValue("gc"))
			if name == "heap" && gc > 0 {
				runtime.GC()
			}
			debug, _ := strconv.Atoi(r.FormValue("debug"))
			if debug != 0 {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			} else {
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
			}
			p.WriteTo(w, debug)
		}
	})
}

func durationExceedsWriteTimeout(r *http.Request, seconds float64) bool {
	srv, ok := r.Context().Value(http.ServerContextKey).(*http.Server)
	return ok && srv.WriteTimeout != 0 && seconds >= srv.WriteTimeout.Seconds()
}

func sleep(w http.ResponseWriter, d time.Duration) {
	var clientGone <-chan bool
	if cn, ok := w.(http.CloseNotifier); ok {
		clientGone = cn.CloseNotify()
	}
	select {
	case <-time.After(d):
	case <-clientGone:
	}
}

func serveError(w http.ResponseWriter, status int, txt string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Go-Pprof", "1")
	w.Header().Del("Content-Disposition")
	w.WriteHeader(status)
	fmt.Fprintln(w, txt)
}
