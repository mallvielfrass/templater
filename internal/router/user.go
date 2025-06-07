package router

import (
	"encoding/json"
	"net/http"
)

func (root *Router) CreateTempUser(w http.ResponseWriter, req *http.Request) {
	user := root.userStorage.CreateTempUser()
	jwt := generateJWT(root.jwtSecret, user)
	// Создание JSON-ответа
	response := map[string]string{
		"user": user,
		"jwt":  jwt,
	}

	// Установка заголовка Content-Type
	w.Header().Set("Content-Type", "application/json")
	// Установка статуса OK
	w.WriteHeader(http.StatusCreated)
	// Сериализация ответа в JSON и запись в ResponseWriter
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		return
	}
	return
}
