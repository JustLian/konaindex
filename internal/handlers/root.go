package handlers

import "net/http"

func GetRootHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"message": "KonaIndex API"})
}
