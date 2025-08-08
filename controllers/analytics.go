package controllers

import (
	"blog-server/models"
	"blog-server/utils"
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TrackRequest struct {
	Path      string `json:"path" binding:"required"`
	Referer   string `json:"referer"`
	EventType string `json:"event_type" binding:"required"` // page_view, article_click, etc.
	ArticleID *uint  `json:"article_id,omitempty"`
}

type AnalyticsResponse struct {
	OnlineUsers    int64            `json:"online_users"`
	TodayVisitors  int64            `json:"today_visitors"`
	TodayPageViews int64            `json:"today_page_views"`
	TopPaths       map[string]int64 `json:"top_paths"`
	TopArticles    map[string]int64 `json:"top_articles"`
}

// Track 数据收集接口
func Track(c *gin.Context) {
	var req TrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取客户端信息
	ipAddress := utils.GetRealClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	userAgentHash := utils.HashUserAgent(userAgent)

	// 生成会话ID（基于IP和User-Agent的简单哈希）
	sessionData := fmt.Sprintf("%s_%s_%s", ipAddress, userAgent, time.Now().Format("2006-01-02"))
	sessionID := fmt.Sprintf("%x", md5.Sum([]byte(sessionData)))

	// 构建追踪数据
	trackingData := utils.TrackingData{
		Timestamp: time.Now(),
		Path:      req.Path,
		IPAddress: ipAddress,
		UserAgent: userAgentHash,
		Referer:   req.Referer,
		EventType: req.EventType,
		ArticleID: req.ArticleID,
		SessionID: sessionID,
	}

	// 存储到Redis
	ctx := c.Request.Context()
	if err := utils.StoreTrackingData(ctx, trackingData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "数据存储失败: " + err.Error(),
		})
		return
	}

	// 更新在线用户数量
	if err := utils.UpdateOnlineUsers(ctx, ipAddress); err != nil {
		// 这个错误不影响主要功能，只记录日志
		fmt.Printf("更新在线用户数量失败: %v\n", err)
	}

	// 更新页面访问统计
	if err := utils.UpdatePageViewStats(ctx, req.Path, req.EventType, req.ArticleID); err != nil {
		fmt.Printf("更新页面访问统计失败: %v\n", err)
	}

	// 更新独立访客统计
	if err := utils.UpdateUniqueVisitors(ctx, ipAddress); err != nil {
		fmt.Printf("更新独立访客统计失败: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "数据收集成功",
		"data": gin.H{
			"timestamp":  trackingData.Timestamp,
			"session_id": sessionID,
		},
	})
}

// GetRealTimeStats 获取实时统计数据
func GetRealTimeStats(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取在线用户数量
	onlineUsers, err := utils.GetOnlineUsersCount(ctx)
	if err != nil {
		onlineUsers = 0
	}

	// 获取今日统计数据
	todayStats, err := utils.GetTodayStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取统计数据失败: " + err.Error(),
		})
		return
	}

	// 解析统计数据
	var todayVisitors, todayPageViews int64
	if val, ok := todayStats["unique_visitors"]; ok {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			todayVisitors = parsed
		}
	}
	if val, ok := todayStats["page_views"]; ok {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			todayPageViews = parsed
		}
	}

	// 统计热门路径
	topPaths := make(map[string]int64)
	topArticles := make(map[string]int64)
	
	for key, val := range todayStats {
		if count, err := strconv.ParseInt(val, 10, 64); err == nil {
			if key[:5] == "path:" {
				path := key[5:]
				topPaths[path] = count
			} else if key[:8] == "article:" {
				articleID := key[8:]
				topArticles[articleID] = count
			}
		}
	}

	response := AnalyticsResponse{
		OnlineUsers:    onlineUsers,
		TodayVisitors:  todayVisitors,
		TodayPageViews: todayPageViews,
		TopPaths:       topPaths,
		TopArticles:    topArticles,
	}

	c.JSON(http.StatusOK, response)
}

// GetDailyStats 获取指定日期的统计数据
func GetDailyStats(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// 验证日期格式并解析为时间对象
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "日期格式错误，请使用YYYY-MM-DD格式",
		})
		return
	}

	// 查询数据库中的统计数据，使用DATE函数进行日期比较
	var dailyStats models.DailyStats
	if err := models.DB.Where("DATE(date) = DATE(?)", parsedDate).First(&dailyStats).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "未找到指定日期的统计数据",
			"date": dateStr,
		})
		return
	}

	c.JSON(http.StatusOK, dailyStats)
}

// GetStatsRange 获取日期范围内的统计数据
func GetStatsRange(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请提供start_date和end_date参数",
		})
		return
	}

	// 验证日期格式
	_, err1 := time.Parse("2006-01-02", startDate)
	_, err2 := time.Parse("2006-01-02", endDate)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "日期格式错误，请使用YYYY-MM-DD格式",
		})
		return
	}

	var statsRange []models.DailyStats
	if err := models.DB.Where("date BETWEEN ? AND ?", startDate, endDate).
		Order("date ASC").Find(&statsRange).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询统计数据失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       statsRange,
		"start_date": startDate,
		"end_date":   endDate,
		"count":      len(statsRange),
	})
}

