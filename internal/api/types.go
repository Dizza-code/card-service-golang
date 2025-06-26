package api

type CardControls struct {
	AllowedChannels   []string        `json:"allowedChannels" bson:"allowedChannels"`
	BlockedChannels   []string        `json:"blockedChannels" bson:"blockedChannels"`
	AllowedMerchants  []string        `json:"allowedMerchants" bson:"allowedMerchants"`
	BlockedMerchants  []string        `json:"blockedMerchants" bson:"blockedMerchants"`
	AllowedCategories []string        `json:"allowedCategories" bson:"allowedCategories"`
	BlockedCategories []string        `json:"blockedCategories" bson:"blockedCategories"`
	SpendingLimits    []SpendingLimit `json:"spendingLimits" bson:"spendingLimits"`
}

type SpendingLimit struct {
	Amount   int    `json:"amount"`
	Interval string `json:"interval"` // e.g., "daily", "weekly", "monthly"
}

type CardMetadata struct {
	Name string `json:"name,omitempty" bson:"name"`
}

type CardDetails struct {
	Last4          string `json:"last4" bson:"last4"`
	Expiry         string `json:"expiry" bson:"expiry"`
	CardHolderName string `json:"cardHolderName" bson:"cardHolderName"`
}
