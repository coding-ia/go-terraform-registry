package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func JsonResponse(w http.ResponseWriter, httpStatus int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
	}
}
