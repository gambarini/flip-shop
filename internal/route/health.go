package route

import (
	"net/http"
	"time"

	"github.com/gambarini/flip-shop/utils"
)

// health returns basic health information including uptime and version.
func health(srv *utils.AppServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uptime := time.Since(srv.StartTime())
		resp := map[string]string{
			"status":  "ok",
			"uptime":  uptime.String(),
			"version": srv.Version,
		}
		srv.RespondJSON(w, http.StatusOK, resp)
	}
}
