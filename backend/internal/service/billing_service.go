package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"massrouter.ai/backend/internal/model"
	"massrouter.ai/backend/internal/repository"
	"massrouter.ai/backend/pkg/cache"

	"github.com/redis/go-redis/v9"
)

const (
	billingQueueKey = "queue:billing:records"
)

type billingService struct {
	paymentRepo     repository.PaymentRecordRepository
	billingRepo     repository.BillingRecordRepository
	modelRepo       repository.ModelRepository
	redisClient     *cache.RedisClient
	stopChan        chan struct{}
	mu              sync.RWMutex
	workerRunning   bool
	lastProcessedAt time.Time
	totalProcessed  int64
	errorsLastHour  int
	processingTimes []time.Duration
}

func NewBillingService(
	paymentRepo repository.PaymentRecordRepository,
	billingRepo repository.BillingRecordRepository,
	modelRepo repository.ModelRepository,
	redisClient *cache.RedisClient,
) BillingService {
	return &billingService{
		paymentRepo:   paymentRepo,
		billingRepo:   billingRepo,
		modelRepo:     modelRepo,
		redisClient:   redisClient,
		stopChan:      make(chan struct{}),
		workerRunning: false,
	}
}

func (s *billingService) GetBalance(ctx context.Context, userID string) (*BalanceInfo, error) {
	totalPaid, err := s.paymentRepo.GetUserTotalPaid(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total paid: %w", err)
	}

	totalUsed, err := s.billingRepo.GetTotalCostByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total used: %w", err)
	}

	balance := totalPaid - totalUsed
	isOverdue := balance < 0

	var nextBilling *time.Time
	if balance > 0 {
		thirtyDaysLater := time.Now().Add(30 * 24 * time.Hour)
		nextBilling = &thirtyDaysLater
	}

	return &BalanceInfo{
		Balance:     balance,
		CreditLimit: 0,
		NextBilling: nextBilling,
		IsOverdue:   isOverdue,
	}, nil
}

func (s *billingService) GetPaymentHistory(ctx context.Context, userID string, page, limit int) (*PaymentHistoryResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	payments, err := s.paymentRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment history: %w", err)
	}

	total := int64(len(payments))

	start := offset
	if start > len(payments) {
		start = len(payments)
	}
	end := start + limit
	if end > len(payments) {
		end = len(payments)
	}

	pagedPayments := payments[start:end]

	paymentItems := make([]*PaymentItem, len(pagedPayments))
	for i, payment := range pagedPayments {
		paymentItems[i] = &PaymentItem{
			ID:            payment.ID,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			PaymentMethod: payment.PaymentMethod,
			Status:        payment.Status,
			TransactionID: payment.TransactionID,
			CreatedAt:     payment.CreatedAt,
			PaidAt:        payment.PaidAt,
		}
	}

	return &PaymentHistoryResponse{
		Payments: paymentItems,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}, nil
}

func (s *billingService) CreatePayment(ctx context.Context, userID string, req *CreatePaymentRequest) (*PaymentInfo, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("payment amount must be positive")
	}

	payment := &model.PaymentRecord{
		UserID:        userID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		Status:        "pending",
		Metadata:      model.JSONB{"return_url": req.ReturnURL},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	expiresAt := time.Now().Add(30 * time.Minute)

	return &PaymentInfo{
		ID:         payment.ID,
		Amount:     payment.Amount,
		Currency:   payment.Currency,
		Status:     payment.Status,
		PaymentURL: fmt.Sprintf("/payments/%s/process", payment.ID),
		ExpiresAt:  &expiresAt,
	}, nil
}

func (s *billingService) ProcessPaymentWebhook(ctx context.Context, payload []byte, signature string) error {
	return nil
}

