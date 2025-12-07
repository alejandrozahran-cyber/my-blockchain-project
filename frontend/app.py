from flask import Flask, render_template, jsonify
import requests
import threading
import time

app = Flask(__name__)

class NUSADashboard:
    def __init__(self):
        self.l3_url = "http://localhost:8000"
        self.l1_url = "http://localhost:8545"
        self.services = {}
    
    def check_services(self):
        services = {
            "L3 AI Engine": self.l3_url + "/health",
            "L1 Blockchain": self.l1_url + "/health"
        }
        
        for name, url in services.items():
            try:
                resp = requests.get(url, timeout=3)
                self.services[name] = {"status": "online", "data": resp.json()}
            except:
                self.services[name] = {"status": "offline"}

dashboard = NUSADashboard()

@app.route('/')
def index():
    return render_template('index.html')

@app.route('/api/services')
def get_services():
    dashboard.check_services()
    return jsonify(dashboard.services)

@app.route('/api/simulate')
def simulate():
    try:
        resp = requests.get(dashboard.l3_url + "/povc/simulate", timeout=10)
        return jsonify(resp.json())
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/api/wallet/new')
def new_wallet():
    try:
        resp = requests.get(dashboard.l3_url + "/wallet/generate", timeout=5)
        return jsonify(resp.json())
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    # Auto-refresh services every 10 seconds
    def refresh_services():
        while True:
            dashboard.check_services()
            time.sleep(10)
    
    thread = threading.Thread(target=refresh_services, daemon=True)
    thread.start()
    
    app.run(debug=True, port=5000)