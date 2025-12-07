import React, { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [services, setServices] = useState([])
  const [wallet, setWallet] = useState(null)
  const [simulation, setSimulation] = useState(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    checkServices()
  }, [])

  const checkServices = async () => {
    const servicesToCheck = [
      { name: 'L3 AI', url: 'http://localhost:8000/health' },
      { name: 'L1 Node', url: 'http://localhost:8545/health' }
    ]
    
    const updated = await Promise.all(
      servicesToCheck.map(async (service) => {
        try {
          const res = await fetch(service.url)
          return { ...service, status: 'online', data: await res.json() }
        } catch {
          return { ...service, status: 'offline' }
        }
      })
    )
    
    setServices(updated)
  }

  const generateWallet = async () => {
    setLoading(true)
    try {
      const res = await fetch('http://localhost:8000/wallet/generate')
      const data = await res.json()
      setWallet(data.wallet)
    } catch (error) {
      console.error('Error:', error)
    }
    setLoading(false)
  }

  const runSimulation = async () => {
    setLoading(true)
    try {
      const res = await fetch('http://localhost:8000/povc/simulate')
      const data = await res.json()
      setSimulation(data)
    } catch (error) {
      console.error('Error:', error)
    }
    setLoading(false)
  }

  return (
    <div className="app">
      <header>
        <h1>🌐 NUSA Chain Dashboard</h1>
      </header>

      <div className="dashboard">
        {/* Service Status */}
        <div className="card">
          <h3>Service Status</h3>
          {services.map(service => (
            <div key={service.name} className={`service ${service.status}`}>
              <span>{service.name}</span>
              <span className="status">{service.status}</span>
            </div>
          ))}
          <button onClick={checkServices}>Refresh</button>
        </div>

        {/* Wallet Generator */}
        <div className="card">
          <h3>Wallet Generator</h3>
          <button onClick={generateWallet} disabled={loading}>
            {loading ? 'Generating...' : 'Generate Wallet'}
          </button>
          {wallet && (
            <div className="output">
              <p>Address: {wallet.address}</p>
              <p>Private Key: {wallet.private_key.substring(0, 16)}...</p>
            </div>
          )}
        </div>

        {/* PoVC Simulation */}
        <div className="card">
          <h3>PoVC Simulation</h3>
          <button onClick={runSimulation} disabled={loading}>
            {loading ? 'Running...' : 'Run Simulation'}
          </button>
          {simulation && (
            <div className="output">
              <p>Total Rewards: {simulation.total_rewards_distributed} NUSA</p>
              <p>Participants: {simulation.total_participants}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default App