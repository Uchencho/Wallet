package app

import (
	"fmt"
	"net/http"
)

func FundAccount(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authorized, _, err := checkAuth(req)
	if !authorized {
		unAuthorizedResponse(w, err)
		return
	}

	switch req.Method {
	case http.MethodPost:
		if hitPaystack("alozyuche@gmail.com", "2000000") {
			fmt.Fprint(w, `{"Message" : "Hit successful"}`)

		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"Message" : "Method not allowed"}`)
	}
}
