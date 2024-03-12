package domain

// OrderStatus status for the submitted order.
type OrderStatus string

const (
	// OrderStatusPending represents pending status for submitted order.
	OrderStatusPending OrderStatus = "PENDING"
	// OrderStatusProcessing represents processing status for submitted order.
	OrderStatusProcessing OrderStatus = "PROCESSING"
	// OrderStatusCompleted represents completed status for submitted order.
	OrderStatusCompleted OrderStatus = "COMPLETED"
	// OrderStatusFailed represents failed status for submitted order.
	OrderStatusFailed OrderStatus = "FAILED"
)

// OrderDetails contains order details for the submitted certificate request to a Certificate Authority
type OrderDetails struct {
	ID            string      `json:"id"`
	Status        OrderStatus `json:"status"`
	CertificateID string      `json:"certificateId"`
	ErrorMessage  string      `json:"errorMessage"`
}
