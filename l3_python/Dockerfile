FROM python:3.11-slim

WORKDIR /app

# Install system dependencies hanya jika diperlukan
RUN apt-get update && apt-get install -y --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first for better caching
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY ai_engine/ ./ai_engine/

EXPOSE 8000

CMD ["uvicorn", "ai_engine.main:app", "--host", "0.0.0.0", "--port", "8000"]