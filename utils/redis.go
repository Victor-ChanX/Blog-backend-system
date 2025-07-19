package utils

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitRedis 初始化Redis连接
func InitRedis() error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")
	db := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if dbNum, err := strconv.Atoi(dbStr); err == nil {
			db = dbNum
		}
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis连接失败: %v", err)
	}

	log.Println("Redis连接成功")
	return nil
}

// HashUserAgent 对User-Agent进行SHA256哈希
func HashUserAgent(userAgent string) string {
	hash := sha256.Sum256([]byte(userAgent))
	return fmt.Sprintf("%x", hash)
}

// GetTodayKey 获取今天的Redis键名
func GetTodayKey(prefix string) string {
	today := time.Now().Format("2006-01-02")
	return fmt.Sprintf("%s:%s", prefix, today)
}

// TrackingData Redis中存储的追踪数据结构
type TrackingData struct {
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent_hash"`
	Referer   string    `json:"referer"`
	EventType string    `json:"event_type"`
	ArticleID *uint     `json:"article_id,omitempty"`
	SessionID string    `json:"session_id"`
}

// StoreTrackingData 存储追踪数据到Redis
func StoreTrackingData(ctx context.Context, data TrackingData) error {
	// 存储到今天的列表中
	todayKey := GetTodayKey("tracking")
	
	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化追踪数据失败: %v", err)
	}

	// 存储到Redis列表
	err = RedisClient.LPush(ctx, todayKey, jsonData).Err()
	if err != nil {
		return fmt.Errorf("存储追踪数据到Redis失败: %v", err)
	}

	// 设置键的过期时间为3天
	RedisClient.Expire(ctx, todayKey, 72*time.Hour)

	return nil
}

// UpdateOnlineUsers 更新在线用户数量
func UpdateOnlineUsers(ctx context.Context, ipAddress string) error {
	key := "online_users"
	// 使用IP地址作为成员，设置30分钟过期
	score := float64(time.Now().Unix())
	
	// 添加到有序集合
	err := RedisClient.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: ipAddress,
	}).Err()
	if err != nil {
		return err
	}

	// 清理30分钟前的记录
	thirtyMinutesAgo := float64(time.Now().Add(-30 * time.Minute).Unix())
	RedisClient.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%.0f", thirtyMinutesAgo))

	return nil
}

// GetOnlineUsersCount 获取在线用户数量
func GetOnlineUsersCount(ctx context.Context) (int64, error) {
	count, err := RedisClient.ZCard(ctx, "online_users").Result()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// UpdatePageViewStats 更新页面访问统计
func UpdatePageViewStats(ctx context.Context, path string, eventType string, articleID *uint) error {
	todayKey := GetTodayKey("stats")
	
	// 总页面访问量
	RedisClient.HIncrBy(ctx, todayKey, "page_views", 1)
	
	// 路径访问统计
	pathKey := fmt.Sprintf("path:%s", path)
	RedisClient.HIncrBy(ctx, todayKey, pathKey, 1)
	
	// 如果是文章点击，更新文章统计
	if eventType == "article_click" && articleID != nil {
		articleKey := fmt.Sprintf("article:%d", *articleID)
		RedisClient.HIncrBy(ctx, todayKey, articleKey, 1)
	}

	// 设置过期时间
	RedisClient.Expire(ctx, todayKey, 72*time.Hour)
	
	return nil
}

// UpdateUniqueVisitors 更新独立访客统计
func UpdateUniqueVisitors(ctx context.Context, ipAddress string) error {
	todayKey := GetTodayKey("visitors")
	
	// 使用Redis Set存储今日访客IP
	isNew, err := RedisClient.SAdd(ctx, todayKey, ipAddress).Result()
	if err != nil {
		return err
	}

	// 如果是新访客，更新统计
	if isNew > 0 {
		statsKey := GetTodayKey("stats")
		RedisClient.HIncrBy(ctx, statsKey, "unique_visitors", 1)
	}

	// 设置过期时间
	RedisClient.Expire(ctx, todayKey, 72*time.Hour)
	
	return nil
}

// GetTodayStats 获取今日统计数据
func GetTodayStats(ctx context.Context) (map[string]string, error) {
	todayKey := GetTodayKey("stats")
	stats, err := RedisClient.HGetAll(ctx, todayKey).Result()
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// GetTrackingDataForDate 获取指定日期的追踪数据
func GetTrackingDataForDate(ctx context.Context, date string) ([]TrackingData, error) {
	key := fmt.Sprintf("tracking:%s", date)
	
	// 获取所有数据
	jsonDataList, err := RedisClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var trackingDataList []TrackingData
	for _, jsonData := range jsonDataList {
		var data TrackingData
		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			log.Printf("反序列化追踪数据失败: %v", err)
			continue
		}
		trackingDataList = append(trackingDataList, data)
	}

	return trackingDataList, nil
}

// ClearTrackingDataForDate 清理指定日期的追踪数据
func ClearTrackingDataForDate(ctx context.Context, date string) error {
	key := fmt.Sprintf("tracking:%s", date)
	return RedisClient.Del(ctx, key).Err()
}