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

### Run subscriber
<h2><a id="user-content-whats-tron" class="anchor" aria-hidden="true" href="#whats-tron"><svg class="octicon octicon-link" viewBox="0 0 16 16" version="1.1" width="16" height="16" aria-hidden="true"><path fill-rule="evenodd" d="M4 9h1v1H4c-1.5 0-3-1.69-3-3.5S2.55 3 4 3h4c1.45 0 3 1.69 3 3.5 0 1.41-.91 2.72-2 3.25V8.59c.58-.45 1-1.27 1-2.09C10 5.22 8.98 4 8 4H4c-.98 0-2 1.22-2 2.5S3 9 4 9zm9-3h-1v1h1c1 0 2 1.22 2 2.5S13.98 12 13 12H9c-.98 0-2-1.22-2-2.5 0-.83.42-1.64 1-2.09V6.25c-1.09.53-2 1.84-2 3.25C6 11.31 7.55 13 9 13h4c1.45 0 3-1.69 3-3.5S14.5 6 13 6z"></path></svg></a>WS Subscriber</h2>
<p>Ws Subscriber has three endpoints for subscribing for specific events from chain</p>
<p>`newHeads`, `logs` and `newPendingTransactions` provide subscription endpoints for pulling new finalized block headers, neon pending transactions and specific solidity contract event logs </p>
<p>Internally Subscriber uses rpc endpoint for pulling data from chain and connects to proxy Mempool for checking the list of pending transactions for subscribers</p>

<div class="highlight highlight-source-shell"><pre>
SOLANA_RPC_ENDPOINT=https://api.devnet.solana.com
</pre></div>
<div class="highlight highlight-source-shell"><pre>
go run cmd/neon-wssubscriber/main.go
</pre></div>

<details>
<summary>Correct output</summary>
<div class="highlight highlight-source-shell"><pre>
{"level":"info","message":"Metrics server inicialization has been skipped"}
{"level":"info","message":"Service wssubscriber version  started"}
{"level":"info","message":"block pulling from rpc started ... "}
{"level":"info","message":"newHeads broadcaster sources registered"}
{"level":"info","message":"pending transaction pulling from mempool started ... "}
{"level":"info","message":"pendingTransaction broadcaster sources registered"}
{"level":"info","message":"logs pulling from blocks started ... "}
{"level":"info","message":"newLogs broadcaster sources registered"}
{"level":"info","message":"logParser: latest processed block slot signature 0"}
{"level":"info","message":"Service wssubscriber has been stopped"}
...
</pre></div>
</details>

### Environment setup

```bash
# ENV
NEON_SERVICE_ENV = development / production

# Logs
NS_LOG_LEVEL = info
NS_LOG_USE_FILE = false
NS_LOG_PATH = logs


# DATABASES


## WSSUBSCRIBER
SOLANA_RPC_ENDPOINT=https://api.devnet.solana.com
NEON_WEBSOCKET_PORT=9090


## INDEXER
NS_DB_INDEXER_HOSTNAME = "localhost"
NS_DB_INDEXER_POST = 5432
NS_DB_INDEXER_SSLMODE = disable
NS_DB_INDEXER_USERNAME = "neon-indexer"
NS_DB_INDEXER_PASSWORD = ""
NS_DB_INDEXER_DATABASE = "neon-indexer"

# SUBSCRIBER SERVICE
NS_CFG_SUBSCRIBER_INTERVAL = 5 # 5 seconds

# METRICS

## PROXY
NS_METRICS_PROXY_LISTEN_ADDRESS = "127.0.0.1"
NS_METRICS_PROXY_LISTEN_PORT = 20501
NS_METRICS_PROXY_INTERVAL = 5 # 5 seconds

## INDEXER
NS_METRICS_INDEXER_LISTEN_ADDRESS = "127.0.0.1"
NS_METRICS_INDEXER_LISTEN_PORT = 20502
NS_METRICS_INDEXER_INTERVAL = 5 # 5 seconds

## SUBSCRIBER
NS_METRICS_SUBSCRIBER_LISTEN_ADDRESS = "127.0.0.1"
NS_METRICS_SUBSCRIBER_LISTEN_PORT = 20503
NS_METRICS_SUBSCRIBER_INTERVAL = 5 # 5 seconds

## MEMPOOL
NS_METRICS_MEMPOOL_LISTEN_ADDRESS = "127.0.0.1"
NS_METRICS_MEMPOOL_LISTEN_PORT = 20504
NS_METRICS_MEMPOOL_INTERVAL = 5 # 5 seconds

```
