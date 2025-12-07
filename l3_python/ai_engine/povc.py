import numpy as np
from typing import Dict, List, Tuple
from datetime import datetime, timedelta
import json
import hashlib

class ProofOfValueCreation:
    def __init__(self, total_supply: int = 25_000_000):
        self.total_supply = total_supply
        self.monthly_reward_pool = 100_000  # NUSA per month
        
    def calculate_user_score(self, user_data: Dict) -> Dict:
        """Calculate NUSA Value Score (NVS) for a user"""
        
        # Activity metrics (0-1 scale)
        activity_score = user_data.get('daily_active_minutes', 0) / 240  # Max 4 hours/day
        contribution_score = min(user_data.get('contributions_count', 0) / 100, 1.0)
        community_score = min(user_data.get('community_interactions', 0) / 50, 1.0)
        
        # Quality metrics
        quality_multiplier = user_data.get('quality_score', 0.5)  # AI evaluated
        
        # Age bonus (long-term participants)
        days_active = user_data.get('days_active', 0)
        age_bonus = min(days_active / 365, 1.0) * 0.2
        
        # Calculate NVS
        nvs = (
            activity_score * 0.3 +
            contribution_score * 0.4 + 
            community_score * 0.2 +
            age_bonus
        ) * quality_multiplier
        
        # Cap at 1.0
        nvs = min(max(nvs, 0), 1.0)
        
        return {
            "nvs_score": round(nvs, 4),
            "breakdown": {
                "activity": round(activity_score, 3),
                "contribution": round(contribution_score, 3),
                "community": round(community_score, 3),
                "age_bonus": round(age_bonus, 3),
                "quality_multiplier": quality_multiplier
            }
        }
    
    def anti_whale_mechanism(self, wallet_address: str, current_balance: float) -> Dict:
        """Apply anti-whale rules"""
        
        balance_percentage = (current_balance / self.total_supply) * 100
        
        result = {
            "wallet": wallet_address,
            "balance": current_balance,
            "percentage_of_supply": round(balance_percentage, 4),
            "reward_multiplier": 1.0,
            "transfer_fee_percentage": 0,
            "warnings": []
        }
        
        # Tier 1: Warning zone (0.5% - 1%)
        if balance_percentage > 0.5:
            reduction = min((balance_percentage - 0.5) * 2, 50)  # Max 50% reduction
            result["reward_multiplier"] = 1 - (reduction / 100)
            result["warnings"].append("Reward reduction active")
            result["transfer_fee_percentage"] = 1
        
        # Tier 2: High concentration (1% - 2%)
        if balance_percentage > 1:
            reduction = 50 + min((balance_percentage - 1) * 25, 50)  # 50-100% reduction
            result["reward_multiplier"] = 1 - (reduction / 100)
            result["warnings"].append("High concentration penalty")
            result["transfer_fee_percentage"] = 3
        
        # Tier 3: Whale zone (>2%)
        if balance_percentage > 2:
            result["reward_multiplier"] = 0
            result["warnings"].append("WHALE: No rewards")
            result["transfer_fee_percentage"] = 10
        
        return result
    
    def calculate_monthly_reward(self, user_data: Dict) -> Dict:
        """Calculate monthly PoVC reward"""
        
        # Calculate NVS score
        score_result = self.calculate_user_score(user_data)
        nvs = score_result["nvs_score"]
        
        # Base reward
        base_reward = self.monthly_reward_pool * nvs
        
        # Apply anti-whale rules
        whale_check = self.anti_whale_mechanism(
            user_data.get("wallet_address", "unknown"),
            user_data.get("wallet_balance", 0)
        )
        
        # Final reward
        final_reward = base_reward * whale_check["reward_multiplier"]
        
        return {
            "wallet": user_data.get("wallet_address", "unknown"),
            "nvs_score": nvs,
            "base_reward": round(base_reward, 2),
            "final_reward": round(final_reward, 2),
            "whale_check": whale_check,
            "distribution_date": datetime.now().strftime("%Y-%m-%d"),
            "next_distribution": (datetime.now() + timedelta(days=30)).strftime("%Y-%m-%d")
        }
    
    def simulate_distribution(self, users_data: List[Dict]) -> Dict:
        """Simulate monthly distribution for multiple users"""
        
        results = []
        total_distributed = 0
        
        for user in users_data:
            reward = self.calculate_monthly_reward(user)
            results.append(reward)
            total_distributed += reward["final_reward"]
        
        # Wealth distribution metrics
        balances = [u.get("wallet_balance", 0) for u in users_data]
        
        if balances:
            gini_coefficient = self.calculate_gini(balances)
            top_10_percent = np.percentile(balances, 90)
            median_balance = np.median(balances)
        else:
            gini_coefficient = 0
            top_10_percent = 0
            median_balance = 0
        
        return {
            "simulation_date": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            "total_participants": len(users_data),
            "total_distributed": round(total_distributed, 2),
            "average_reward": round(total_distributed / len(users_data) if users_data else 0, 2),
            "wealth_distribution": {
                "gini_coefficient": round(gini_coefficient, 4),
                "top_10_percent_threshold": round(top_10_percent, 2),
                "median_balance": round(median_balance, 2)
            },
            "individual_rewards": results
        }
    
    def calculate_gini(self, x):
        """Calculate Gini coefficient for wealth distribution"""
        x = np.array(x)
        if len(x) == 0:
            return 0
        x = np.sort(x)
        n = len(x)
        index = np.arange(1, n + 1)
        return ((np.sum((2 * index - n - 1) * x)) / (n * np.sum(x)))