func (s *billingService) GetBillingRecords(ctx context.Context, userID string, page, limit int) (*BillingRecordsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	records, err := s.billingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing records: %w", err)
	}

	total := int64(len(records))

	start := offset
	if start > len(records) {
		start = len(records)
	}
	end := start + limit
	if end > len(records) {
		end = len(records)
	}

	pagedRecords := records[start:end]

	recordItems := make([]*BillingRecordItem, len(pagedRecords))
	for i, record := range pagedRecords {
		modelObj, _ := s.modelRepo.FindByID(ctx, record.ModelID)
		modelName := "Unknown"
		providerName := "Unknown"
		if modelObj != nil {
			modelName = modelObj.Name
			providerName = modelObj.Provider.Name
		}

		recordItems[i] = &BillingRecordItem{
			BillingRecord: record,
			ModelName:     modelName,
			ProviderName:  providerName,
		}
	}

	return &BillingRecordsResponse{
		Records: recordItems,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

func (s *billingService) CalculateCost(ctx context.Context, modelID string, inputTokens, outputTokens int) (*CostCalculation, error) {
	modelObj, err := s.modelRepo.FindByID(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}
	if modelObj == nil {
		return nil, fmt.Errorf("model not found")
	}

	inputCost := float64(inputTokens) * modelObj.InputPrice / 1000.0
	outputCost := float64(outputTokens) * modelObj.OutputPrice / 1000.0
	totalCost := inputCost + outputCost

	return &CostCalculation{
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		InputCost:    inputCost,
		OutputCost:   outputCost,
		TotalCost:    totalCost,
		ModelName:    modelObj.Name,
		ProviderName: modelObj.Provider.Name,
	}, nil
}

func (s *billingService) CreateBillingRecord(ctx context.Context, req *CreateBillingRecordRequest) error {
	// If Redis client is not available, fallback to synchronous creation
	if s.redisClient == nil {
		return s.createBillingRecordSync(ctx, req)
	}

	// Push to Redis queue for asynchronous processing
	jobData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal billing record: %w", err)
	}

	if err := s.redisClient.LPush(ctx, billingQueueKey, jobData); err != nil {
		// If queue fails, fallback to synchronous creation
		return s.createBillingRecordSync(ctx, req)
	}

	return nil
}

func (s *billingService) createBillingRecordSync(ctx context.Context, req *CreateBillingRecordRequest) error {
	billingRecord := &model.BillingRecord{
		UserID:         req.UserID,
		APIKeyID:       req.APIKeyID,
		ModelID:        req.ModelID,
		RequestTokens:  req.RequestTokens,
		ResponseTokens: req.ResponseTokens,
		TotalTokens:    req.TotalTokens,
		Cost:           req.Cost,
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
	}

	if err := s.billingRepo.Create(ctx, billingRecord); err != nil {
		return fmt.Errorf("failed to create billing record: %w", err)
	}

	return nil
}

func (s *billingService) StartBillingWorker() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.workerRunning {
		return
	}

	s.workerRunning = true
	go s.processBillingQueue()
}

func (s *billingService) StopBillingWorker() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.workerRunning {
		return
	}

	close(s.stopChan)
	s.workerRunning = false
}

func (s *billingService) processBillingQueue() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
			// Try to process a job from the queue
			ctx := context.Background()

			// Use BRPop with timeout of 5 seconds to allow checking stopChan
			result, err := s.redisClient.BRPop(ctx, 5*time.Second, billingQueueKey)
			if err != nil {
				// If error is due to timeout (redis.Nil), continue to check stopChan
				if err == redis.Nil {
					continue
				}
				// Log error and continue
				fmt.Printf("Error reading from billing queue: %v\n", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if len(result) < 2 {
				continue
			}

			jobData := result[1]
			var req CreateBillingRecordRequest
			if err := json.Unmarshal([]byte(jobData), &req); err != nil {
				fmt.Printf("Failed to unmarshal billing record: %v\n", err)
				continue
			}

			// Process the billing record synchronously
			if err := s.createBillingRecordSync(ctx, &req); err != nil {
				fmt.Printf("Failed to create billing record: %v\n", err)
				// Optionally: push to dead letter queue or retry
				// For now, just log the error
			} else {
				fmt.Printf("Successfully processed billing record for user %s, cost: %.6f\n", req.UserID, req.Cost)
				// Update statistics
				s.mu.Lock()
				s.lastProcessedAt = time.Now()
				s.totalProcessed++
				s.mu.Unlock()
			}
		}
	}
}

func (s *billingService) GetQueueStatus(ctx context.Context) (*QueueStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// TODO: Implement Redis LLen method in RedisClient
	// For now, return 0 queue length
	queueLen := int64(0)

	// Calculate average processing time
	var avgProcessingTime float64
	if len(s.processingTimes) > 0 {
		var total time.Duration
		for _, t := range s.processingTimes {
			total += t
		}
		avgProcessingTime = total.Seconds() / float64(len(s.processingTimes))
	}

	return &QueueStatus{
		QueueLength:       queueLen,
		WorkerRunning:     s.workerRunning,
		LastProcessedAt:   s.lastProcessedAt,
		TotalProcessed:    s.totalProcessed,
		ErrorsLastHour:    s.errorsLastHour,
		AvgProcessingTime: avgProcessingTime,
	}, nil
}
