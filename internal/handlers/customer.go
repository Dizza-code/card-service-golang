package handlers

import (
	"card-service/internal/api"
	"card-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// customerHandler handles customer-related HTTP requests.
type CustomerHandler struct {
	customerService          *services.CustomerService // CustomerService instance
	defaultSettlementAccount string
	Logger                   *zap.Logger
}

// NewCustomerHandler creates a new customer handler.
func NewCustomerHandler(customerService *services.CustomerService, defaultSettlementAccount string, logger *zap.Logger) *CustomerHandler {
	return &CustomerHandler{
		customerService:          customerService,
		defaultSettlementAccount: defaultSettlementAccount,
		Logger:                   logger,
	}
}

type CreateCustomerRequest struct {
	Name            string `json:"name" binding:"required"`
	FirstName       string `json:"firstName" binding:"required"`
	LastName        string `json:"lastName" binding:"required"`
	MiddleName      string `json:"middleName"`
	Email           string `json:"email" binding:"required,email"`
	PhoneNumber     string `json:"phoneNumber" binding:"required"`
	Title           string `json:"title" binding:"required,oneof=Mr Ms Mrs Dr"`
	Gender          string `json:"gender" binding:"required,oneof=M F"`
	DateOfBirth     string `json:"dateOfBirth" binding:"required"`
	NationalityCode string `json:"nationalityCode" binding:"required,len=2"`
	IDType          string `json:"idType" binding:"required,oneof=bvn nin"`
	IDNumber        string `json:"idNumber" binding:"required"`
	IssuingCountry  string `json:"issuingCountry" binding:"required,len=2"`
	UserID          int    `json:"userId" binding:"required"`
	Ref             string `json:"ref" binding:"required"`
	// SettlementAccount string `json:"settlementAccount"`
}

type CreateCustomerResponse struct {
	CustomerID      string               `json:"customerId"`
	Name            string               `json:"name"`
	Email           string               `json:"email"`
	Balance         float64              `json:"balance"`
	AccountID       string               `json:"accountId"`
	DepositChannels []api.DepositChannel `json:"depositChannels"`
}

// CreateCustomer handles POST /api/customers request to create a new customer and sub account.
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {

	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.Logger.Info("Received CreateCustomer request",
		zap.String("email", req.Email),
		zap.String("phoneNumber", req.PhoneNumber),
		zap.String("name", req.Name),
	)

	// Use provided settlement account or default
	// settlementAccount := req.SettlementAccount
	// if settlementAccount == "" {
	// 	settlementAccount = h.defaultSettlementAccount
	// }

	// create customer and sub account using the service
	customerID, accountID, depositChannels, err := h.customerService.CreateCustomer(
		req.Name,
		req.FirstName,
		req.LastName,
		req.MiddleName,
		req.Email,
		req.PhoneNumber,
		req.Title,
		req.Gender,
		req.DateOfBirth,
		req.NationalityCode,
		req.IDType,
		req.IDNumber,
		req.IssuingCountry,
		req.UserID,
		req.Ref,
		// settlementAccount,
	)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	if err != nil {
		h.Logger.Error("Failed to create customer", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//Return success response
	response := CreateCustomerResponse{
		CustomerID: customerID,
		Name:       req.Name,
		Email:      req.Email,
		// Balance:         0,
		AccountID:       accountID,
		DepositChannels: depositChannels,
	}
	c.JSON(http.StatusCreated, response)
}
