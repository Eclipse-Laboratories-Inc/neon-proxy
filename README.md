# neon-proxy

## Proxy

### Build project

```bash
make build
```

### Run proxy

```bash
make run-proxy
```

### Run indexer

```bash
make run-indexer
```

### Environment setup

```bash
# ENV
NEON_SERVICE_ENV = development / production

# Logs
NEON_SERVICE_LOG_LEVEL = info
NEON_SERVICE_LOG_USE_FILE = false
NEON_SERVICE_LOG_PATH = logs

# DATABASES

## INDEXER
NS_DB_INDEXER_HOSTNAME = "localhost"
NS_DB_INDEXER_POST = 5432
NS_DB_INDEXER_SSLMODE = disable
NS_DB_INDEXER_USERNAME = "neon-indexer"
NS_DB_INDEXER_PASSWORD = ""
NS_DB_INDEXER_DATABASE = "neon-indexer"

# SUBSCRIBER SERVICE
NEON_SUBSCRIBER_INTERVAL = 5 # 5 seconds
```
