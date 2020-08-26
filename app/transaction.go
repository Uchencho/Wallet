package app

import (
	"net/http"
)

func FundAccount(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authorized, username, err := checkAuth(req)
	if !authorized {
		unAuthorizedResponse(w, err)
		return
	}

	switch req.Method {
	case http.MethodPost:

	}
}
