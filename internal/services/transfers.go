package services

import (
	"context"
	"fmt"
	"strings"
	"transfers-api/internal/config"
	"transfers-api/internal/enums"
	"transfers-api/internal/known_errors"
	"transfers-api/internal/logging"
	"transfers-api/internal/models"
)

//go:generate mockery --name TransfersRepository --structname TransfersRepositoryMock --filename transfers_repository_mock.go --output mocks --outpkg mocks

type TransfersRepository interface {
	Create(ctx context.Context, transfer models.Transfer) (string, error)
	GetByID(ctx context.Context, id string) (models.Transfer, error)
	GetTransfersByUserID(ctx context.Context, userID string) ([]models.Transfer, error)
	Update(ctx context.Context, transfer models.Transfer) error
	Delete(ctx context.Context, id string) error
}

type TransfersPublisher interface {
	Publish(operation string, transferID string) error
}

type TransfersService struct {
	businessCfg        config.BusinessConfig
	transfersRepo      TransfersRepository
	transfersCCache    TransfersRepository
	transfersPublisher TransfersPublisher
}

func NewTransfersService(businessCfg config.BusinessConfig, transfersRepo TransfersRepository, transfersCCache TransfersRepository, transfersPublisher TransfersPublisher) *TransfersService {
	return &TransfersService{
		businessCfg:        businessCfg,
		transfersRepo:      transfersRepo,
		transfersCCache:    transfersCCache,
		transfersPublisher: transfersPublisher,
	}
}

func (s *TransfersService) Create(ctx context.Context, transfer models.Transfer) (string, error) {
	if strings.TrimSpace(transfer.SenderID) == "" {
		return "", fmt.Errorf("sender_id is required: %w", known_errors.ErrBadRequest)
	}
	if strings.TrimSpace(transfer.ReceiverID) == "" {
		return "", fmt.Errorf("receiver_id is required: %w", known_errors.ErrBadRequest)
	}
	if transfer.Currency == enums.CurrencyUnknown {
		return "", fmt.Errorf("invalid currency %s: %w", transfer.Currency.String(), known_errors.ErrBadRequest)
	}
	if transfer.Amount <= 0 {
		return "", fmt.Errorf("amount should be greater than 0: %w", known_errors.ErrBadRequest)
	}
	if strings.TrimSpace(transfer.State) == "" { // TODO: replace with enums.ParseState
		return "", fmt.Errorf("state is required: %w", known_errors.ErrBadRequest)
	}
	id, err := s.transfersRepo.Create(ctx, transfer)
	if err != nil {
		return "", fmt.Errorf("error creating transfer in repository: %w", err)
	}
	logging.Logger.Infof("Transfer created in DB with ID: %s", id)

	// Publish transfer creation event
	go func() {
		if err := s.transfersPublisher.Publish("created", id); err != nil {
			logging.Logger.Warnf("error publishing transfer creation event: %v", err)
		}
	}()

	transfer.ID = id
	if _, err := s.transfersCCache.Create(ctx, transfer); err != nil {
		logging.Logger.Warnf("error creating transfer in ccache: %w", err)
	}
	logging.Logger.Infof("Transfer created in ccache with ID: %s", id)
	return id, nil
}

func (s *TransfersService) GetByID(ctx context.Context, id string) (models.Transfer, error) {
	transfer, err := s.transfersCCache.GetByID(ctx, id)
	if err == nil {
		logging.Logger.Infof("Transfer retrieved from ccache with ID: %s", id)
		return transfer, nil
	}

	transfer, err = s.transfersRepo.GetByID(ctx, id)
	if err != nil {
		return models.Transfer{}, fmt.Errorf("error getting transfer %s from repository: %w", id, err)
	}

	logging.Logger.Infof("Transfer retrieved from DB with ID: %s", id)

	if _, err := s.transfersCCache.Create(ctx, transfer); err != nil {
		logging.Logger.Warnf("error creating transfer in ccache: %w", err)
	}

	logging.Logger.Infof("Transfer created in ccache with ID: %s", id)

	return transfer, nil
}

func (s *TransfersService) Update(ctx context.Context, transfer models.Transfer) error {
	if strings.TrimSpace(transfer.ID) == "" {
		return fmt.Errorf("ID is required: %w", known_errors.ErrBadRequest)
	}
	if strings.TrimSpace(transfer.SenderID) == "" &&
		strings.TrimSpace(transfer.ReceiverID) == "" &&
		transfer.Currency == enums.CurrencyUnknown &&
		transfer.Amount <= 0 &&
		strings.TrimSpace(transfer.State) == "" {
		return fmt.Errorf("error updating transfer %s: no fields to update: %w", transfer.ID, known_errors.ErrBadRequest)
	}
	if err := s.transfersRepo.Update(ctx, transfer); err != nil {
		return fmt.Errorf("error updating transfer %s in repository: %w", transfer.ID, err)
	}

	// Publish transfer update event
	go func() {
		if err := s.transfersPublisher.Publish("updated", transfer.ID); err != nil {
			logging.Logger.Warnf("error publishing transfer update event: %v", err)
		}
	}()

	if err := s.transfersCCache.Update(ctx, transfer); err != nil {
		logging.Logger.Warnf("error updating transfer in ccache: %w", err)
	}

	return nil
}

func (s *TransfersService) Delete(ctx context.Context, id string) error {
	if err := s.transfersRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error deleting transfer %s from repository: %w", id, err)
	}

	// Publish transfer deletion event
	go func() {
		if err := s.transfersPublisher.Publish("deleted", id); err != nil {
			logging.Logger.Warnf("error publishing transfer deletion event: %v", err)
		}
	}()

	if err := s.transfersCCache.Delete(ctx, id); err != nil {
		logging.Logger.Warnf("error deleting transfer from ccache: %w", err)
	}

	return nil
}

func (s *TransfersService) GetTransfersByUserID(ctx context.Context, userID string) ([]models.Transfer, error) {
	transfer, err := s.transfersCCache.GetTransfersByUserID(ctx, userID)
	if err == nil {
		logging.Logger.Infof("Transfer retrieved from ccache with ID: %s", userID)
		return transfer, nil
	}

	transfer, err = s.transfersRepo.GetTransfersByUserID(ctx, userID)
	if err != nil {
		return []models.Transfer{}, fmt.Errorf("error getting transfer for user %s from repository: %w", userID, err)
	}

	logging.Logger.Infof("Transfer retrieved from DB with ID: %s", userID)

	for _, t := range transfer {
		if _, err := s.transfersCCache.Create(ctx, t); err != nil {
			logging.Logger.Warnf("error creating transfer in ccache: %w", err)
		}
	}

	logging.Logger.Infof("Transfer created in ccache with ID: %s", userID)

	return transfer, nil
}
