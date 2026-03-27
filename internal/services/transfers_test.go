package services

import (
	"context"
	"testing"
	"time"
	"transfers-api/internal/config"
	"transfers-api/internal/enums"
	"transfers-api/internal/models"
	"transfers-api/internal/services/mocks"

	"github.com/stretchr/testify/assert"
)

func TestTransfersService_GetByID(t *testing.T) {
	var (
		ctx                = context.Background()
		cfg                = config.BusinessConfig{TransferMinAmount: 10}
		transfersRepo      = mocks.NewTransfersRepositoryMock(t)
		transfersCCache    = mocks.NewTransfersRepositoryMock(t)
		transfersPublisher = mocks.NewTransfersPublisherMock(t)
	)

	for _, testCase := range []struct {
		name             string
		transferID       string
		expectedTransfer models.Transfer
	}{
		{
			name:       "Transfer successfully retrieved with USD",
			transferID: "TRF-20250326-001",
			expectedTransfer: models.Transfer{
				ID:         "TRF-20250326-001",
				SenderID:   "USER-JOHN-123",
				ReceiverID: "USER-MARY-789",
				Amount:     500.50,
				Currency:   enums.CurrencyUSD,
			},
		},
		{
			name:       "Transfer successfully retrieved with EUR",
			transferID: "TRF-20250326-002",
			expectedTransfer: models.Transfer{
				ID:         "TRF-20250326-002",
				SenderID:   "USER-ALICE-456",
				ReceiverID: "USER-CARLOS-321",
				Amount:     1250.75,
				Currency:   enums.CurrencyEUR,
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			transfersCCache.On("GetByID", ctx, testCase.transferID).Return(testCase.expectedTransfer, nil)
			svc := NewTransfersService(cfg, transfersRepo, transfersCCache, transfersPublisher)
			// when
			transfer, err := svc.GetByID(ctx, testCase.transferID)
			// then
			assert.Nil(t, err)
			assert.Equal(t, testCase.transferID, transfer.ID)
		})
	}
}

func TestTransfersService_Delete(t *testing.T) {
	// given
	var (
		ctx                = context.Background()
		cfg                = config.BusinessConfig{TransferMinAmount: 10}
		transfersRepo      = mocks.NewTransfersRepositoryMock(t)
		transfersCCache    = mocks.NewTransfersRepositoryMock(t)
		transfersPublisher = mocks.NewTransfersPublisherMock(t)
	)

	for _, testCase := range []struct {
		name        string
		transferID  string
		repoError   error
		expectError bool
	}{
		{
			name:        "Transfer successfully deleted from repository and cache",
			transferID:  "TRF-20250326-DELETE-001",
			repoError:   nil,
			expectError: false,
		},
		{
			name:        "Error deleting transfer from repository due to database issue",
			transferID:  "TRF-20250326-DELETE-002",
			repoError:   assert.AnError,
			expectError: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			transfersRepo.On("Delete", ctx, testCase.transferID).Return(testCase.repoError)
			if testCase.repoError == nil {
				transfersCCache.On("Delete", ctx, testCase.transferID).Return(nil)
				transfersPublisher.On("Publish", "deleted", testCase.transferID).Return(nil)
			}
			svc := NewTransfersService(cfg, transfersRepo, transfersCCache, transfersPublisher)

			err := svc.Delete(ctx, testCase.transferID)

			if testCase.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			time.Sleep(100 * time.Millisecond)
		})
	}
}
