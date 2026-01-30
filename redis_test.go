package go_redis_wrapper

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redsync/redsync/v4"
	"github.com/redis/go-redis/v9"
)

func TestNewClusterClient(t *testing.T) {
	opts := &redis.ClusterOptions{
		Addrs: []string{"localhost:6379"},
	}

	client := NewClusterClient(opts)

	if client == nil {
		t.Error("Expected non-nil client")
	}

	if client.Type != ClusterClientType {
		t.Errorf("Expected client type %d, got %d", ClusterClientType, client.Type)
	}

	if client.rs == nil {
		t.Error("Expected redsync instance to be initialized")
	}
}

func TestNewClient(t *testing.T) {
	opts := &redis.Options{
		Addr: "localhost:6379",
	}

	client := NewClient(opts)

	if client == nil {
		t.Error("Expected non-nil client")
	}

	if client.Type != SimpleClientType {
		t.Errorf("Expected client type %d, got %d", SimpleClientType, client.Type)
	}

	if client.rs == nil {
		t.Error("Expected redsync instance to be initialized")
	}
}

func TestStartMiniRedis(t *testing.T) {
	client, err := StartMiniRedis()
	if err != nil {
		t.Fatalf("Failed to start MiniRedis: %v", err)
	}
	defer client.Close()

	if client == nil {
		t.Error("Expected non-nil client")
	}

	if client.Type != TestClientType {
		t.Errorf("Expected client type %d, got %d", TestClientType, client.Type)
	}

	// Test that we can ping the client
	err = client.Ping(context.Background())
	if err != nil {
		t.Errorf("Failed to ping MiniRedis: %v", err)
	}
}

func TestClientAdapter(t *testing.T) {
	// Start a mini redis instance for testing
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	uc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer uc.Close()

	tests := []struct {
		name        string
		clientType  uint8
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "simple client type",
			clientType: SimpleClientType,
			wantErr:    false,
		},
		{
			name:       "cluster client type",
			clientType: ClusterClientType,
			wantErr:    false,
		},
		{
			name:       "test client type",
			clientType: TestClientType,
			wantErr:    false,
		},
		{
			name:        "invalid client type",
			clientType:  empty,
			wantErr:     true,
			expectedErr: ErrInvalidClientType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := ClientAdapter(uc, tt.clientType)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if client == nil {
				t.Error("Expected non-nil client")
			}

			if client.Type != tt.clientType {
				t.Errorf("Expected client type %d, got %d", tt.clientType, client.Type)
			}
		})
	}
}

func TestCastToRedisCluster(t *testing.T) {
	tests := []struct {
		name        string
		client      redis.UniversalClient
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "valid cluster client",
			client:  redis.NewClusterClient(&redis.ClusterOptions{Addrs: []string{"localhost:6379"}}),
			wantErr: false,
		},
		{
			name:        "simple client (should fail)",
			client:      redis.NewClient(&redis.Options{Addr: "localhost:6379"}),
			wantErr:     true,
			expectedErr: ErrCastToClusterClient,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.client.Close()

			cluster, err := CastToRedisCluster(tt.client)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				if cluster != nil {
					t.Error("Expected nil cluster client")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if cluster == nil {
				t.Error("Expected non-nil cluster client")
			}
		})
	}
}

func TestSimplePing(t *testing.T) {
	// Start a mini redis instance for testing
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	ctx := context.Background()
	err = SimplePing(ctx, client)
	if err != nil {
		t.Errorf("SimplePing failed: %v", err)
	}
}

