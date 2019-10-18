package transfer

import (
	"context"
	"time"

	"github.com/cockroachdb/apd"
	"github.com/go-kit/kit/log"
	uuid "github.com/satori/go.uuid"

	wallet "github.com/xsleonard/gokit-example"
)

type loggingService struct {
	logger log.Logger
	wallet.Service
}

// NewLoggingService creates a Service with logging
func NewLoggingService(logger log.Logger, s wallet.Service) wallet.Service {
	return loggingService{
		logger:  logger,
		Service: s,
	}
}

func (s loggingService) Transfer(ctx context.Context, to, from uuid.UUID, amount *apd.Decimal) (p *wallet.Payment, err error) {
	defer func(begin time.Time) {
		logger := s.logger
		if err != nil {
			logger = log.With(logger, "err", err)
		}
		s.logger.Log("operation", "transfer", "to", to, "from", from, "amount", amount, "took", time.Since(begin))
	}(time.Now())

	return s.Service.Transfer(ctx, to, from, amount)
}

func (s loggingService) Payments(ctx context.Context) (p []wallet.Payment, err error) {
	defer func(begin time.Time) {
		logger := s.logger
		if err != nil {
			logger = log.With(logger, "err", err)
		}
		s.logger.Log("operation", "payments", "took", time.Since(begin))
	}(time.Now())

	return s.Service.Payments(ctx)
}

func (s loggingService) Accounts(ctx context.Context) (a []wallet.Account, err error) {
	defer func(begin time.Time) {
		logger := s.logger
		if err != nil {
			logger = log.With(logger, "err", err)
		}
		s.logger.Log("operation", "accounts", "took", time.Since(begin))
	}(time.Now())

	return s.Service.Accounts(ctx)
}
