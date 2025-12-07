import numpy as np
import matplotlib.pyplot as plt
from datetime import datetime, timedelta

class NUSAEconomicSimulator:
    def __init__(self):
        self.total_supply = 25_000_000
        self.monthly_reward = 100_000
        self.users = []
        
    def add_user(self, initial_balance, contribution_rate):
        self.users.append({
            'balance': initial_balance,
            'contribution': contribution_rate,
            'history': []
        })
    
    def simulate_month(self):
        total_contributions = sum(u['contribution'] for u in self.users)
        
        for user in self.users:
            # Calculate reward based on contribution
            if total_contributions > 0:
                reward_share = user['contribution'] / total_contributions
                reward = self.monthly_reward * reward_share
            else:
                reward = 0
            
            # Anti-whale check
            balance_percentage = (user['balance'] / self.total_supply) * 100
            if balance_percentage > 2:
                reward = 0
            elif balance_percentage > 0.5:
                reduction = min((balance_percentage - 0.5) * 2, 100)
                reward = reward * ((100 - reduction) / 100)
            
            # Update balance
            user['balance'] += reward
            user['history'].append(user['balance'])
        
        return [u['balance'] for u in self.users]
    
    def run_simulation(self, months=12):
        results = []
        for month in range(months):
            balances = self.simulate_month()
            results.append(balances)
            print(f"Month {month+1}: {balances}")
        return results

# Quick simulation
if __name__ == "__main__":
    sim = NUSAEconomicSimulator()
    
    # Add sample users
    sim.add_user(1000, 0.3)  # Small contributor
    sim.add_user(10000, 0.5) # Medium contributor
    sim.add_user(50000, 0.2) # Large balance (potential whale)
    
    results = sim.run_simulation(12)
    
    # Plot results
    plt.figure(figsize=(10, 6))
    for i, user_balances in enumerate(zip(*results)):
        plt.plot(range(1, 13), user_balances, label=f'User {i+1}')
    
    plt.title('NUSA Chain Wealth Distribution Simulation (12 Months)')
    plt.xlabel('Month')
    plt.ylabel('Balance (NUSA)')
    plt.legend()
    plt.grid(True)
    plt.savefig('simulation.png')
    print("✅ Simulation saved to simulation.png")