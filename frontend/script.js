// frontend/script.js
class NUSADashboard {
    constructor() {
        this.endpoints = {
            l3: 'http://localhost:8000',
            l1: 'http://localhost:8545',
            l2: 'http://localhost:8081'
        };
        
        this.state = {
            services: {},
            wallet: null,
            simulation: null,
            transactions: [],
            stats: {},
            connectionStatus: 'checking'
        };
        
        this.init();
    }
    
    init() {
        this.loadState();
        this.checkAllServices();
        this.updateDashboard();
        this.setupEventListeners();
        this.startAutoRefresh();
        
        // Show welcome message
        this.showToast('NUSA Chain Dashboard loaded!', 'success');
    }
    
    loadState() {
        const saved = localStorage.getItem('nusa-dashboard-state');
        if (saved) {
            this.state = { ...this.state, ...JSON.parse(saved) };
        }
    }
    
    saveState() {
        localStorage.setItem('nusa-dashboard-state', JSON.stringify(this.state));
    }
    
    async checkAllServices() {
        const services = [
            { 
                name: 'L3 AI Engine', 
                url: `${this.endpoints.l3}/health`,
                icon: '🤖',
                description: 'AI Neutrality Engine'
            },
            { 
                name: 'L1 Blockchain', 
                url: `${this.endpoints.l1}/health`,
                icon: '⛓️',
                description: 'Core Blockchain Node'
            },
            { 
                name: 'L2 VM Layer', 
                url: `${this.endpoints.l2}`,
                icon: '⚡',
                description: 'WASM VM Runtime'
            },
            { 
                name: 'Redis Cache', 
                url: null,
                icon: '🗄️',
                description: 'In-memory Database',
                type: 'internal'
            },
            { 
                name: 'PostgreSQL', 
                url: null,
                icon: '💾',
                description: 'Main Database',
                type: 'internal'
            }
        ];
        
        const results = [];
        
        for (const service of services) {
            if (service.type === 'internal') {
                results.push({
                    ...service,
                    status: 'online',
                    timestamp: new Date().toISOString(),
                    data: { status: 'running_in_docker' }
                });
                continue;
            }
            
            try {
                const response = await fetch(service.url, { 
                    signal: AbortSignal.timeout(3000) 
                });
                const data = await response.json();
                
                results.push({
                    ...service,
                    status: 'online',
                    timestamp: new Date().toISOString(),
                    data: data
                });
            } catch (error) {
                results.push({
                    ...service,
                    status: 'offline',
                    timestamp: new Date().toISOString(),
                    error: error.message
                });
            }
        }
        
        this.state.services = results;
        this.updateServiceStatus();
        this.saveState();
    }
    
    updateServiceStatus() {
        const container = document.getElementById('serviceStatus');
        if (!container) return;
        
        container.innerHTML = this.state.services.map(service => `
            <div class="service ${service.status}">
                <div class="service-icon">${service.icon}</div>
                <div class="service-name">${service.name}</div>
                <div class="service-description">${service.description}</div>
                <div class="service-status status-${service.status}">
                    ${service.status.toUpperCase()}
                </div>
                ${service.data?.status ? 
                    `<div class="service-detail">${service.data.status}</div>` : ''}
            </div>
        `).join('');
        
        // Update connection status
        const onlineCount = this.state.services.filter(s => s.status === 'online').length;
        const totalCount = this.state.services.length;
        
        this.state.connectionStatus = onlineCount === totalCount ? 'healthy' :
                                     onlineCount > 0 ? 'degraded' : 'offline';
        
        this.updateConnectionStatus();
    }
    
    updateConnectionStatus() {
        const statusEl = document.getElementById('connectionStatus');
        if (!statusEl) return;
        
        const statusMap = {
            healthy: { text: 'All Systems Operational', class: 'status-online', icon: '✅' },
            degraded: { text: 'Partial Outage', class: 'status-warning', icon: '⚠️' },
            offline: { text: 'System Offline', class: 'status-offline', icon: '❌' }
        };
        
        const status = statusMap[this.state.connectionStatus];
        statusEl.innerHTML = `
            ${status.icon} ${status.text}
        `;
        statusEl.className = `status-badge ${status.class}`;
    }
    
