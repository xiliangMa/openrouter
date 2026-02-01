package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"massrouter.ai/backend/internal/model"
)

type paymentRecordRepository struct {
	*GormRepository[model.PaymentRecord]
}

func NewPaymentRecordRepository(db *gorm.DB) PaymentRecordRepository {
	return &paymentRecordRepository{
		GormRepository: NewGormRepository[model.PaymentRecord](db),
	}
}

func (r *paymentRecordRepository) FindByUserID(ctx context.Context, userID string) ([]*model.PaymentRecord, error) {
	var records []*model.PaymentRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&records).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find payment records by user: %w", err)
	}
	return records, nil
}

func (r *paymentRecordRepository) FindByTransactionID(ctx context.Context, transactionID string) (*model.PaymentRecord, error) {
	var record model.PaymentRecord
	err := r.db.WithContext(ctx).
		Where("transaction_id = ?", transactionID).
		First(&record).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find payment record by transaction ID: %w", err)
	}
	return &record, nil
}

func (r *paymentRecordRepository) FindByStatus(ctx context.Context, status string) ([]*model.PaymentRecord, error) {
	var records []*model.PaymentRecord
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&records).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find payment records by status: %w", err)
	}
	return records, nil
}

func (r *paymentRecordRepository) UpdateStatus(ctx context.Context, paymentID, status string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if paidAt != nil {
		updates["paid_at"] = paidAt
	}

	result := r.db.WithContext(ctx).
		Model(&model.PaymentRecord{}).
		Where("id = ?", paymentID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update payment status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("payment record not found")
	}
	return nil
}

func (r *paymentRecordRepository) GetUserTotalPaid(ctx context.Context, userID string) (float64, error) {
	var totalPaid float64

	err := r.db.WithContext(ctx).
		Model(&model.PaymentRecord{}).
		Where("user_id = ? AND status = ?", userID, "completed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalPaid).Error

	if err != nil {
		return 0, fmt.Errorf("failed to get user total paid: %w", err)
	}
	return totalPaid, nil
}