// GetTopPages 获取热门页面统计
func GetTopPages(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	var heatmapData []models.PageHeatmap
	if err := models.DB.Where("date = ?", dateStr).
		Order("views DESC, clicks DESC").
		Limit(limit).
		Find(&heatmapData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询热门页面失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date":  dateStr,
		"data":  heatmapData,
		"limit": limit,
	})
}

// GetTrackingEvents 获取详细访问记录
func GetTrackingEvents(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 50
	}
	offset := (page - 1) * limit

	// 过滤参数
	path := c.Query("path")
	eventType := c.Query("event_type")
	ipAddress := c.Query("ip_address")

	// 构建查询
	query := models.DB.Model(&models.TrackingEvent{})
	
	// 日期过滤
	startDate := dateStr + " 00:00:00"
	endDate := dateStr + " 23:59:59"
	query = query.Where("timestamp BETWEEN ? AND ?", startDate, endDate)

	// 其他过滤条件
	if path != "" {
		query = query.Where("path LIKE ?", "%"+path+"%")
	}
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	if ipAddress != "" {
		query = query.Where("ip_address = ?", ipAddress)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取数据
	var events []models.TrackingEvent
	if err := query.Order("timestamp DESC").
		Offset(offset).
		Limit(limit).
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询访问记录失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": events,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
		"filters": gin.H{
			"date":       dateStr,
			"path":       path,
			"event_type": eventType,
			"ip_address": ipAddress,
		},
	})
}

// GetIPStats 获取IP访问统计
func GetIPStats(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit < 1 || limit > 1000 {
		limit = 100
	}

	startDate := dateStr + " 00:00:00"
	endDate := dateStr + " 23:59:59"

	// 统计IP访问次数
	type IPStat struct {
		IPAddress string `json:"ip_address"`
		Count     int64  `json:"count"`
		LastVisit string `json:"last_visit"`
	}

	var ipStats []IPStat
	if err := models.DB.Model(&models.TrackingEvent{}).
		Select("ip_address, COUNT(*) as count, MAX(timestamp) as last_visit").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Group("ip_address").
		Order("count DESC").
		Limit(limit).
		Scan(&ipStats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询IP统计失败: " + err.Error(),
		})
		return
	}

	// 获取总IP数量
	var totalIPs int64
	models.DB.Model(&models.TrackingEvent{}).
		Select("DISTINCT ip_address").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Count(&totalIPs)

	c.JSON(http.StatusOK, gin.H{
		"date":      dateStr,
		"data":      ipStats,
		"total_ips": totalIPs,
		"limit":     limit,
	})
}

// GetUserAgentStats 获取User-Agent统计
func GetUserAgentStats(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 500 {
		limit = 50
	}

	startDate := dateStr + " 00:00:00"
	endDate := dateStr + " 23:59:59"

	type UserAgentStat struct {
		UserAgent string `json:"user_agent_hash"`
		Count     int64  `json:"count"`
		LastSeen  string `json:"last_seen"`
	}

	var uaStats []UserAgentStat
	if err := models.DB.Model(&models.TrackingEvent{}).
		Select("user_agent, COUNT(*) as count, MAX(timestamp) as last_seen").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Group("user_agent").
		Order("count DESC").
		Limit(limit).
		Scan(&uaStats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询User-Agent统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date":  dateStr,
		"data":  uaStats,
		"limit": limit,
	})
}

// GetRefererStats 获取来源统计
func GetRefererStats(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 500 {
		limit = 50
	}

	startDate := dateStr + " 00:00:00"
	endDate := dateStr + " 23:59:59"

	type RefererStat struct {
		Referer string `json:"referer"`
		Count   int64  `json:"count"`
	}

	var refererStats []RefererStat
	if err := models.DB.Model(&models.TrackingEvent{}).
		Select("referer, COUNT(*) as count").
		Where("timestamp BETWEEN ? AND ? AND referer != ''", startDate, endDate).
		Group("referer").
		Order("count DESC").
		Limit(limit).
		Scan(&refererStats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询来源统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date":  dateStr,
		"data":  refererStats,
		"limit": limit,
	})
}

