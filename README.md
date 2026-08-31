# song (Magic Link Authentication System)

**MIT License © Azzurro Technology Inc.**

## Overview

The song project implements a magic link-based passwordless security system. This approach eliminates traditional password vulnerabilities by using unique, time-limited links for authentication. song also serves as a static data server capable of handling both public and private data with advanced security features.

## Installation

### Prerequisites
- Go 1.20+
- SQLite database (optional)
- **github.com/gin-gonic/gin** v1.9.1 - HTTP server framework

### Installation Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/azzurro-tech/song.git
   cd song
   ```

2. Install Go dependencies:
   ```bash
   go mod download
   ```

3. Start the song authentication server:
   ```bash
   cd azzurrotech/song
   ./song --port 8083
   ```

4. Access the song web interface:
   ```
   http://localhost:8083
   http://localhost:8083/song/config
   http://localhost:8083/song/admin
   ```

## Usage (Standalone)

### Basic Operations

**Authentication Management**
```bash
# Health check
curl http://localhost:8083/health

# Generate magic link
curl -X POST http://localhost:8083/api/auth/generate \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user@example.com","device_info":{"device_type":"mobile","device_id":"device-123","user_agent":"Mobile App"}}'

# Validate magic link
curl -X POST http://localhost:8083/api/auth/validate \
  -H "Content-Type: application/json" \
  -d '{"link":"https://auth.example.com/validate?token=abc123","user_id":"user@example.com","device_info":{"device_type":"mobile","ip_address":"192.168.1.100"}}'

# Revoke magic link
curl -X POST http://localhost:8083/api/auth/revoke \
  -H "Content-Type: application/json" \
  -d '{"link":"https://auth.example.com/validate?token=abc123","user_id":"user@example.com"}'
```

**Static Data Management**
```bash
# Serve public data
curl http://localhost:8083/static/public/data.json