    async runPoVCSimulation() {
        const loadingEl = document.getElementById('simLoading');
        const outputEl = document.getElementById('simOutput');
        
        this.showLoading(loadingEl);
        outputEl.innerHTML = '';
        
        try {
            const response = await fetch(`${this.endpoints.l3}/povc/simulate`);
            const data = await response.json();
            
            this.state.simulation = data;
            
            let html = `
                <div class="simulation-result">
                    <h4>💰 PoVC Simulation Results</h4>
                    <div class="stats-grid">
                        <div class="stat">
                            <div class="stat-value">${data.total_participants}</div>
                            <div class="stat-label">Participants</div>
                        </div>
                        <div class="stat">
                            <div class="stat-value">${data.total_rewards_distributed}</div>
                            <div class="stat-label">Total NUSA</div>
                        </div>
                        <div class="stat">
                            <div class="stat-value">${data.average_reward}</div>
                            <div class="stat-label">Average per User</div>
                        </div>
                    </div>
            `;
            
            if (data.individual_rewards && data.individual_rewards.length > 0) {
                html += `
                    <h5>Individual Rewards:</h5>
                    <div class="table-container">
                        <table class="table">
                            <thead>
                                <tr>
                                    <th>Wallet</th>
                                    <th>NVS Score</th>
                                    <th>Reward</th>
                                    <th>Status</th>
                                </tr>
                            </thead>
                            <tbody>
                                ${data.individual_rewards.map(reward => `
                                    <tr>
                                        <td><code>${reward.wallet.substring(0, 12)}...</code></td>
                                        <td>${reward.nvs_score.toFixed(4)}</td>
                                        <td><strong>${reward.final_reward} NUSA</strong></td>
                                        <td>
                                            <span class="badge ${
                                                reward.whale_check?.is_whale ? 'badge-danger' : 
                                                reward.whale_check?.warnings?.length > 0 ? 'badge-warning' : 
                                                'badge-success'
                                            }">
                                                ${reward.whale_check?.is_whale ? 'Whale' : 
                                                  reward.whale_check?.warnings?.length > 0 ? 'Warning' : 'Normal'}
                                            </span>
                                        </td>
                                    </tr>
                                `).join('')}
                            </tbody>
                        </table>
                    </div>
                `;
            }
            
            html += `</div>`;
            outputEl.innerHTML = html;
            
            this.showToast('Simulation completed successfully!', 'success');
            
        } catch (error) {
            outputEl.innerHTML = `
                <div class="error-message">
                    <h4>❌ Error</h4>
                    <p>${error.message}</p>
                    <p>Make sure the L3 AI Engine is running.</p>
                </div>
            `;
            this.showToast('Simulation failed!', 'error');
        } finally {
            this.hideLoading(loadingEl);
        }
    }
    
    async generateWallet() {
        const loadingEl = document.getElementById('walletLoading');
        const outputEl = document.getElementById('walletOutput');
        
        this.showLoading(loadingEl);
        outputEl.innerHTML = '';
        
        try {
            const response = await fetch(`${this.endpoints.l3}/wallet/generate`);
            const data = await response.json();
            
            this.state.wallet = data.wallet;
            
            const html = `
                <div class="wallet-card">
                    <h4>👛 New Wallet Generated</h4>
                    <div class="wallet-info">
                        <div class="info-row">
                            <span class="label">Address:</span>
                            <code class="address">${data.wallet.address}</code>
                            <button class="btn btn-sm" onclick="copyToClipboard('${data.wallet.address}')">
                                Copy
                            </button>
                        </div>
                        <div class="info-row">
                            <span class="label">Private Key:</span>
                            <code>${data.wallet.private_key.substring(0, 16)}...</code>
                            <span class="warning">⚠️ Keep this secret!</span>
                        </div>
                        <div class="info-row">
                            <span class="label">Generated:</span>
                            <span>${new Date(data.generated_at).toLocaleString()}</span>
                        </div>
                    </div>
                    <div class="wallet-actions">
                        <button class="btn btn-sm" onclick="saveWallet()">
                            💾 Save to Local Storage
                        </button>
                        <button class="btn btn-sm btn-outline" onclick="showQRCode('${data.wallet.address}')">
                            📱 Show QR Code
                        </button>
                    </div>
                    <div class="warning-box">
                        <strong>⚠️ IMPORTANT:</strong> ${data.wallet.warning}
                    </div>
                </div>
            `;
            
            outputEl.innerHTML = html;
            this.showToast('Wallet generated successfully!', 'success');
            
        } catch (error) {
            outputEl.innerHTML = `
                <div class="error-message">
                    <h4>❌ Error</h4>
                    <p>${error.message}</p>
                </div>
            `;
            this.showToast('Failed to generate wallet!', 'error');
        } finally {
            this.hideLoading(loadingEl);
        }
    }
    
    async checkWhaleStatus() {
        const balance = document.getElementById('balanceInput').value;
        const outputEl = document.getElementById('whaleOutput');
        
        if (!balance || isNaN(balance) || balance <= 0) {
            outputEl.innerHTML = `
                <div class="error-message">
                    Please enter a valid positive balance
                </div>
            `;
            return;
        }
        
        outputEl.innerHTML = '<div class="spinner"></div> Checking...';
        
        try {
            const response = await fetch(`${this.endpoints.l3}/povc/anti-whale/${balance}`);
            const data = await response.json();
            
            const check = data.balance_check;
            
            let statusClass = 'badge-success';
            let statusText = 'Normal';
            
            if (check.warnings?.includes('WHALE')) {
                statusClass = 'badge-danger';
                statusText = 'WHALE';
            } else if (check.warnings?.length > 0) {
                statusClass = 'badge-warning';
                statusText = 'Warning';
            }
            
            const html = `
                <div class="whale-check-result">
                    <h4>🐋 Whale Status Check</h4>
                    <div class="status-indicator ${statusClass.replace('badge-', '')}">
                        <span class="badge ${statusClass}">${statusText}</span>
                    </div>
                    
                    <div class="check-details">
                        <div class="detail-row">
                            <span class="label">Balance:</span>
                            <span class="value">${check.balance} NUSA</span>
                        </div>
                        <div class="detail-row">
                            <span class="label">% of Supply:</span>
                            <span class="value">${check.percentage}%</span>
                        </div>
                        <div class="detail-row">
                            <span class="label">Reward Multiplier:</span>
                            <span class="value">${check.reward_multiplier}x</span>
                        </div>
                        <div class="detail-row">
                            <span class="label">Transfer Fee:</span>
                            <span class="value">${check.transfer_fee}%</span>
                        </div>
                    </div>
                    
                    ${check.warnings?.length > 0 ? `
                        <div class="warnings">
                            <h5>⚠️ Warnings:</h5>
                            <ul>
                                ${check.warnings.map(warning => `<li>${warning}</li>`).join('')}
                            </ul>
                        </div>
                    ` : ''}
                    
                    <div class="recommendation">
                        <strong>💡 Recommendation:</strong>
                        ${check.percentage > 2 ? 
                            'Consider distributing your holdings to maintain rewards eligibility.' :
                          check.percentage > 0.5 ?
                            'Your rewards may be reduced. Consider maintaining balance below 0.5% of supply.' :
                            'Your balance is within healthy limits. Continue contributing!'}
                    </div>
                </div>
            `;
            
            outputEl.innerHTML = html;
            
        } catch (error) {
            outputEl.innerHTML = `
                <div class="error-message">
                    <h4>❌ Error</h4>
                    <p>${error.message}</p>
                </div>
            `;
        }
    }
    
    async testIntegration() {
        const loadingEl = document.getElementById('integrationLoading');
        const outputEl = document.getElementById('integrationOutput');
        
        this.showLoading(loadingEl);
        outputEl.innerHTML = '';
        
        try {
            const response = await fetch(`${this.endpoints.l1}/povc/test`);
            const data = await response.json();
            
            let html = `
                <div class="integration-result">
                    <h4>🔗 L1-L3 Integration Test</h4>
                    <div class="test-status ${data.success ? 'success' : 'error'}">
                        ${data.success ? '✅ SUCCESS' : '❌ FAILED'}
                    </div>
                    
                    <div class="test-details">
            `;
            
            if (data.success) {
                html += `
                    <div class="detail-row">
                        <span class="label">Status:</span>
                        <span class="value">${data.status || 'completed'}</span>
                    </div>
                    <div class="detail-row">
                        <span class="label">Block Created:</span>
                        <span class="value">#${data.block_created}</span>
                    </div>
                    <div class="detail-row">
                        <span class="label">Transaction Hash:</span>
                        <code class="value">${data.transaction?.hash || 'N/A'}</code>
                    </div>
                    <div class="detail-row">
                        <span class="label">AI Response:</span>
                        <span class="value">Received successfully</span>
                    </div>
                `;
            } else {
                html += `
                    <div class="detail-row">
                        <span class="label">Error:</span>
                        <span class="value error">${data.error || 'Unknown error'}</span>
                    </div>
                `;
            }
            
            html += `
                    </div>
                </div>
            `;
            
            outputEl.innerHTML = html;
            
            if (data.success) {
                this.showToast('Integration test passed!', 'success');
            } else {
                this.showToast('Integration test failed!', 'error');
            }
            
        } catch (error) {
            outputEl.innerHTML = `
                <div class="error-message">
                    <h4>❌ Error</h4>
                    <p>${error.message}</p>
                    <p>The L1 Blockchain Node may not be running.</p>
                </div>
            `;
            this.showToast('Integration test failed!', 'error');
        } finally {
            this.hideLoading(loadingEl);
        }
    }
    
    async getBlockchainInfo() {
        try {
            const response = await fetch(`${this.endpoints.l1}/`);
            const data = await response.json();
            return data;
        } catch (error) {
            return null;
        }
    }
    
    async getLatestTransactions() {
        // This would connect to a real transaction endpoint
        // For now, return mock data
        return [
            {
                hash: '0x123...abc',
                from: '0x456...def',
                to: '0x789...ghi',
                value: '1000 NUSA',
                timestamp: new Date().toISOString(),
                status: 'confirmed'
            }
        ];
    }
    
    updateDashboard() {
        this.updateStats();
        this.updateTransactionList();
        this.updateNetworkInfo();
    }
    
    async updateStats() {
        const statsEl = document.getElementById('stats');
        if (!statsEl) return;
        
        const blockchainInfo = await this.getBlockchainInfo();
        
        const stats = [
            {
                label: 'Total Supply',
                value: '25,000,000 NUSA',
                icon: '💰',
                change: '+0%',
                trend: 'stable'
            },
            {
                label: 'Active Wallets',
                value: blockchainInfo?.active_wallets || 'Loading...',
                icon: '👛',
                change: '+12',
                trend: 'up'
            },
            {
                label: 'Block Height',
                value: `#${blockchainInfo?.block_height || '0'}`,
                icon: '⛓️',
                change: '+1',
                trend: 'up'
            },
            {
                label: 'Monthly Rewards',
                value: '100,000 NUSA',
                icon: '🎁',
                change: 'Next: 7 days',
                trend: 'info'
            },
            {
                label: 'Avg. NVS Score',
                value: '0.65',
                icon: '📊',
                change: '+0.02',
                trend: 'up'
            },
            {
                label: 'Network Status',
                value: this.state.connectionStatus === 'healthy' ? 'Healthy' : 
                       this.state.connectionStatus === 'degraded' ? 'Degraded' : 'Offline',
                icon: '📡',
                change: '',
                trend: this.state.connectionStatus
            }
        ];
        
        statsEl.innerHTML = stats.map(stat => `
            <div class="stat-card">
                <div class="stat-icon">${stat.icon}</div>
                <div class="stat-value">${stat.value}</div>
                <div class="stat-label">${stat.label}</div>
                ${stat.change ? `
                    <div class="stat-change trend-${stat.trend}">
                        ${stat.change}
                    </div>
                ` : ''}
            </div>
        `).join('');
    }
    
    async updateTransactionList() {
        const txListEl = document.getElementById('transactionList');
        if (!txListEl) return;
        
        const transactions = await this.getLatestTransactions();
        
        if (transactions.length === 0) {
            txListEl.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">📭</div>
                    <div class="empty-text">No transactions yet</div>
                </div>
            `;
            return;
        }
        
        txListEl.innerHTML = `
            <table class="table">
                <thead>
                    <tr>
                        <th>Hash</th>
                        <th>From</th>
                        <th>To</th>
                        <th>Value</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    ${transactions.map(tx => `
                        <tr>
                            <td><code>${tx.hash.substring(0, 10)}...</code></td>
                            <td><code>${tx.from.substring(0, 8)}...</code></td>
                            <td><code>${tx.to.substring(0, 8)}...</code></td>
                            <td>${tx.value}</td>
                            <td>
                                <span class="badge badge-success">
                                    ${tx.status}
                                </span>
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;
    }
    
    updateNetworkInfo() {
        const infoEl = document.getElementById('networkInfo');
        if (!infoEl) return;
        
        const onlineServices = this.state.services.filter(s => s.status === 'online').length;
        const totalServices = this.state.services.length;
        
        infoEl.innerHTML = `
            <div class="network-status">
                <div class="status-summary">
                    <span class="status-text">Services: ${onlineServices}/${totalServices} online</span>
                    <div class="status-bar">
                        <div class="status-fill" style="width: ${(onlineServices/totalServices)*100}%"></div>
                    </div>
                </div>
                <div class="network-details">
                    <div class="detail">
                        <span class="label">Chain ID:</span>
                        <span class="value">2024</span>
                    </div>
                    <div class="detail">
                        <span class="label">Consensus:</span>
                        <span class="value">Proof of Value Creation</span>
                    </div>
                    <div class="detail">
                        <span class="label">Last Updated:</span>
                        <span class="value">${new Date().toLocaleTimeString()}</span>
                    </div>
                </div>
            </div>
        `;
    }
    
    setupEventListeners() {
        // Service refresh button
        const refreshBtn = document.getElementById('refreshServices');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.checkAllServices();
                this.showToast('Refreshing services...', 'info');
            });
        }
        
        // Auto-refresh toggle
        const autoRefreshToggle = document.getElementById('autoRefreshToggle');
        if (autoRefreshToggle) {
            autoRefreshToggle.addEventListener('change', (e) => {
                if (e.target.checked) {
                    this.startAutoRefresh();
                    this.showToast('Auto-refresh enabled (30s)', 'info');
                } else {
                    this.stopAutoRefresh();
                    this.showToast('Auto-refresh disabled', 'info');
                }
            });
        }
        
        // Clear outputs button
        const clearBtn = document.getElementById('clearOutputs');
        if (clearBtn) {
            clearBtn.addEventListener('click', () => {
                document.querySelectorAll('.output').forEach(el => {
                    el.innerHTML = '';
                });
                this.showToast('All outputs cleared', 'info');
            });
        }
        
        // Export data button
        const exportBtn = document.getElementById('exportData');
        if (exportBtn) {
            exportBtn.addEventListener('click', () => {
                this.exportData();
            });
        }
    }
    
    startAutoRefresh() {
        if (this.autoRefreshInterval) {
            clearInterval(this.autoRefreshInterval);
        }
        
        this.autoRefreshInterval = setInterval(() => {
            this.checkAllServices();
            this.updateDashboard();
        }, 30000); // 30 seconds
    }
    
    stopAutoRefresh() {
        if (this.autoRefreshInterval) {
            clearInterval(this.autoRefreshInterval);
            this.autoRefreshInterval = null;
        }
    }
    
    showLoading(element) {
        if (element) {
            element.classList.add('active');
        }
    }
    
    hideLoading(element) {
        if (element) {
            element.classList.remove('active');
        }
    }
    
    showToast(message, type = 'info') {
        // Create toast element
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.innerHTML = `
            <div class="toast-icon">${this.getToastIcon(type)}</div>
            <div class="toast-message">${message}</div>
            <button class="toast-close" onclick="this.parentElement.remove()">×</button>
        `;
        
        // Add to container
        const container = document.getElementById('toastContainer') || this.createToastContainer();
        container.appendChild(toast);
        
        // Auto remove after 5 seconds
        setTimeout(() => {
            if (toast.parentElement) {
                toast.remove();
            }
        }, 5000);
    }
    
    getToastIcon(type) {
        const icons = {
            success: '✅',
            error: '❌',
            warning: '⚠️',
            info: 'ℹ️'
        };
        return icons[type] || 'ℹ️';
    }
    
    createToastContainer() {
        const container = document.createElement('div');
        container.id = 'toastContainer';
        container.className = 'toast-container';
        document.body.appendChild(container);
        return container;
    }
    
    exportData() {
        const data = {
            timestamp: new Date().toISOString(),
            services: this.state.services,
            wallet: this.state.wallet,
            simulation: this.state.simulation,
            connectionStatus: this.state.connectionStatus
        };
        
        const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `nusa-dashboard-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        
        this.showToast('Data exported successfully!', 'success');
    }
}

// Utility functions
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        const dashboard = window.nusaDashboard;
        dashboard.showToast('Copied to clipboard!', 'success');
    });
}

function saveWallet() {
    const dashboard = window.nusaDashboard;
    if (dashboard.state.wallet) {
        localStorage.setItem('nusa-wallet', JSON.stringify(dashboard.state.wallet));
        dashboard.showToast('Wallet saved to local storage!', 'success');
    }
}

function showQRCode(address) {
    // This would generate a QR code for the address
    // For now, show a message
    const dashboard = window.nusaDashboard;
    dashboard.showToast(`QR Code for: ${address.substring(0, 16)}...`, 'info');
}

// Initialize when page loads
document.addEventListener('DOMContentLoaded', () => {
    window.nusaDashboard = new NUSADashboard();
    
    // Make functions available globally
    window.copyToClipboard = copyToClipboard;
    window.saveWallet = saveWallet;
    window.showQRCode = showQRCode;
});