package route

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/utils"
	"github.com/gofrs/uuid"
)

// getCart handles GET /cart/{cartID} returning the current cart state.
// It validates the cartID format, maps domain errors to HTTP status codes,
// and mirrors the JSON returned by submit/post cart handlers.
func getCart(srv *utils.AppServer, cartRepo repo.ICartRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cartID := srv.Vars(r)["cartID"]

		// Validate UUID format early as 422 Unprocessable Entity per guidelines
		if _, err := uuid.FromString(cartID); err != nil {
			srv.ResponseErrorEntityUnproc(w, fmt.Errorf("invalid cartID format: %w", err))
			return
		}

		// FindCartByID does not take a tx; it's a KV read outside tx semantics.
		found, err := cartRepo.FindCartByID(cartID)
		if err != nil {
			srv.ResponseErrorNotfound(w, err)
		}

		switch {
		case errors.Is(err, repo.ErrCartNotFound):
			srv.ResponseErrorNotfound(w, err)
			return
		case err != nil:
			srv.ResponseErrorServerErr(w, fmt.Errorf("error fetching cart: %w", err))
			return
		}

		srv.RespondJSON(w, http.StatusOK, found)
	}
}
