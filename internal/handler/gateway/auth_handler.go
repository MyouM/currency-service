package handler

import (
	"currency-service/internal/kafka"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (hr *HandlerRelations) LoginHandler(
	producer kafka.ProducerFuncs,
	consumer kafka.ConsumerFuncs) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var (
			user     kafka.AuthRequest
			authResp kafka.AuthResponse
		)
		defer consumer.Close()
		defer producer.Close()

		dec := json.NewDecoder(req.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&user); err != nil {
			hr.Log.Error("Decode error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		value := fmt.Sprintf(
			`{"type":"%s","login":"%s","password":"%s"}`,
			"login",
			user.Login,
			user.Password)
		err := producer.Send(req.Context(), []byte(user.Login), []byte(value))
		if err != nil {
			hr.Log.Error("Kafka error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		msg, err := consumer.Listen(req.Context())
		if err != nil {
			hr.Log.Error("Kafka error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(msg.Value, &authResp); err != nil {
			hr.Log.Error("Unmarshal error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		if authResp.Error != "" {
			hr.Log.Warn("Auth error:", zap.String("", authResp.Error))
			http.Error(
				w,
				fmt.Sprintf("Error: %s", authResp.Error),
				http.StatusBadRequest)
			return
		}

		if err = hr.Redis.SetToken(authResp.Token); err != nil {
			hr.Log.Error("Redis error", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusInternalServerError)
			return
		}
		hr.Log.Info("User logged in", zap.String("Login", user.Login))
		fmt.Fprintf(w, "Token: %s", authResp.Token)
	}
}

func (hr *HandlerRelations) RegistrationHandler(
	producer kafka.ProducerFuncs,
	consumer kafka.ConsumerFuncs) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var (
			user     kafka.AuthRequest
			authResp kafka.AuthResponse
		)
		defer consumer.Close()
		defer producer.Close()

		dec := json.NewDecoder(req.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&user); err != nil {
			hr.Log.Error("Decode error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}
		hr.Log.Info("Registration operation started")

		value := fmt.Sprintf(
			`{"type":"%s","login":"%s","password":"%s"}`,
			"register",
			user.Login,
			user.Password)
		err := producer.Send(req.Context(), []byte(user.Login), []byte(value))
		if err != nil {
			hr.Log.Error("Kafka error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		msg, err := consumer.Listen(req.Context())
		if err != nil {
			hr.Log.Error("Kafka error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(msg.Value, &authResp); err != nil {
			hr.Log.Error("Unmarshal error:", zap.Error(err))
			http.Error(
				w,
				"Somethings gone wrong",
				http.StatusBadRequest)
			return
		}

		if authResp.Error != "" {
			hr.Log.Warn("Auth error:", zap.String("", authResp.Error))
			http.Error(
				w,
				fmt.Sprintf("Error: %s", authResp.Error),
				http.StatusBadRequest)
			return
		}

		if err = hr.Redis.SetToken(authResp.Token); err != nil {
			hr.Log.Error("Redis error", zap.Error(err))
			http.Error(
				w,
				fmt.Sprint("Registration done, but token is not set"),
				http.StatusInternalServerError)
			return
		}
		hr.Log.Info("New registration", zap.String("Login", user.Login))
		fmt.Fprintf(w, "Token: %s", authResp.Token)
	}
}
