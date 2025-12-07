from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Dict, List, Optional
import redis
import json
from datetime import datetime
from povc import ProofOfValueCreation

app = FastAPI(
    title="NUSA Chain AI Engine",
    version="0.2.0",
    description="AI Neutrality Engine for Proof of Value Creation",
    docs_url="/docs",
    redoc_url="/redoc"
)

# Initialize engines
povc_engine = ProofOfValueCreation()

# Redis connection
try:
    redis_client = redis.Redis(
        host='redis',
        port=6379,
        decode_responses=True,
        socket_connect_timeout=3
    )
    redis_client.ping()
    redis_connected = True
except:
    redis_connected = False
    redis_client = None

# Models
class UserData(BaseModel):
    wallet_address: str
    daily_active_minutes: float = 0
    contributions_count: int = 0
    community_interactions: int = 0
    days_active: int = 0
    wallet_balance: float = 0
    quality_score: float = 0.5
    metadata: Optional[Dict] = None

class BatchRequest(BaseModel):
    users: List[UserData]
    simulation_id: Optional[str] = None

@app.get("/")
async def root():
    return {
        "service": "NUSA Chain AI Engine",
        "version": "0.2.0",
        "description": "Proof of Value Creation (PoVC) Implementation",
        "total_supply": povc_engine.total_supply,
        "monthly_reward_pool": povc_engine.monthly_reward_pool,
        "redis_connected": redis_connected,
        "endpoints": {
            "/health": "Service health check",
            "/povc/calculate": "Calculate PoVC reward for single user (POST)",
            "/povc/batch": "Calculate rewards for multiple users (POST)",
            "/povc/anti-whale/{address}/{balance}": "Check whale status",
            "/povc/simulate": "Run economic simulation with sample data"
        }
    }

@app.get("/health")
async def health():
    return {
        "status": "healthy",
        "timestamp": datetime.now().isoformat(),
        "service": "nusa-ai-engine",
        "version": "0.2.0",
        "redis": "connected" if redis_connected else "disconnected"
    }

@app.post("/povc/calculate")
async def calculate_povc_reward(user_data: UserData):
    """Calculate PoVC reward for a single user"""
    try:
        result = povc_engine.calculate_monthly_reward(user_data.dict())
        
        # Store in Redis if available
        if redis_connected:
            key = f"povc:user:{user_data.wallet_address}:{datetime.now().strftime('%Y-%m')}"
            redis_client.setex(key, 2592000, json.dumps(result))  # 30 days TTL
        
        return {
            "success": True,
            "data": result,
            "calculated_at": datetime.now().isoformat()
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/povc/batch")
async def calculate_batch_povc(batch: BatchRequest):
    """Calculate PoVC rewards for multiple users"""
    try:
        users_dict = [user.dict() for user in batch.users]
        result = povc_engine.simulate_distribution(users_dict)
        
        # Store simulation in Redis
        if redis_connected and batch.simulation_id:
            key = f"povc:simulation:{batch.simulation_id}"
            redis_client.setex(key, 86400, json.dumps(result))  # 24 hours TTL
        
        return {
            "success": True,
            "simulation_id": batch.simulation_id or "no-id",
            "data": result,
            "calculated_at": datetime.now().isoformat()
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/povc/anti-whale/{wallet_address}/{balance}")
async def check_anti_whale(wallet_address: str, balance: float):
    """Check anti-whale status for a wallet"""
    try:
        result = povc_engine.anti_whale_mechanism(wallet_address, balance)
        return {
            "success": True,
            "data": result,
            "checked_at": datetime.now().isoformat()
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/povc/simulate")
async def simulate_distribution():
    """Run a simulation with sample data"""
    # Generate sample users
    import random
    sample_users = []
    
    for i in range(10):
        user = UserData(
            wallet_address=f"0xSampleUser{i:03d}",
            daily_active_minutes=random.uniform(30, 300),
            contributions_count=random.randint(1, 200),
            community_interactions=random.randint(0, 100),
            days_active=random.randint(1, 365),
            wallet_balance=random.uniform(100, 50000),
            quality_score=random.uniform(0.3, 1.0)
        )
        sample_users.append(user)
    
    # Calculate distribution
    users_dict = [user.dict() for user in sample_users]
    result = povc_engine.simulate_distribution(users_dict)
    
    return {
        "success": True,
        "note": "Sample simulation with 10 users",
        "data": result,
        "simulated_at": datetime.now().isoformat()
    }

# Wallet generation endpoint
@app.get("/wallet/generate")
async def generate_wallet():
    """Generate a new NUSA wallet"""
    import secrets
    import hashlib
    
    # Generate random private key
    private_key = secrets.token_hex(32)
    
    # Simple address generation (for demo)
    public_key = hashlib.sha256(private_key.encode()).hexdigest()
    address = "0x" + public_key[:40]
    
    # Generate mnemonic (simplified)
    words = ["nusa", "chain", "block", "crypto", "value", "creation", "ai", "neutrality"]
    mnemonic = " ".join([secrets.choice(words) for _ in range(12)])
    
    return {
        "success": True,
        "data": {
            "private_key": private_key,
            "address": address,
            "mnemonic": mnemonic,
            "generated_at": datetime.now().isoformat(),
            "warning": "This is for demo only! Use proper key generation for production."
        }
    }