# Serve private data (with authentication)
curl -H "Authorization: Bearer magic-link-token" http://localhost:8083/static/private/data.json
```

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Main song status page |
| `/song` | GET | HTML config viewer |
| `/song/config` | GET | View configuration |
| `/song/config` | POST | Update configuration |
| `/song/admin` | GET | HTML admin panel |
| `/song/admin` | POST | Update admin settings |
| `/health` | GET | Health check |
| `/api/auth/generate` | POST | Generate magic link |
| `/api/auth/validate` | POST | Validate magic link |
| `/api/auth/revoke` | POST | Revoke magic link |
| `/static/*` | GET | Serve static files |

## Integration with ATP

### Service Registration

The song project registers with ATP as an authentication service that provides passwordless magic link authentication:

```go
// Example song service registration
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    
    // Health check endpoint
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy"})
    })
    
    // Magic link generation API
    auth := r.Group("/api/auth")
    {
        auth.POST("/generate", generateMagicLink)
        auth.POST("/validate", validateMagicLink)
        auth.POST("/revoke", revokeMagicLink)
    }
    
    // Static data API
    static := r.Group("/static")
    {
        static.GET("/*path", serveStaticFile)
    }
    
    // Service registration with ATP
    r.POST("/register", func(c *gin.Context) {
        config := map[string]interface{}{
            "name": "song",
            "endpoint": "http://localhost:8083",
            "health": "/health",
            "auth_endpoint": "/api/auth",
            "static_endpoint": "/static",
            "config_endpoint": "/api/song/config",
            "admin_endpoint": "/api/song/admin"
        }
        
        response, err := registerWithATP(config)
        if err != nil {
            c.JSON(500, gin.H{"error": "registration failed"})
            return
        }
        
        c.JSON(200, response)
    })
    
    r.Run(":8083")
}
```

### Authentication Integration

The song project integrates with ATP for centralized authentication management:

```yaml
# atp/config/integrations.yaml
integrations:
  azzurrotech:
    song:
      health_check: /health
      auth_endpoint: /api/auth
      static_endpoint: /static
      config_endpoint: /api/song/config
      admin_endpoint: /api/song/admin
      auth_required: true
      token_expiry_hours: 24
```

### Authentication Pipeline

1. **Magic Link Generation**: Generate unique, time-limited magic links
2. **Link Distribution**: Send magic links to users via email or other channels
3. **Authentication**: Users click links to authenticate
4. **Token Validation**: Validate magic links and tokens
5. **Session Management**: Manage authenticated sessions
6. **Integration**: Integrate with ATP authentication framework

## Development Setup

### Local Development

```bash
# Start song server
cd azzurrotech/song
./song --port 8083

# Or with Go run
cd azzurrotech/song
go run .
```

### Testing

```bash
# Run all tests
cd azzurrotech/song
go test ./...

# Run specific test packages
cd azzurrotech/song
go test ./pkg/auth/...
go test ./internal/...

# Run integration tests
cd azzurrotech/song
go test ./integration/...

# Test API endpoints
curl http://localhost:8083/health
curl -X POST http://localhost:8083/api/auth/generate -d '{"user_id":"test@example.com"}'
```

### Building

```bash
# Build for production
cd azzurrotech/song
go build -o song ./cmd

# Build with specific options
cd azzurrotech/song
go build -ldflags="-port=8083" -o song ./cmd

# Build with SQLite configuration
cd azzurrotech/song
DB_PATH=./data/song.db ./song --port 8083
```

## Performance Optimization

### Magic Link Generation

- **Efficient Link Generation**: Optimized for high-volume link generation
- **Caching**: Cache frequently used tokens and data
- **Connection Pooling**: Efficient database connection management
- **Compression**: Compress sensitive data transfers
- **Load Balancing**: Supports horizontal scaling

### Token Management

```go
// Token management optimization
var tokenCache = make(map[string]*Token)

func generateMagicLink(userID string, deviceInfo DeviceInfo) (*MagicLink, error) {
    // Generate secure token
    token := generateSecureToken()
    
    // Cache token for validation
    tokenCache[token] = &Token{
        UserID:     userID,
        DeviceInfo: deviceInfo,
        CreatedAt:  time.Now(),
        ExpiresAt:  time.Now().Add(24 * time.Hour),
        Used:       false,
    }
    
    // Generate magic link
    magicLink := &MagicLink{
        Token:     token,
        UserID:    userID,
        ExpiresAt: time.Now().Add(24 * time.Hour),
        URL:       fmt.Sprintf("https://auth.example.com/validate?token=%s", token),
    }
    
    return magicLink, nil
}
```

## Monitoring

### Health Monitoring

```bash
# song health check
curl http://localhost:8083/health

# Magic link generation
curl -X POST http://localhost:8083/api/auth/generate -d '{"user_id":"test@example.com"}'

# Magic link validation
curl -X POST http://localhost:8083/api/auth/validate -d '{"link":"https://auth.example.com/validate?token=test","user_id":"test@example.com"}'

# Configuration health
curl http://localhost:8083/song/config
```

### Metrics Collection

song collects and reports:

- **Authentication Attempts**: Magic link generation and validation attempts
- **Success Rates**: Authentication success/failure rates
- **Device Tracking**: Device information and authentication patterns
- **Security Events**: Security events and threats detected
- **Performance Metrics**: Authentication processing performance
- **Error Rates**: Authentication error tracking

## Security Features

### song Security

- **Passwordless Authentication**: No passwords stored on servers
- **Time-Limited Links**: Magic links expire after a set time period
- **Secure Link Generation**: Cryptographically secure link generation
- **Single-Use Authentication**: Each link can only be used once
- **Device Binding**: Links can be restricted to specific devices
- **JWT Tokens**: Secure token-based authentication
- **HTTPS Enforcement**: Enforces secure communication
- **Input Validation**: Validates all input to prevent injection attacks
- **Rate Limiting**: Prevents abuse of authentication endpoints

### Authentication Security

song provides secure authentication handling:

- **Passwordless Login**: Secure passwordless login experience
- **Enhanced Security**: Reduced attack surface and improved security
- **User Convenience**: Simple and intuitive authentication
- **Device Management**: Comprehensive device management
- **API Integration**: RESTful API for authentication services
- **Monitoring**: Comprehensive authentication monitoring
- **Configuration**: Flexible configuration options

## Troubleshooting

### Common Issues

1. **Magic Link Not Sent**
   ```bash
   # Check song logs
   $ tail -f song.log
   
   # Test email configuration
   $ grep -r "email" song.log
   
   # Check song health
   $ curl http://localhost:8083/health
   ```

2. **Magic Link Expired**
   ```bash
   # Check link expiry logic
   $ grep -r "expiry" song.log
   
   # Test magic link generation
   $ curl -X POST http://localhost:8083/api/auth/generate -d '{"user_id":"test@example.com"}'
   
   # Check song configuration
   $ curl http://localhost:8083/song/config
   ```

3. **Validation Errors**
   ```bash
   # Check token validation
   $ grep -r "validate" song.log
   
   # Test magic link validation
   $ curl -X POST http://localhost:8083/api/auth/validate -d '{"link":"https://auth.example.com/validate?token=test","user_id":"test@example.com"}'
   
   # Check token cache
   $ grep -r "cache" song.log
   ```

### Debugging Commands

```bash
# Enable debug logging
export SONG_LOG_LEVEL=debug

# Check song logs
$ tail -f song.log

# Monitor system resources
$ top
$ free -h

# Test authentication endpoints
$ curl http://localhost:8083/health
$ curl -X POST http://localhost:8083/api/auth/generate -d '{"user_id":"test@example.com"}'

# Check song configuration
$ curl http://localhost:8083/song/config
```

## API Specifications

### High Maturity API (REST-based)

```http
POST /api/auth/generate
Content-Type: application/json

{
  "user_id": "user@example.com",
  "device_info": {
    "device_type": "mobile",
    "device_id": "device-123",
    "user_agent": "Mobile App"
  }
}

HTTP/1.1 200 OK
{
  "magic_link": "https://auth.example.com/token/abc123",
  "expires_at": "2024-01-01T12:00:00Z",
  "token_id": "token_abc123"
}
```

### song APIs

```http
POST /api/auth/generate - Generate magic link
POST /api/auth/validate - Validate magic link
POST /api/auth/revoke - Revoke magic link
GET /health - Health check
GET /static/* - Serve static files
```

### Authentication API

```http
POST /api/auth/generate - Generate magic link
POST /api/auth/validate - Validate magic link
POST /api/auth/revoke - Revoke magic link
```

### song-specific Endpoints

```http
GET /api/song/config - View song configuration
POST /api/song/config - Update song configuration
GET /api/song/admin - View admin information
POST /api/song/admin - Modify admin settings
```

## Testing

### Unit Tests

```go
// Test magic link generation
testMagicLinkGeneration(t *testing.T)

// Test token validation
testTokenValidation(t *testing.T)

// Test API endpoints
testAPIEndpoints(t *testing.T)
```

### Integration Tests

```bash
# Start song server
$ ./song --port 8083 &

# Run integration tests
$ curl http://localhost:8083/health
$ curl -X POST http://localhost:8083/api/auth/generate -d '{"user_id":"test@example.com"}'
```

## Performance Considerations

- **Link Generation**: Monitor for high-volume link generation
- **Token Storage**: Efficient token storage and management
- **Database Operations**: Optimize database queries
- **Network I/O**: Cache frequently used tokens
- **Memory Usage**: Monitor token storage and caching

## Future Enhancements

- **Multi-Factor Authentication**: Add MFA support
- **Biometric Authentication**: Integrate biometric authentication
- **Advanced Analytics**: Add authentication analytics
- **API Gateway**: Add API management capabilities
- **Multi-tenancy**: Support multiple tenants and organizations

## Conclusion

The song project provides a secure, passwordless authentication system using magic links. It offers enhanced security features while maintaining a user-friendly experience. The system integrates seamlessly with the ATP platform and provides comprehensive authentication and authorization capabilities.

Key benefits:

- **Passwordless Authentication**: Secure passwordless login experience
- **Enhanced Security**: Reduced attack surface and improved security
- **User Convenience**: Simple and intuitive authentication
- **Device Management**: Comprehensive device management
- **API Integration**: RESTful API for authentication services
- **Monitoring**: Comprehensive authentication monitoring
- **Configuration**: Flexible configuration options

This authentication system is production-ready and can be easily integrated into web applications and services with robust passwordless authentication capabilities.

---

*Document Version: 1.0*
*Created: 2026-08-25*
*Last Updated: 2026-08-25*
*Status: Production Ready*

**License:** MIT License © Azzurro Technology Inc.