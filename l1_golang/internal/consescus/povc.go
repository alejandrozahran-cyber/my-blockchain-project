package consensus

import (
	"encoding/json"
	"net/http"
	"time"
)

type PoVConsensus struct {
	TotalSupply      float64
	MonthlyReward    float64
	AIEngineURL      string
	httpClient       *http.Client
}

func NewPoVConsensus(totalSupply, monthlyReward float64, aiEngineURL string) *PoVConsensus {
	return &PoVConsensus{
		TotalSupply:   totalSupply,
		MonthlyReward: monthlyReward,
		AIEngineURL:   aiEngineURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type UserData struct {
	WalletAddress   string  `json:"wallet_address"`
	WalletBalance   float64 `json:"wallet_balance"`
	DailyActivity   float64 `json:"daily_activity"`
	Contributions   int     `json:"contributions"`
	CommunityScore  float64 `json:"community_score"`
	QualityScore    float64 `json:"quality_score"`
	DaysActive      int     `json:"days_active"`
}

type PoVCResponse struct {
	Success bool        `json:"success"`
	Data    RewardData  `json:"data"`
}

type RewardData struct {
	Wallet      string  `json:"wallet"`
	NVSScore    float64 `json:"nvs_score"`
	BaseReward  float64 `json:"base_reward"`
	FinalReward float64 `json:"final_reward"`
	Timestamp   string  `json:"timestamp"`
}

func (p *PoVConsensus) CalculateReward(userData UserData) (*RewardData, error) {
	// Prepare request to AI Engine
	requestBody, err := json.Marshal(userData)
	if err != nil {
		return nil, err
	}

	// Call AI Engine
	resp, err := p.httpClient.Post(p.AIEngineURL+"/povc/calculate", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result PoVCResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("AI Engine returned error")
	}

	return &result.Data, nil
}

func (p *PoVConsensus) AntiWhaleCheck(walletBalance float64) (float64, float64) {
	// Calculate percentage of total supply
	percentage := (walletBalance / p.TotalSupply) * 100
	
	// Anti-whale rules
	var rewardMultiplier float64 = 1.0
	var transferFee float64 = 0

	// Tier 1: 0.5% - 1%
	if percentage > 0.5 {
		reduction := min((percentage-0.5)*2, 50)
		rewardMultiplier = 1 - (reduction / 100)
		transferFee = 1
	}

	// Tier 2: 1% - 2%
	if percentage > 1 {
		reduction := 50 + min((percentage-1)*25, 50)
		rewardMultiplier = 1 - (reduction / 100)
		transferFee = 3
	}

	// Tier 3: >2%
	if percentage > 2 {
		rewardMultiplier = 0
		transferFee = 10
	}

	return rewardMultiplier, transferFee
}

func (p *PoVConsensus) BatchCalculate(users []UserData) ([]RewardData, error) {
	// Prepare batch request
	type BatchRequest struct {
		Users []UserData `json:"users"`
	}

	requestBody, err := json.Marshal(BatchRequest{Users: users})
	if err != nil {
		return nil, err
	}

	// Call AI Engine
	resp, err := p.httpClient.Post(p.AIEngineURL+"/povc/batch", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool        `json:"success"`
		Data    struct {
			IndividualRewards []RewardData `json:"individual_rewards"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("AI Engine batch processing failed")
	}

	return result.Data.IndividualRewards, nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}