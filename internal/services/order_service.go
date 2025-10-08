package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"delivery-management/internal/cache"
	"delivery-management/internal/db"
	"delivery-management/internal/models"
	"gorm.io/gorm"
)

type OrderService struct {
	db        *db.Database
	cache     *cache.Cache
	processor *OrderProcessor
}

type OrderProcessor struct {
	workerCount int
	jobQueue    chan *models.Order
	quit        chan bool
	wg          sync.WaitGroup
	cache       *cache.Cache
}

func NewOrderService(database *db.Database, redisCache *cache.Cache) *OrderService {
	processor := &OrderProcessor{
		workerCount: 5,
		jobQueue:    make(chan *models.Order, 100),
		quit:        make(chan bool),
		cache:       redisCache,
	}

	service := &OrderService{
		db:        database,
		cache:     redisCache,
		processor: processor,
	}

	// Start order processor workers
	service.processor.start()
	
	return service
}

func (p *OrderProcessor) start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *OrderProcessor) worker(id int) {
	defer p.wg.Done()
	log.Printf("Order processor worker %d started", id)

	for {
		select {
		case order := <-p.jobQueue:
			p.processOrder(order)
		case <-p.quit:
			log.Printf("Order processor worker %d stopping", id)
			return
		}
	}
}

func (p *OrderProcessor) processOrder(order *models.Order) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in order processing: %v", r)
		}
	}()

	ctx := context.Background()
	
	// Simulate order progression: created -> dispatched -> in_transit -> delivered
	statuses := []models.OrderStatus{
		models.StatusDispatched,
		models.StatusInTransit,
		models.StatusDelivered,
	}

	for _, status := range statuses {
		// Wait before transitioning (simulate real processing time)
		time.Sleep(10 * time.Second)

		// Check if order was cancelled
		if order.Status == models.StatusCancelled {
			return
		}

		// Update order status
		order.Status = status
		
		// Publish status update
		statusUpdate := map[string]interface{}{
			"order_id": order.ID,
			"status":   status,
			"timestamp": time.Now(),
		}
		
		if data, err := json.Marshal(statusUpdate); err == nil {
			p.cache.Publish(ctx, fmt.Sprintf("order:%d", order.ID), string(data))
			p.cache.Publish(ctx, "orders:updates", string(data))
		}

		log.Printf("Order %d status updated to %s", order.ID, status)
	}
}

func (p *OrderProcessor) stop() {
	close(p.quit)
	p.wg.Wait()
	log.Println("All order processor workers stopped")
}

func (s *OrderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest, customerID uint) (*models.OrderResponse, error) {
	order := &models.Order{
		CustomerID:  customerID,
		Status:      models.StatusCreated,
		Items:       req.Items,
		Description: req.Description,
		Address:     req.Address,
	}

	if err := s.db.WithContext(ctx).Create(order).Error; err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Load customer information
	if err := s.db.WithContext(ctx).Preload("Customer").First(order, order.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load order with customer: %w", err)
	}

	// Queue order for processing
	select {
	case s.processor.jobQueue <- order:
		log.Printf("Order %d queued for processing", order.ID)
	default:
		log.Printf("Order processing queue is full, order %d will be processed later", order.ID)
	}

	return order.ToResponse(), nil
}

func (s *OrderService) GetOrdersByCustomer(ctx context.Context, customerID uint) ([]*models.OrderResponse, error) {
	var orders []models.Order
	if err := s.db.WithContext(ctx).Preload("Customer").Where("customer_id = ?", customerID).Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	responses := make([]*models.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = order.ToResponse()
	}

	return responses, nil
}

func (s *OrderService) GetAllOrders(ctx context.Context) ([]*models.OrderResponse, error) {
	var orders []models.Order
	if err := s.db.WithContext(ctx).Preload("Customer").Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to get all orders: %w", err)
	}

	responses := make([]*models.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = order.ToResponse()
	}

	return responses, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, orderID uint, customerID uint, isAdmin bool) (*models.OrderResponse, error) {
	var order models.Order
	query := s.db.WithContext(ctx).Preload("Customer")
	
	if !isAdmin {
		query = query.Where("customer_id = ?", customerID)
	}
	
	if err := query.First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order.ToResponse(), nil
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID uint, customerID uint, isAdmin bool) error {
	var order models.Order
	query := s.db.WithContext(ctx)
	
	if !isAdmin {
		query = query.Where("customer_id = ?", customerID)
	}
	
	if err := query.First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("order not found")
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	if !order.Status.CanTransitionTo(models.StatusCancelled) {
		return fmt.Errorf("cannot cancel order in %s status", order.Status)
	}

	order.Status = models.StatusCancelled
	if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Publish cancellation update
	statusUpdate := map[string]interface{}{
		"order_id": order.ID,
		"status":   models.StatusCancelled,
		"timestamp": time.Now(),
	}
	
	if data, err := json.Marshal(statusUpdate); err == nil {
		s.cache.Publish(ctx, fmt.Sprintf("order:%d", order.ID), string(data))
		s.cache.Publish(ctx, "orders:updates", string(data))
	}

	return nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID uint, req *models.UpdateOrderStatusRequest) error {
	var order models.Order
	if err := s.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("order not found")
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	if !req.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", req.Status)
	}

	if !order.Status.CanTransitionTo(req.Status) {
		return fmt.Errorf("cannot transition from %s to %s", order.Status, req.Status)
	}

	order.Status = req.Status
	if err := s.db.WithContext(ctx).Save(&order).Error; err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Publish status update
	statusUpdate := map[string]interface{}{
		"order_id": order.ID,
		"status":   req.Status,
		"timestamp": time.Now(),
	}
	
	if data, err := json.Marshal(statusUpdate); err == nil {
		s.cache.Publish(ctx, fmt.Sprintf("order:%d", order.ID), string(data))
		s.cache.Publish(ctx, "orders:updates", string(data))
	}

	return nil
}

func (s *OrderService) Stop() {
	s.processor.stop()
}