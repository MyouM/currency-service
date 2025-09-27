package auth

import (
	"context"
	"currency-service/internal/config"
	"currency-service/internal/jwt"
	"currency-service/internal/kafka"
	log "currency-service/internal/logger"
	"currency-service/internal/repository/postgres"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

//go:generate mockgen -source=auth.go -destination=mock/mock_auth.go -package=mock_auth

var (
	errorForUser = errors.New("Auth service have some trouble, try later.")
)

type AuthFuncs interface {
	RegisterService(context.Context, []byte) error
	LoginService(context.Context, []byte) error
	LoginCheck() ([]byte, error)
	RegisterCheck() ([]byte, error)
}

type AuthTools struct {
	Repo     postgres.AuthPsqlFuncs
	Req      kafka.AuthRequest
	Producer kafka.ProducerFuncs
	Consumer kafka.ConsumerFuncs
}

func StartAuthService(
	sigCtx context.Context,
	cfg *config.KafkaConfig,
	repo postgres.AuthPsqlFuncs) {
	var (
		req    kafka.AuthRequest
		wg     sync.WaitGroup
		logger = log.GetLogger()

		consumer = kafka.NewConsumer(
			cfg.BrokerHost,
			kafka.GatewayAuthTopic,
			kafka.GroupID)
		producer = kafka.NewProducer(
			cfg.BrokerHost,
			kafka.AuthGatewayTopic)
	)
	defer consumer.Close()
	defer producer.Close()

	ctx, cancel := context.WithCancel(sigCtx)
	defer cancel()

	for {
		m, err := consumer.Listen(ctx)
		if err != nil {
			logger.Error("Kafka auth request error:", zap.Error(err))
			cancel()
			break
		}
		if err = json.Unmarshal(m.Value, &req); err != nil {
			logger.Error("Unmarshal error:", zap.Error(err))
			err := producer.Send(ctx, m.Key, makeMsg(req.Type, "", err.Error()))
			if err != nil {
				cancel()
				logger.Error("Kafka error", zap.Error(err))
				break
			}
			continue
		}

		tools := &AuthTools{
			Repo:     repo,
			Req:      req,
			Producer: producer}
		wg.Add(1)
		switch req.Type {
		case "register":
			go func() {
				err := tools.RegisterService(ctx, m.Key)
				if err != nil {
					logger.Error("Kafka error", zap.Error(err))
					cancel()
				}
				wg.Done()
			}()
		case "login":
			go func() {
				err := tools.LoginService(ctx, m.Key)
				if err != nil {
					logger.Error("Kafka error", zap.Error(err))
					cancel()
				}
				wg.Done()
			}()
		}

	}
	wg.Wait()
	logger.Info("Auth service work done")
}

func (t *AuthTools) RegisterService(ctx context.Context, key []byte) error {
	msg, err := t.RegisterCheck()
	if err != nil {
		return t.Producer.Send(ctx, key, makeMsg(t.Req.Type, "", err.Error()))
	}
	return t.Producer.Send(ctx, key, msg)
}

func (t *AuthTools) LoginService(ctx context.Context, key []byte) error {
	msg, err := t.LoginCheck()
	if err != nil {
		return t.Producer.Send(ctx, key, makeMsg(t.Req.Type, "", err.Error()))
	}
	return t.Producer.Send(ctx, key, msg)
}

func (t *AuthTools) LoginCheck() ([]byte, error) {
	success, err := t.Repo.LogIn(t.Req.Login, t.Req.Password)
	if err != nil {
		return []byte{}, errorForUser
	}
	if !success {
		err = fmt.Errorf("User %s is not exist", t.Req.Login)
		return []byte{}, err
	}

	token, err := jwt.MakeJWT(t.Req.Login, t.Req.Password)
	if err != nil {
		return []byte{}, errorForUser
	}
	return makeMsg(t.Req.Type, token, ""), nil
}

func (t *AuthTools) RegisterCheck() ([]byte, error) {
	exist, err := t.Repo.IsLoginExist(t.Req.Login)
	if err != nil {
		return []byte{}, errorForUser
	}
	if exist {
		err = fmt.Errorf("Login already in use")
		return []byte{}, err
	}
	if err = t.Repo.AddUser(t.Req.Login, t.Req.Password); err != nil {
		return []byte{}, errorForUser
	}
	token, err := jwt.MakeJWT(t.Req.Login, t.Req.Password)
	if err != nil {
		return []byte{}, errorForUser
	}
	return makeMsg(t.Req.Type, token, ""), nil
}

func makeMsg(tp, tkn, err string) []byte {
	msg := fmt.Sprintf(`{"type":"%s","token":"%s","error":"%s"}`, tp, tkn, err)
	return []byte(msg)
}
