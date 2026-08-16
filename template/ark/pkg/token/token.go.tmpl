package token

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/morehao/go-ark-template/pkg/dbclient"
	"github.com/morehao/golib/glog"
)

const (
	TokenBlacklistKeyPrefix        = "token:blacklist:"
	RefreshTokenBlacklistKeyPrefix = "refreshToken:blacklist:"
	TokenExpireDuration            = 24 * time.Hour
	RefreshTokenExpireDuration     = 7 * 24 * time.Hour
)

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}

func AddTokenToBlacklist(ctx context.Context, token string, expireDuration time.Duration) error {
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return nil
	}
	key := TokenBlacklistKeyPrefix + HashToken(token)
	if err := dbclient.RedisCli.Set(ctx, key, "1", expireDuration).Err(); err != nil {
		glog.Errorf(ctx, "[token.AddTokenToBlacklist] Redis Set token fail, err:%v", err)
		return err
	}
	return nil
}

func AddRefreshTokenToBlacklist(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return nil
	}
	key := RefreshTokenBlacklistKeyPrefix + HashToken(token)
	if err := dbclient.RedisCli.Set(ctx, key, "1", RefreshTokenExpireDuration).Err(); err != nil {
		glog.Errorf(ctx, "[token.AddRefreshTokenToBlacklist] Redis Set refresh token fail, err:%v", err)
		return err
	}
	return nil
}

func IsTokenBlacklisted(ctx context.Context, token string) bool {
	if dbclient.RedisCli == nil {
		return false
	}
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		return false
	}
	key := TokenBlacklistKeyPrefix + HashToken(token)
	exists, err := dbclient.RedisCli.Exists(ctx, key).Result()
	if err != nil {
		return false
	}
	return exists > 0
}
