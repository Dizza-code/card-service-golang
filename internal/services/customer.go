package services

import (
	"card-service/internal/api"
	"card-service/internal/models"
	"card-service/internal/store"
	"context"
	"time"

	"go.uber.org/zap"
)

//logic to create a customer and a sub account

//customer service handles customer and sub accounts operations

type CustomerService struct {
	store     *store.Store //MongoDB store
	apiClient *api.Client  //API client for external services
	logger    *zap.Logger  //Logger for logging
}

// NewCustomerService initializes a new CustomerService instance with the provided store, API client, and logger.
func NewCustomerService(store *store.Store, apiClient *api.Client, logger *zap.Logger) *CustomerService {
	return &CustomerService{store: store, apiClient: apiClient, logger: logger}
}

//CreateCustomer creatres a customer and a sub account and stores them in the mongoDB

func (s *CustomerService) CreateCustomer(name, firstName, lastName, middleName, email, phoneNumber, title, gender, dob, nationalityCode, idType, idNumber, issuingCountry string,
	userID int, ref string) (string, string, []api.DepositChannel, error) {
	s.logger.Info("Starting CreateCustomer",
		zap.String("email", email),
		zap.String("phoneNumber", phoneNumber),
		zap.String("gender", gender),
	)

	//start mongoDB transaction to ensure consistency
	session, err := s.store.Client.StartSession()
	if err != nil {
		s.logger.Error("Failed to start MongoDB session", zap.Error(err))
		return "", "", nil, err
	}
	defer session.EndSession(context.Background())

	req := api.CreateCustomerRequest{
		Name: name,
		Type: "individual",
		Claims: api.CustomerClaims{
			IndividualInformation: api.IndividualInformation{
				FirstName:       firstName,
				LastName:        lastName,
				MiddleName:      middleName,
				Email:           email,
				PhoneNumber:     phoneNumber,
				Title:           title,
				Gender:          gender,
				DateOfBirth:     dob,
				NationalityCode: nationalityCode,
			},
			IndividualIdentity: api.IndividualIdentity{
				Type:           idType,
				ID:             idNumber,
				IssuingCountry: issuingCountry,
			},
		},
		Verifications: []api.CustomerVerification{
			{Type: "tier-2", Status: "verified"},
		},
		Metadata: api.CustomerMetadata{
			UserID: userID,
			Ref:    ref,
		},
	}
	customerID, err := s.apiClient.CreateCustomer(req)
	if err != nil {
		s.logger.Error("Failed to create customer in Allawee API", zap.Error(err))
		return "", "", nil, err
	}
	s.logger.Info("Created customer", zap.String("customerID", customerID))

	//Create sub account
	vaReq := api.CreateSubAccountRequest{
		Name:            name,
		Type:            "sub",
		Currency:        "NGN",
		Customer:        customerID,
		DepositChannels: []string{"bank-account"},
		// SettlementAccount: "",
	}
	accountID, depositChannels, err := s.apiClient.CreateSubAccount(vaReq)
	if err != nil {
		s.logger.Error("Failed to create sub account in Allawee API", zap.Error(err))
		return "", "", nil, err
	}
	s.logger.Info("Created sub account", zap.String("subAccountID", accountID))

	//store customer in MongoDB
	customer := models.Customer{
		CustomerID: customerID,
		Name:       name,
		Email:      email,
		// Balance:      0, //initial balance is 0
		AccountID: accountID,
		CreatedAt: time.Now(),
	}
	_, err = s.store.Customers.InsertOne(context.Background(), customer)
	if err != nil {
		s.logger.Error("Failed to store customer in MongoDB", zap.Error(err))
		return "", "", nil, err
	}

	var depositChannelModels []models.DepositChannel
	for _, dc := range depositChannels {
		depositChannelModels = append(depositChannelModels, models.DepositChannel{
			AccountName:   dc.AccountName,
			AccountNumber: dc.AccountNumber,
			BankName:      dc.BankName,
			BankCode:      dc.BankCode,
			Type:          dc.Type,
		})
	}
	//store account in Mongo DB
	account := models.Account{
		AccountID:       accountID,
		CustomerID:      customerID,
		Name:            name,
		DepositChannels: depositChannelModels,
		Status:          "active", //default status is active
		CreatedAt:       time.Now(),
	}

	_, err = s.store.Accounts.InsertOne(context.Background(), account)
	if err != nil {
		s.logger.Error("failed to store account in MongoDb", zap.Error(err))
		return "", "", nil, err
	}

	return customerID, accountID, depositChannels, nil
}
