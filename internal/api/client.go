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
	baseURL       string
	secureBaseURL string
	apiKey        string
	client        *http.Client
	logger        *zap.Logger
}

// create a new client API
func NewClient(baseURL, secureBaseURL, apiKey string) *Client {
	logger := zap.NewExample()
	logger.Info("Initializing API client", zap.String("baseURL", baseURL), zap.String("secureBaseURL", secureBaseURL), zap.String("apiKeyPrefix", apiKey[:4]))
	return &Client{
		baseURL:       baseURL,
		secureBaseURL: secureBaseURL,
		apiKey:        apiKey,
		client:        &http.Client{},
		logger:        logger,
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

// create card linking

type LinkCardRequest struct {
	Pan           string        `json:"pan"`      // Primary Account Number (PAN) of the card
	Customer      string        `json:"customer"` // ID of the customer associated with the card
	FundingSource string        `json:"fundingSource"`
	Reference     string        `json:"reference,omitempty"` // Reference ID for the card
	Controls      *CardControls `json:"controls,omitempty"`  // Card controls for spending limits and channels
	Metadata      *CardMetadata `json:"metadata,omitempty"`  // Metadata for the card
}

type LinkCardResponse struct {
	Code string `json:"code"` // Response code
	Data struct {
		ID            string       `json:"id"`       // Unique identifier for the linked card
		Customer      string       `json:"customer"` // ID of the customer associated with the card
		Details       CardDetails  `json:"details"`  // Details of the linked card
		Program       string       `json:"program"`
		Type          string       `json:"type"`     // Type of the card (e.g., "sub", "physical")
		Status        string       `json:"status"`   // Status of the linked card (e.g., "active", "inactive")
		Currency      string       `json:"currency"` // Currency of the card
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

type ActivateCardRequest struct {
	Cvv string `json:"cvv"` // Card Verification Value
	Pin string `json:"pin"` //
}
type ActivateCardResponse struct {
	Code    string `json:"code"`    // Response code
	Message string `json:"message"` // Response message
}

func (c *Client) LinkCard(req LinkCardRequest) (LinkCardResponse, error) {
	var response LinkCardResponse

	if req.Pan == "" || req.Customer == "" {
		c.logger.Error("Invalid Linkcard request", zap.String("pan", req.Pan), zap.String("customer", req.Customer))
		return response, fmt.Errorf("pan and customer required")
	}
	body, err := json.Marshal(req)
	if err != nil {
		c.logger.Error("Failed to marshal LinkCard request", zap.Error(err))
		return response, fmt.Errorf("failed to marshal request: %w", err)
	}
	//create request
	httpReq, err := http.NewRequest("POST", c.secureBaseURL+"/cards/link", bytes.NewBuffer(body))
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

func (c *Client) ActivateCard(CardID string, req ActivateCardRequest) (ActivateCardResponse, error) {
	var response ActivateCardResponse
	if CardID == "" || req.Cvv == "" || req.Pin == "" {
		c.logger.Error("Invalid ActivateCard request",
			zap.String("CardID", CardID),
			zap.String("cvv", req.Cvv),
			zap.String("pin", req.Pin),
		)
		return response, fmt.Errorf("CardID, cvv and pin are required")
	}
	body, err := json.Marshal(req)
	if err != nil {
		c.logger.Error("Failed to marshal ActivateCard request", zap.Error(err))
		return response, fmt.Errorf("failed to marshal request: %w", err)
	}
	//create request
	httpReq, err := http.NewRequest("POST", c.secureBaseURL+"/cards/"+CardID+"/activate", bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error("fialed to create activate card HTTP request", zap.Error(err))
		return response, fmt.Errorf("failed to create ActivateCard request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	c.logger.Info("Sending Activate Card request",
		zap.String("url", httpReq.URL.String()),
		zap.String("body", string(body)),
		zap.String("authHeader", "Bearer "+c.apiKey[:4]+"..."),
		zap.Any("headers", httpReq.Header),
	)
	// send request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("Failed to send ActivateCard request", zap.Error(err))
		return response, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	c.logger.Debug("Received response headers",
		zap.Any("headers", resp.Header),
		zap.Int64("contentLength", resp.ContentLength),
	)
	// Read response body
	respBody, err := io.ReadAll((resp.Body))
	if err != nil {
		c.logger.Error("Failed to read ActivateCard response body", zap.Error(err))
		return response, fmt.Errorf("failed to read response: %w", err)
	}
	c.logger.Info("Received ActivateCard response",
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var allaweeErr AllaweeError
		if err := json.Unmarshal(respBody, &allaweeErr); err == nil && allaweeErr.Code != "" {
			c.logger.Error("ActivateCard request failed",
				zap.Int("status", resp.StatusCode),
				zap.String("code", allaweeErr.Code),
				zap.String("message", allaweeErr.Message),
			)
			return response, fmt.Errorf("allawee error: %s - %s", allaweeErr.Code, allaweeErr.Message)
		}
		c.logger.Error("ActivateCard request failed",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return response, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(respBody))
	}

	//unmarshal the successful response
	if err := json.Unmarshal(respBody, &response); err != nil {
		c.logger.Error("Failed to unmarshal ActivateCard response", zap.Error(err))
		return response, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	c.logger.Info("ActivateCard request successful",
		zap.String("cardID", CardID),
		zap.String("responseCode", response.Code),
		zap.String("responseMessage", response.Message),
	)
	if response.Code != "success" {
		c.logger.Error("ActivateCard request failed",
			zap.String("response", string(respBody)),
		)
		return response, fmt.Errorf("request failed: %s", string(respBody))
	}
	if response.Message == "" {
		c.logger.Error("ActivateCard response message is empty", zap.String("response", string(respBody)))
		return response, fmt.Errorf("ActivateCard response message is empty")
	}
	return response, nil
}

func (c *Client) GetAccountBalance(accountID string) (GetAccountBalanceResponse, error) {
	var response GetAccountBalanceResponse
	if accountID == "" {
		c.logger.Error("Invalid GetAccountBalance request", zap.String("accountID", accountID))
		return response, fmt.Errorf("accountID is required")
	}

	httpReq, err := http.NewRequest("GET", c.baseURL+"/accounts/"+accountID+"/balance", nil)
	if err != nil {
		return response, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	c.logger.Info("Sending GetAccountBalance request",
		zap.String("url", httpReq.URL.String()),
		zap.String("authHeader", "Bearer "+c.apiKey[:4]+"..."),
	)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.Error("Failed to send request", zap.Error(err))
		return response, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", zap.Error(err))
		return response, fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.Debug("Received GetAccountBalance response",
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	if resp.StatusCode != http.StatusOK {
		var allaweeErr AllaweeError
		if err := json.Unmarshal(respBody, &allaweeErr); err == nil && allaweeErr.Code != "" {
			c.logger.Error("GetAccountBalance request failed",
				zap.Int("status", resp.StatusCode),
				zap.String("code", allaweeErr.Code),
				zap.String("message", allaweeErr.Message),
			)
			return response, fmt.Errorf("allawee error: %s - %s", allaweeErr.Code, allaweeErr.Message)
		}
		c.logger.Error("GetAccountBalance request failed",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(respBody)),
		)
		return response, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(respBody))
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		c.logger.Error("Failed to unmarshal response", zap.Error(err))
		return response, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Code != "success" {
		c.logger.Error("Invalid response", zap.String("response", string(respBody)))
		return response, fmt.Errorf("request failed: %s", string(respBody))
	}

	return response, nil
}
