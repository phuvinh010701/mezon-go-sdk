# mezon-go-sdk

The official Go SDK for the [Mezon](https://mezon.ai) platform.

## Installation

```bash
go get github.com/phuvinh010701/mezon-go-sdk
```

## Package Structure

```
mezon-go-sdk/
├── client/          # Top-level Client struct and functional options
├── auth/            # Authentication (API key, custom Authenticator)
├── types/           # Shared request/response types
├── errors/          # SDK error types and sentinel errors
├── transport/       # HTTP transport helpers
└── internal/
    └── httpclient/  # Internal HTTP execution layer (not exported)
```

## Quick Start

```go
import "github.com/phuvinh010701/mezon-go-sdk/client"

c, err := client.New(
    client.WithAPIKey("your-api-key"),
)
if err != nil {
    log.Fatal(err)
}
```

## Configuration Options

| Option | Description |
|---|---|
| `WithAPIKey(key)` | Set bearer-token API key |
| `WithBaseURL(url)` | Override the API base URL |
| `WithHTTPClient(hc)` | Inject a custom `*http.Client` |
| `WithLogger(l)` | Set a custom `*slog.Logger` |
| `WithAuthenticator(a)` | Set a custom `auth.Authenticator` |

## Requirements

- Go 1.21+
