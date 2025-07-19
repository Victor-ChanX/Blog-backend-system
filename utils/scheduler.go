package utils

import (
	"blog-server/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// StartScheduler 启动定时任务
func StartScheduler() {
	// 每天凌晨0:05执行数据转存任务
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 5, 0, 0, now.Location())
			duration := next.Sub(now)
			
			log.Printf("下次数据转存任务将在 %v 后执行", duration)
			time.Sleep(duration)
			
			if err := TransferDataToPostgreSQL(); err != nil {
				log.Printf("数据转存任务失败: %v", err)
			} else {
				log.Println("数据转存任务完成")
			}
		}
	}()
}

// TransferDataToPostgreSQL 将Redis数据转存到PostgreSQL
func TransferDataToPostgreSQL() error {
	// 转存昨天的数据
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	return TransferSpecificDateData(yesterday)
}

// TransferSpecificDateData 转存指定日期的数据
func TransferSpecificDateData(date string) error {
	ctx := context.Background()

	log.Printf("开始转存 %s 的数据", date)

	// 1. 获取追踪数据
	trackingDataList, err := GetTrackingDataForDate(ctx, date)
	if err != nil {
		return fmt.Errorf("获取追踪数据失败: %v", err)
	}

	if len(trackingDataList) == 0 {
		log.Printf("日期 %s 没有追踪数据", date)
		return nil
	}

	// 2. 转存追踪事件到数据库
	if err := storeTrackingEvents(trackingDataList); err != nil {
		return fmt.Errorf("存储追踪事件失败: %v", err)
	}

	// 3. 生成并存储每日统计数据
	if err := generateDailyStats(date, trackingDataList); err != nil {
		return fmt.Errorf("生成每日统计失败: %v", err)
	}

	// 4. 生成并存储页面热力图数据
	if err := generatePageHeatmap(date, trackingDataList); err != nil {
		return fmt.Errorf("生成页面热力图失败: %v", err)
	}

	// 5. 清理Redis中的数据
	if err := ClearTrackingDataForDate(ctx, date); err != nil {
		log.Printf("清理Redis数据失败: %v", err) // 不作为致命错误
	}

	log.Printf("成功转存 %s 的数据，共 %d 条记录", date, len(trackingDataList))
	return nil
}

// storeTrackingEvents 存储追踪事件到数据库
func storeTrackingEvents(trackingDataList []TrackingData) error {
	var events []models.TrackingEvent

	for _, data := range trackingDataList {
		event := models.TrackingEvent{
			Timestamp: data.Timestamp,
			Path:      data.Path,
			IPAddress: data.IPAddress,
			UserAgent: data.UserAgent,
			Referer:   data.Referer,
			EventType: data.EventType,
			ArticleID: data.ArticleID,
			SessionID: data.SessionID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		events = append(events, event)
	}

	// 批量插入
	if err := models.DB.CreateInBatches(events, 100).Error; err != nil {
		return err
	}

	return nil
}

// generateDailyStats 生成每日统计数据
func generateDailyStats(date string, trackingDataList []TrackingData) error {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return err
	}

	// 统计数据
	pageViews := int64(len(trackingDataList))
	uniqueVisitors := make(map[string]bool)
	articleClicks := int64(0)
	pathCounts := make(map[string]int64)
	articleCounts := make(map[string]int64)

	for _, data := range trackingDataList {
		// 独立访客统计
		uniqueVisitors[data.IPAddress] = true

		// 文章点击统计
		if data.EventType == "article_click" {
			articleClicks++
			if data.ArticleID != nil {
				key := fmt.Sprintf("%d", *data.ArticleID)
				articleCounts[key]++
			}
		}

		// 路径访问统计
		pathCounts[data.Path]++
	}

	// 生成热门页面和文章的JSON
	topPagesJSON, _ := json.Marshal(pathCounts)
	topArticlesJSON, _ := json.Marshal(articleCounts)

	// 检查是否已存在该日期的统计
	var existingStats models.DailyStats
	if err := models.DB.Where("date = ?", parsedDate).First(&existingStats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新记录
			dailyStats := models.DailyStats{
				Date:           parsedDate,
				PageViews:      pageViews,
				UniqueVisitors: int64(len(uniqueVisitors)),
				ArticleClicks:  articleClicks,
				TopPages:       string(topPagesJSON),
				TopArticles:    string(topArticlesJSON),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			if err := models.DB.Create(&dailyStats).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// 更新现有记录
		existingStats.PageViews = pageViews
		existingStats.UniqueVisitors = int64(len(uniqueVisitors))
		existingStats.ArticleClicks = articleClicks
		existingStats.TopPages = string(topPagesJSON)
		existingStats.TopArticles = string(topArticlesJSON)
		existingStats.UpdatedAt = time.Now()

		if err := models.DB.Save(&existingStats).Error; err != nil {
			return err
		}
	}

	return nil
}

// generatePageHeatmap 生成页面热力图数据
func generatePageHeatmap(date string, trackingDataList []TrackingData) error {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return err
	}

	// 统计每个路径的访问和点击
	pathStats := make(map[string]struct {
		views  int64
		clicks int64
	})

	for _, data := range trackingDataList {
		stats := pathStats[data.Path]
		stats.views++

		if data.EventType == "article_click" {
			stats.clicks++
		}

		pathStats[data.Path] = stats
	}

	// 批量处理页面热力图数据
	var heatmapData []models.PageHeatmap
	for path, stats := range pathStats {
		heatmap := models.PageHeatmap{
			Date:      parsedDate,
			Path:      path,
			Clicks:    stats.clicks,
			Views:     stats.views,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		heatmapData = append(heatmapData, heatmap)
	}

	// 先删除当天已有的数据，再插入新数据
	if err := models.DB.Where("date = ?", parsedDate).Delete(&models.PageHeatmap{}).Error; err != nil {
		return err
	}

	if len(heatmapData) > 0 {
		if err := models.DB.CreateInBatches(heatmapData, 50).Error; err != nil {
			return err
		}
	}

	return nil
}

// ManualTransferData 手动转存指定日期的数据（用于接口调用）
func ManualTransferData(date string) error {
	return TransferSpecificDateData(date)
}

// GetRedisDataSummary 获取Redis中的数据摘要
func GetRedisDataSummary(ctx context.Context) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// 获取今天的数据量
	today := time.Now().Format("2006-01-02")
	todayData, err := GetTrackingDataForDate(ctx, today)
	if err == nil {
		summary["today_records"] = len(todayData)
	}

	// 获取在线用户数
	onlineUsers, err := GetOnlineUsersCount(ctx)
	if err == nil {
		summary["online_users"] = onlineUsers
	}

	// 获取今日统计
	todayStats, err := GetTodayStats(ctx)
	if err == nil {
		summary["today_stats"] = todayStats
	}

	return summary, nil
}