// GetSessionStats 获取会话统计
func GetSessionStats(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	startDate := dateStr + " 00:00:00"
	endDate := dateStr + " 23:59:59"

	// 会话统计
	type SessionStat struct {
		SessionID    string `json:"session_id"`
		EventCount   int64  `json:"event_count"`
		FirstVisit   string `json:"first_visit"`
		LastVisit    string `json:"last_visit"`
		IPAddress    string `json:"ip_address"`
		Duration     int64  `json:"duration_seconds"`
	}

	var sessionStats []SessionStat
	if err := models.DB.Model(&models.TrackingEvent{}).
		Select(`session_id, 
				COUNT(*) as event_count, 
				MIN(timestamp) as first_visit, 
				MAX(timestamp) as last_visit, 
				ip_address,
				EXTRACT(EPOCH FROM (MAX(timestamp) - MIN(timestamp))) as duration_seconds`).
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Group("session_id, ip_address").
		Order("event_count DESC").
		Limit(100).
		Scan(&sessionStats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询会话统计失败: " + err.Error(),
		})
		return
	}

	// 总会话数
	var totalSessions int64
	models.DB.Model(&models.TrackingEvent{}).
		Select("DISTINCT session_id").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Count(&totalSessions)

	c.JSON(http.StatusOK, gin.H{
		"date":           dateStr,
		"data":           sessionStats,
		"total_sessions": totalSessions,
	})
}

// GetEventTypeStats 获取事件类型统计
func GetEventTypeStats(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	startDate := dateStr + " 00:00:00"
	endDate := dateStr + " 23:59:59"

	type EventTypeStat struct {
		EventType string `json:"event_type"`
		Count     int64  `json:"count"`
	}

	var eventStats []EventTypeStat
	if err := models.DB.Model(&models.TrackingEvent{}).
		Select("event_type, COUNT(*) as count").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Group("event_type").
		Order("count DESC").
		Scan(&eventStats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询事件类型统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date": dateStr,
		"data": eventStats,
	})
}

// GetHourlyStats 获取按小时统计
func GetHourlyStats(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	startDate := dateStr + " 00:00:00"
	endDate := dateStr + " 23:59:59"

	type HourlyStat struct {
		Hour  int   `json:"hour"`
		Count int64 `json:"count"`
	}

	var hourlyStats []HourlyStat
	if err := models.DB.Model(&models.TrackingEvent{}).
		Select("EXTRACT(HOUR FROM timestamp) as hour, COUNT(*) as count").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Group("EXTRACT(HOUR FROM timestamp)").
		Order("hour").
		Scan(&hourlyStats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询小时统计失败: " + err.Error(),
		})
		return
	}

	// 填充24小时数据
	hourlyMap := make(map[int]int64)
	for _, stat := range hourlyStats {
		hourlyMap[stat.Hour] = stat.Count
	}

	var result []HourlyStat
	for i := 0; i < 24; i++ {
		result = append(result, HourlyStat{
			Hour:  i,
			Count: hourlyMap[i],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"date": dateStr,
		"data": result,
	})
}

// GetPathAnalysis 获取路径详细分析
func GetPathAnalysis(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	startDate := dateStr + " 00:00:00"
	endDate := dateStr + " 23:59:59"

	type PathAnalysis struct {
		Path            string  `json:"path"`
		TotalViews      int64   `json:"total_views"`
		UniqueVisitors  int64   `json:"unique_visitors"`
		AverageStayTime float64 `json:"average_stay_time"`
		BounceRate      float64 `json:"bounce_rate"`
	}

	var pathAnalysis []PathAnalysis
	if err := models.DB.Model(&models.TrackingEvent{}).
		Select(`path, 
				COUNT(*) as total_views,
				COUNT(DISTINCT ip_address) as unique_visitors`).
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Group("path").
		Order("total_views DESC").
		Scan(&pathAnalysis).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询路径分析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date": dateStr,
		"data": pathAnalysis,
	})
}

// GetAdvancedStats 获取高级统计数据
func GetAdvancedStats(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	startDate := dateStr + " 00:00:00"
	endDate := dateStr + " 23:59:59"

	// 基础统计
	var totalEvents, uniqueIPs, uniqueSessions int64
	models.DB.Model(&models.TrackingEvent{}).
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Count(&totalEvents)
	
	models.DB.Model(&models.TrackingEvent{}).
		Select("DISTINCT ip_address").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Count(&uniqueIPs)
	
	models.DB.Model(&models.TrackingEvent{}).
		Select("DISTINCT session_id").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Count(&uniqueSessions)

	// 新访客vs回访客
	type VisitorType struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	// 平均会话时长
	var avgSessionDuration float64
	models.DB.Model(&models.TrackingEvent{}).
		Select("AVG(EXTRACT(EPOCH FROM (MAX(timestamp) - MIN(timestamp))))").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Group("session_id").
		Scan(&avgSessionDuration)

	c.JSON(http.StatusOK, gin.H{
		"date":                  dateStr,
		"total_events":          totalEvents,
		"unique_visitors":       uniqueIPs,
		"unique_sessions":       uniqueSessions,
		"avg_session_duration":  avgSessionDuration,
		"avg_events_per_session": float64(totalEvents) / float64(uniqueSessions),
	})
}