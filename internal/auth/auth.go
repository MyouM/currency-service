package auth

import (
	"context"
	"currency-service/internal/config"
	"currency-service/internal/jwt"
	kafkaCur "currency-service/internal/kafka"
	log "currency-service/internal/logger"
	postgres "currency-service/internal/repository/postgres/auth"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var (
	errForUser = errors.New("Auth service have some trouble, try later.")
)

type AuthRequest struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	Error string `json:"error"`
}

func StartAuthService(sigCtx context.Context, cfg *config.KafkaConfig) {
	var (
		wg     sync.WaitGroup
		logger = log.GetLogger()
	)
	logger.Info("Auth service work done")
	registerReqReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.BrokerHost},
		Topic:    kafkaCur.RegisterRespTopic,
		GroupID:  "auth-service-group",
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	registerRespWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{cfg.BrokerHost},
		Topic:    kafkaCur.RegisterReqTopic,
		Balancer: &kafka.LeastBytes{},
	})
	loginReqReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.BrokerHost},
		Topic:    kafkaCur.LoginRespTopic,
		GroupID:  "auth-service-group",
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	loginRespWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{cfg.BrokerHost},
		Topic:    kafkaCur.LoginReqTopic,
		Balancer: &kafka.LeastBytes{},
	})
	defer registerReqReader.Close()
	defer registerRespWriter.Close()
	defer loginReqReader.Close()
	defer loginRespWriter.Close()

	ctx, cancel := context.WithCancel(sigCtx)
	defer cancel()
	wg.Add(2)
	go func() {
		err := registerService(
			ctx,
			registerReqReader,
			registerRespWriter)
		if err != nil {
			cancel()
		}
		wg.Done()
	}()
	go func() {
		err := loginService(
			ctx,
			loginReqReader,
			loginRespWriter)
		if err != nil {
			cancel()
		}
		wg.Done()
	}()
	wg.Wait()
	logger.Info("Auth service work done")
}

func loginService(
	ctx context.Context,
	loginReqReader *kafka.Reader,
	loginRespWriter *kafka.Writer) error {
	logger := log.GetLogger()
	for {
		select {
		case <-ctx.Done():
			logger.Info("Register service work over.")
			return nil
		default:
			var req AuthRequest
			m, err := loginReqReader.ReadMessage(ctx)
			if err != nil {
				logger.Error("Kafka auth request error:", zap.Error(err))
				return err
			}
			if err = json.Unmarshal(m.Value, &req); err != nil {
				logger.Error("Unmarshal error:", zap.Error(err))
				kafkaErr := writeMessage(
					req.ID,
					"",
					errForUser.Error(),
					ctx,
					loginRespWriter)
				if kafkaErr != nil {
					return kafkaErr
				}
				continue
			}
			success, err := postgres.LogIn(req.Login, req.Password)
			if err != nil {
				logger.Error("Postgres error:", zap.Error(err))
				kafkaErr := writeMessage(
					req.ID,
					"",
					errForUser.Error(),
					ctx,
					loginRespWriter)
				if kafkaErr != nil {
					return kafkaErr
				}
				continue
			}
			if !success {
				err = fmt.Errorf("User %s is not exist", req.Login)
				kafkaErr := writeMessage(
					req.ID,
					"",
					err.Error(),
					ctx,
					loginRespWriter)
				if kafkaErr != nil {
					return kafkaErr
				}
				continue
			}
			token, err := jwt.MakeJWT(req.Login, req.Password)
			if err != nil {
				logger.Error("JWT error:", zap.Error(err))
				kafkaErr := writeMessage(
					req.ID,
					"",
					errForUser.Error(),
					ctx,
					loginRespWriter)
				if kafkaErr != nil {
					return kafkaErr
				}
				continue
			}
			kafkaErr := writeMessage(
				req.ID,
				token,
				"",
				ctx,
				loginRespWriter)
			if kafkaErr != nil {
				return kafkaErr
			}
			logger.Info("User logged in:", zap.String("User", req.Login))
		}
	}
	return nil
}

func registerService(
	ctx context.Context,
	registerReqReader *kafka.Reader,
	registerRespWriter *kafka.Writer) error {
	logger := log.GetLogger()
	for {
		select {
		case <-ctx.Done():
			logger.Info("Register service work over.")
			return nil
		default:
			var req AuthRequest
			m, err := registerReqReader.ReadMessage(ctx)
			if err != nil {
				logger.Error("Kafka auth request error:", zap.Error(err))
				continue
			}
			if err = json.Unmarshal(m.Value, &req); err != nil {
				logger.Error("Unmarshal error:", zap.Error(err))
				kafkaErr := writeMessage(
					req.ID,
					"",
					errForUser.Error(),
					ctx,
					registerRespWriter)
				if kafkaErr != nil {
					return kafkaErr
				}
				continue
			}
			exist, err := postgres.IsLoginExist(req.Login)
			if err != nil {
				logger.Error("Postgres error:", zap.Error(err))
				kafkaErr := writeMessage(
					req.ID,
					"",
					errForUser.Error(),
					ctx,
					registerRespWriter)
				if kafkaErr != nil {
					return kafkaErr
				}
				continue
			}
			if exist {
				err = fmt.Errorf("Login already in use")
				kafkaErr := writeMessage(
					req.ID,
					"",
					err.Error(),
					ctx,
					registerRespWriter)
				if kafkaErr != nil {
					return kafkaErr
				}
				continue
			}
			if err = postgres.AddUser(req.Login, req.Password); err != nil {
				logger.Error("Postgres error:", zap.Error(err))
				kafkaErr := writeMessage(
					req.ID,
					"",
					errForUser.Error(),
					ctx,
					registerRespWriter)
				if kafkaErr != nil {
					return kafkaErr
				}
				continue
			}
			token, err := jwt.MakeJWT(req.Login, req.Password)
			if err != nil {
				logger.Error("JWT error:", zap.Error(err))
				kafkaErr := writeMessage(
					req.ID,
					"",
					errForUser.Error(),
					ctx,
					registerRespWriter)
				if kafkaErr != nil {
					return kafkaErr
				}
				continue
			}
			kafkaErr := writeMessage(
				req.ID,
				token,
				"",
				ctx,
				registerRespWriter)
			if kafkaErr != nil {
				return kafkaErr
			}
			logger.Info("Create new user:", zap.String("User", req.Login))
		}
	}
	return nil
}

func writeMessage(
	id, token string,
	errorMsg string,
	ctx context.Context,
	writer *kafka.Writer) error {
	logger := log.GetLogger()

	message := fmt.Sprintf(`{"token":"%s","error":"%s"}`, token, errorMsg)
	err := writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte(id),
			Value: []byte(message),
		})
	if err != nil {
		logger.Error("Kafka respone writer error:", zap.Error(err))
		return err
	}
	return nil
}
