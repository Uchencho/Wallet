package app

import (
	"encoding/json"
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
		result, err := hitPaystack("alozyuche@gmail.com", "200000")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"Message" : "Something went wrong"}`)
		}
		jsonresp, _ := json.Marshal(result)
		fmt.Fprint(w, string(jsonresp))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"Message" : "Method not allowed"}`)
	}
}
