from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Dict, List, Optional
import json
import hashlib
from datetime import datetime, timedelta

app = FastAPI(
    title="NUSA Chain AI Engine",
    version="1.0.0",
    description="Proof of Value Creation (PoVC) AI Engine",
    docs_url="/docs",
    redoc_url="/redoc"
)

# ==================== PoVC Engine ====================
class NUSAValueScore:
    def __init__(self):
        self.total_supply = 25_000_000
        self.monthly_reward_pool = 100_000
        
    def calculate_nvs(self, user_data: Dict) -> Dict:
        """Calculate NUSA Value Score (0-1 scale)"""
        
        scores = {
            "activity": min(user_data.get("daily_activity", 0) / 240, 1.0),  # Max 4 hours
            "contributions": min(user_data.get("contributions", 0) / 50, 1.0),
            "community": min(user_data.get("community_score", 0) / 100, 1.0),
            "quality": user_data.get("quality_score", 0.5),
            "age_bonus": min(user_data.get("days_active", 0) / 365, 1.0) * 0.2
        }
        
        # Weighted average
        weights = {"activity": 0.3, "contributions": 0.4, "community": 0.2, "quality": 0.1}
        nvs = sum(scores[k] * weights.get(k, 0) for k in weights.keys())
        nvs = min(max(nvs, 0), 1.0)
        
        return {
            "nvs_score": round(nvs, 4),
            "breakdown": {k: round(v, 3) for k, v in scores.items()}
        }
    
    def anti_whale_check(self, wallet_balance: float) -> Dict:
        """Anti-whale mechanism with 3 tiers"""
        
        percentage = (wallet_balance / self.total_supply) * 100
        
        result = {
            "balance": wallet_balance,
            "percentage": round(percentage, 4),
            "reward_multiplier": 1.0,
            "transfer_fee": 0,
            "warnings": []
        }
        
        # Tier 1: 0.5% - 1%
        if percentage > 0.5:
            reduction = min((percentage - 0.5) * 2, 50)
            result["reward_multiplier"] = 1 - (reduction / 100)
            result["transfer_fee"] = 1
            result["warnings"].append("Reward reduction: 1-50%")
        
        # Tier 2: 1% - 2%  
        if percentage > 1:
            reduction = 50 + min((percentage - 1) * 25, 50)
            result["reward_multiplier"] = 1 - (reduction / 100)
            result["transfer_fee"] = 3
            result["warnings"].append("High concentration: 50-100% reduction")
        
        # Tier 3: >2%
        if percentage > 2:
            result["reward_multiplier"] = 0
            result["transfer_fee"] = 10
            result["warnings"].append("WHALE: No rewards")
        
        return result
    
    def calculate_reward(self, user_data: Dict) -> Dict:
        """Calculate monthly PoVC reward"""
        
        nvs_result = self.calculate_nvs(user_data)
        whale_check = self.anti_whale_check(user_data.get("wallet_balance", 0))
        
        base_reward = self.monthly_reward_pool * nvs_result["nvs_score"]
        final_reward = base_reward * whale_check["reward_multiplier"]
        
        return {
            "wallet": user_data.get("wallet_address", "unknown"),
            "nvs_score": nvs_result["nvs_score"],
            "base_reward": round(base_reward, 2),
            "final_reward": round(final_reward, 2),
            "whale_check": whale_check,
            "timestamp": datetime.now().isoformat()
        }

# Initialize engine
povc_engine = NUSAValueScore()

# ==================== API Models ====================
class UserRequest(BaseModel):
    wallet_address: str
    wallet_balance: float = 0
    daily_activity: float = 0
    contributions: int = 0
    community_score: float = 0
    quality_score: float = 0.5
    days_active: int = 0

class BatchRequest(BaseModel):
    users: List[UserRequest]
    simulation_id: Optional[str] = None

# ==================== API Endpoints ====================
@app.get("/")
async def root():
    return {
        "service": "NUSA Chain AI Engine",
        "version": "1.0.0",
        "description": "Proof of Value Creation Implementation",
        "total_supply": povc_engine.total_supply,
        "monthly_reward_pool": povc_engine.monthly_reward_pool,
        "endpoints": {
            "/health": "Service health",
            "/povc/calculate": "Calculate PoVC reward (POST)",
            "/povc/simulate": "Run simulation with sample data",
            "/wallet/generate": "Generate test wallet"
        }
    }

@app.get("/health")
async def health():
    return {
        "status": "healthy",
        "service": "nusa-ai-engine",
        "timestamp": datetime.now().isoformat()
    }

@app.post("/povc/calculate")
async def calculate_reward(request: UserRequest):
    """Calculate PoVC reward for a user"""
    try:
        result = povc_engine.calculate_reward(request.dict())
        return {"success": True, "data": result}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/povc/simulate")
async def simulate_distribution():
    """Run simulation with sample data"""
    import random
    
    sample_users = []
    for i in range(5):
        user = UserRequest(
            wallet_address=f"0xSample{i:03d}",
            wallet_balance=random.uniform(100, 50000),
            daily_activity=random.uniform(30, 300),
            contributions=random.randint(0, 100),
            community_score=random.uniform(0, 100),
            quality_score=random.uniform(0.3, 1.0),
            days_active=random.randint(1, 365)
        )
        sample_users.append(user)
    
    results = []
    total_rewards = 0
    
    for user in sample_users:
        reward = povc_engine.calculate_reward(user.dict())
        results.append(reward)
        total_rewards += reward["final_reward"]
    
    return {
        "simulation": "PoVC Monthly Distribution",
        "total_participants": len(sample_users),
        "total_rewards_distributed": round(total_rewards, 2),
        "average_reward": round(total_rewards / len(sample_users), 2),
        "individual_rewards": results,
        "timestamp": datetime.now().isoformat()
    }

@app.get("/wallet/generate")
async def generate_wallet():
    """Generate a test wallet (for demo only)"""
    import secrets
    
    private_key = secrets.token_hex(32)
    address = "0x" + hashlib.sha256(private_key.encode()).hexdigest()[:40]
    
    return {
        "wallet": {
            "address": address,
            "private_key": private_key,
            "warning": "FOR DEMONSTRATION ONLY - NOT FOR REAL USE"
        },
        "generated_at": datetime.now().isoformat()
    }

@app.get("/povc/anti-whale/{balance}")
async def check_whale_status(balance: float):
    """Check anti-whale status for a balance"""
    result = povc_engine.anti_whale_check(balance)
    return {
        "balance_check": result,
        "total_supply": povc_engine.total_supply,
        "percentage_of_supply": result["percentage"]
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