func TestClient_Ping(t *testing.T) {
	// Test with MiniRedis
	client, err := StartMiniRedis()
	if err != nil {
		t.Fatalf("Failed to start MiniRedis: %v", err)
	}
	defer client.Close()

	tests := []struct {
		name        string
		clientType  uint8
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "simple client",
			clientType: SimpleClientType,
			wantErr:    false,
		},
		{
			name:       "test client",
			clientType: TestClientType,
			wantErr:    false,
		},
		{
			name:        "invalid client type",
			clientType:  empty,
			wantErr:     true,
			expectedErr: ErrPingNotImplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Override client type for testing
			client.Type = tt.clientType

			err := client.Ping(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestClient_Status(t *testing.T) {
	client, err := StartMiniRedis()
	if err != nil {
		t.Fatalf("Failed to start MiniRedis: %v", err)
	}
	defer client.Close()

	status, err := client.Status()
	if err != nil {
		t.Errorf("Status check failed: %v", err)
	}

	expected := "ok"
	if status != expected {
		t.Errorf("Expected status %v, got %v", expected, status)
	}
}

func TestClient_SlotsCount(t *testing.T) {
	client, err := StartMiniRedis()
	if err != nil {
		t.Fatalf("Failed to start MiniRedis: %v", err)
	}
	defer client.Close()

	tests := []struct {
		name          string
		clientType    uint8
		wantErr       bool
		expectedErr   error
		expectedCount int
	}{
		{
			name:          "simple client",
			clientType:    SimpleClientType,
			wantErr:       false,
			expectedCount: 0,
		},
		{
			name:          "test client",
			clientType:    TestClientType,
			wantErr:       false,
			expectedCount: 0,
		},
		{
			name:        "invalid client type",
			clientType:  empty,
			wantErr:     true,
			expectedErr: ErrSlotsCountNotImplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Override client type for testing
			client.Type = tt.clientType

			count, err := client.SlotsCount(context.Background())

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
				return
			} else if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if count != tt.expectedCount {
				t.Errorf("Expected count %d, got %d", tt.expectedCount, count)
			}
		})
	}
}

func TestClient_Lock(t *testing.T) {
	client, err := StartMiniRedis()
	if err != nil {
		t.Fatalf("Failed to start MiniRedis: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-lock-key"

	mutex, err := client.Lock(ctx, key)
	if err != nil {
		t.Errorf("Lock failed: %v", err)
		return
	}

	if mutex == nil {
		t.Error("Expected non-nil mutex")
		return
	}

	// Clean up
	_, unlockErr := mutex.UnlockContext(ctx)
	if unlockErr != nil {
		t.Errorf("Failed to unlock: %v", unlockErr)
	}
}

func TestClient_Lock_WithOptions(t *testing.T) {
	client, err := StartMiniRedis()
	if err != nil {
		t.Fatalf("Failed to start MiniRedis: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-lock-key-with-options"

	// Test with custom options
	options := []redsync.Option{
		redsync.WithExpiry(5 * time.Second),
		redsync.WithTries(3),
	}

	mutex, err := client.Lock(ctx, key, options...)
	if err != nil {
		t.Errorf("Lock with options failed: %v", err)
		return
	}

	if mutex == nil {
		t.Error("Expected non-nil mutex")
		return
	}

	// Clean up
	_, unlockErr := mutex.UnlockContext(ctx)
	if unlockErr != nil {
		t.Errorf("Failed to unlock: %v", unlockErr)
	}
}

func TestClient_LockKey(t *testing.T) {
	client, err := StartMiniRedis()
	if err != nil {
		t.Fatalf("Failed to start MiniRedis: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-lockkey"

	unlock, err := client.LockKey(ctx, key)
	if err != nil {
		t.Errorf("LockKey failed: %v", err)
		return
	}

	if unlock == nil {
		t.Error("Expected non-nil unlock function")
		return
	}

	// Test unlocking
	err = unlock(ctx)
	if err != nil {
		t.Errorf("Unlock failed: %v", err)
	}
}

func TestClient_LockKey_WithOptions(t *testing.T) {
	client, err := StartMiniRedis()
	if err != nil {
		t.Fatalf("Failed to start MiniRedis: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-lockkey-with-options"

	// Test with custom options
	options := []redsync.Option{
		redsync.WithExpiry(5 * time.Second),
		redsync.WithTries(3),
	}

	unlock, err := client.LockKey(ctx, key, options...)
	if err != nil {
		t.Errorf("LockKey with options failed: %v", err)
		return
	}

	if unlock == nil {
		t.Error("Expected non-nil unlock function")
		return
	}

	// Test unlocking
	err = unlock(ctx)
	if err != nil {
		t.Errorf("Unlock failed: %v", err)
	}
}

// Benchmark tests
func BenchmarkClient_Ping(b *testing.B) {
	client, err := StartMiniRedis()
	if err != nil {
		b.Fatalf("Failed to start MiniRedis: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Ping(ctx)
	}
}

func BenchmarkClient_Lock(b *testing.B) {
	client, err := StartMiniRedis()
	if err != nil {
		b.Fatalf("Failed to start MiniRedis: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "bench-lock-" + strconv.Itoa(i)

		mutex, err := client.Lock(ctx, key)
		if err != nil {
			b.Errorf("Lock failed: %v", err)
			continue
		}
		_, _ = mutex.UnlockContext(ctx)
	}
}
