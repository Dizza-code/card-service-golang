package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
	logger  *zap.Logger
}

// create a new client API
func NewClient(baseURL, apiKey string) *Client {
	logger := zap.NewExample()
	logger.Info("Initializing API client", zap.String("baseURL", baseURL), zap.String("apiKeyPrefix", apiKey[:4]))
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{},
		logger:  logger,
	}
}

type IndividualInformation struct {
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	MiddleName      string `json:"middleName"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phoneNumber"`
	Title           string `json:"title"`
	Gender          string `json:"gender"`
	DateOfBirth     string `json:"dateOfBirth"`
	NationalityCode string `json:"nationalityCode"`
}
type IndividualIdentity struct {
	Type           string `json:"type"`
	ID             string `json:"id"`
	IssuingCountry string `json:"issuingCountry"`
}
type CustomerClaims struct {
	IndividualInformation IndividualInformation `json:"individualInformation"`
	IndividualIdentity    IndividualIdentity    `json:"individualIdentity"`
}

//create customer request struct

type CreateCustomerRequest struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Claims        CustomerClaims         `json:"claims"`
	Verifications []CustomerVerification `json:"verifications,omitempty"`
	Metadata      CustomerMetadata       `json:"metadata"`
}

// customer verification
type CustomerVerification struct {
	Type   string `json:"type"`   // e.g., "tier-2
	Status string `json:"status"` // e.g., "pending", "verified"
}

// customer metadata
type CustomerMetadata struct {
	UserID int    `json:"userId"` // ID of the user in the system
	Ref    string `json:"ref"`    // Reference ID for the customer
}

// custome response - This holds the response from the create customer request
type CreateCustomerResponse struct {
	Code string `json:"code"`
	Data struct {
		ID string `json:"id"`
	} `json:"data"`
}

// create customer sends a request to create a customer
func (c *Client) CreateCustomer(req CreateCustomerRequest) (string, error) {
	// marshal the request payload to json
	body, err := json.Marshal(req)
	if err != nil {
		c.logger.Error("Failed to marshal CreateCustomer request", zap.Error(err))
		return "", err
	}
	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.baseURL+"/customers", bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error("Failed to create CreateCustomer HTTP request", zap.Error(err))
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Log request details
	c.logger.Info("Sending CreateCustomer request",
		zap.String("url", httpReq.URL.String()),
		zap.ByteString("body", body),
		zap.String("authHeader", "Bearer "+c.apiKey[:4]+"..."), // Log key prefix
	)

	//send request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("Failed to send CreateCustomer request", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read CreateCustomer response", zap.Error(err))
		return "", err
	}

	// Log response
	c.logger.Debug("Received CreateCustomer response", zap.Int("status", resp.StatusCode), zap.ByteString("body", respBody))

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		c.logger.Error("CreateCustomer request failed",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return "", fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(respBody))
	}

	//parse the response
	// var result CreateCustomerResponse
	// if err := json.Unmarshal(respBody, &result); err != nil {
	// 	c.logger.Error("Failed to parse CreateCustomer response", zap.Error(err), zap.ByteString("response", respBody))
	// 	return "", err
	// }

	// if result.CustomerID == "" {
	// 	c.logger.Error("Empty customer ID in CreateCustomer response", zap.ByteString("response", respBody))
	// 	return "", fmt.Errorf("empty customer ID in response: %s", respBody)
	// }
	// return result.CustomerID, nil
	var createResp CreateCustomerResponse
	if err := json.Unmarshal(respBody, &createResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if createResp.Code != "success" {
		return "", fmt.Errorf("request failed: %s", string(respBody))
	}

	return createResp.Data.ID, nil
}

// create sub account request
type CreateSubAccountRequest struct {
	Name            string   `json:"name"`            // Name of the sub account
	Type            string   `json:"type"`            // Type of the sub account
	Currency        string   `json:"currency"`        // Currency of the sub account
	Customer        string   `json:"customer"`        // ID of the customer associated with the sub account
	DepositChannels []string `json:"depositChannels"` // Channels for deposits
	// SettlementAccount string   `json:"settlementAccount"` // Settlement accounts for the sub account
}

type DepositChannel struct {
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
	BankName      string `json:"bankName"`
	BankCode      string `json:"bankCode"`
	Type          string `json:"type"`
}

// create sub account response
type CreateSubAccountResponse struct {
	Code string `json:"code"`
	Data struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
		// SettlementAccount string           `json:"settlementAccount"`
		Status          string           `json:"status"`
		Currency        string           `json:"currency"`
		Customer        string           `json:"customer"`
		DepositChannels []DepositChannel `json:"depositChannels"`
	} `json:"data"`
}

