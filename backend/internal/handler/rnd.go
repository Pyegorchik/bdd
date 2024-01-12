package handler

import (
	"net/http"

	rnd "github.com/Pyegorchik/bdd/backend/pkg/random"
)



func (h *handler) Rnd(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	randomStr := rnd.RandomString()
	if err := writeResponse(w, r, http.StatusOK, randomStr); err != nil {
		h.makeErrorResponse(w, r, err, code500)
		return
	}

}