// create sub account sends a request to create a sub account
func (c *Client) CreateSubAccount(req CreateSubAccountRequest) (string, []DepositChannel, error) {
	//marshal the payload to json
	body, err := json.Marshal(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.baseURL+"/accounts", bytes.NewBuffer(body))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	//send request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("Failed to send request", zap.Error(err))
		return "", nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	c.logger.Info("Sending CreateSubAccount request",
		zap.String("url", httpReq.URL.String()),
		zap.String("body", string(body)),
		zap.String("authHeader", "Bearer "+c.apiKey[:4]+"..."),
		zap.Any("headers", httpReq.Header),
	)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", zap.Error(err))
		return "", nil, fmt.Errorf("failed to read response: %w", err)
	}
	if len(respBody) == 0 {
		c.logger.Warn("Received empty response body")
	}

	c.logger.Debug("Received CreateSubAccount response",
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	//accept 201 and 202 status codes
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		c.logger.Error("CreateSubAccount request failed",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return "", nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(respBody))
	}

	var createResp CreateSubAccountResponse
	if err := json.Unmarshal(respBody, &createResp); err != nil {
		c.logger.Error("Failed to unmarshal response", zap.Error(err))
		return "", nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if createResp.Code != "success" {
		c.logger.Error("Request failed", zap.String("response", string(respBody)))
		return "", nil, fmt.Errorf("request failed: %s", string(respBody))
	}
	if createResp.Data.ID == "" {
		c.logger.Error("Sub account ID is empty in response", zap.String("response", string(respBody)))
		return "", nil, fmt.Errorf("Sub account ID is empty in response")
	}

	return createResp.Data.ID, createResp.Data.DepositChannels, nil
}

//create card linking

type CardControls struct {
	AllowedChannels  []string         `json:"allowedChannels,omitempty"`
	BlockedChannels  []string         `json:"blockedChannels,omitempty"`
	AllowedMerchants []string         `json:"allowedMerchants,omitempty"`
	BlockedMerchants []string         `json:"blockedMerchants,omitempty"`
	SpendingLimits   []SpendingLimits `json:"spendingLimits,omitempty"`
}
type SpendingLimits struct {
	Amount   int    `json:"amount"`
	Interval string `json:"interval"` // e.g., "daily", "weekly", "monthly"
}
type CardMetadata struct {
	Name string `json:"name,omitempty"` // Name of the cardholder
}

type LinkCardRequest struct {
	Pan       string        `json:"pan"`                 // Primary Account Number (PAN) of the card
	Customer  string        `json:"customer"`            // ID of the customer associated with the card
	Reference string        `json:"reference,omitempty"` // Reference ID for the card
	Controls  *CardControls `json:"controls,omitempty"`  // Card controls for spending limits and channels
	Metadata  *CardMetadata `json:"metadata,omitempty"`  // Metadata for the card
}

type CardDetails struct {
	Last4          string `json:"last4"`          // Last 4 digits of the card
	ExpiryDate     string `json:"expiryDate"`     // Expiry date of the card in YYYY-MM format
	CardHolderName string `json:"cardHolderName"` // Name of the cardholder
}

type LinkCardResponse struct {
	Code string `json:"code"` // Response code
	Data struct {
		ID            string       `json:"id"`          // Unique identifier for the linked card
		Customer      string       `json:"customer"`    // ID of the customer associated with the card
		CardDetails   CardDetails  `json:"cardDetails"` // Details of the linked card
		Type          string       `json:"type"`        // Type of the card (e.g., "sub", "physical")
		Status        string       `json:"status"`      // Status of the linked card (e.g., "active", "inactive")
		Currency      string       `json:"currency"`    // Currency of the card
		Controls      CardControls `json:"controls"`
		Metadata      CardMetadata `json:"metadata"`      // Metadata for the card
		FundingSource string       `json:"fundingSource"` // Funding source for the card
		Reference     string       `json:"reference"`     // Reference ID for the card
		CreatedAt     string       `json:"createdAt"`     // Creation timestamp of the linked card
		UpdatedAt     string       `json:"updatedAt"`     // Last update timestamp of the linked card

	} `json:"data"` // Data containing the linked card details
}

type AllaweeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *Client) LinkCard(req LinkCardRequest) (LinkCardResponse, error) {
	var response LinkCardResponse
	body, err := json.Marshal(req)
	if err != nil {
		c.logger.Error("Failed to marshal LinkCard request", zap.Error(err))
		return response, fmt.Errorf("failed to marshal request: %w", err)
	}
	//create request
	httpReq, err := http.NewRequest("POST", c.baseURL+"/cards/link", bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error("failed to create LinkCard HTTP request", zap.Error(err))
		return response, fmt.Errorf("failed to create Linkcard request: %w", err)

	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	c.logger.Info("Sending LinkCard request",
		zap.String("url", httpReq.URL.String()),
		zap.String("body", string(body)),
		zap.String("authHeader", "Bearer "+c.apiKey[:4]+"..."),
		zap.Any("headers", httpReq.Header),
	)

	//send request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("Failed to send LinkCard request", zap.Error(err))
		return response, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	c.logger.Debug("Received response headers",
		zap.Any("headers", resp.Header),
		zap.Int64("contentLength", resp.ContentLength),
	)
	//Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read LinkCard response body", zap.Error(err))
		return response, fmt.Errorf("failed to read response: %w", err)
	}

	if len(respBody) == 0 {
		c.logger.Warn("Received empty response body")
		return response, fmt.Errorf("received empty response body")
	}

	c.logger.Debug("Received LinkCard response",
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var allaweeErr AllaweeError
		if err := json.Unmarshal(respBody, &allaweeErr); err == nil && allaweeErr.Code != "" {
			c.logger.Error("LinkCard request failed",
				zap.Int("status", resp.StatusCode),
				zap.String("code", allaweeErr.Code),
				zap.String("message", allaweeErr.Message),
			)
			return response, fmt.Errorf("allawee error: %s - %s", allaweeErr.Code, allaweeErr.Message)
		}
		c.logger.Error("LinkCard request failed",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return response, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(respBody))
	}

	// Unmarshal the successful response
	if err := json.Unmarshal(respBody, &response); err != nil {
		c.logger.Error("Failed to unmarshal LinkCard response", zap.Error(err))
		return response, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}
