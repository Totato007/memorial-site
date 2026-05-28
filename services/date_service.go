package services

import (
	"fmt"
	"time"
)

// CalculateDaysTogether 计算从起始日期到今天的总天数
func CalculateDaysTogether(startDate time.Time) int {
	return int(time.Since(startDate).Hours() / 24)
}

// NextAnniversary 计算下一个周年纪念日及倒计时天数
func NextAnniversary(startDate time.Time) (time.Time, int) {
	now := time.Now()
	anniv := time.Date(now.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, now.Location())
	if anniv.Before(now) || anniv.Equal(now) {
		anniv = time.Date(now.Year()+1, startDate.Month(), startDate.Day(), 0, 0, 0, 0, now.Location())
	}
	daysUntil := int(anniv.Sub(now).Hours() / 24)
	return anniv, daysUntil
}

// FormatDate 格式化日期 "2006年01月02日"
func FormatDate(t time.Time) string {
	return t.Format("2006年01月02日")
}

// OnlineStatus 根据最后活跃时间返回状态文本和 CSS class
func OnlineStatus(lastActive *time.Time) (string, string) {
	if lastActive == nil {
		return "从未上线", "offline"
	}
	d := time.Since(*lastActive)
	minutes := int(d.Minutes())
	switch {
	case minutes < 5:
		return "在线", "online"
	case minutes < 60:
		return fmt.Sprintf("%d分钟前在线", minutes), "away"
	case minutes < 1440:
		return fmt.Sprintf("%d小时前在线", minutes/60), "away"
	default:
		days := minutes / 1440
		if days > 30 {
			return "离线", "offline"
		}
		return fmt.Sprintf("%d天前在线", days), "away"
	}
}

// OnlineStatusShort 短格式在线状态，用于好友列表
func OnlineStatusShort(lastActive *time.Time) string {
	if lastActive == nil {
		return "离线"
	}
	d := time.Since(*lastActive)
	minutes := int(d.Minutes())
	switch {
	case minutes < 5:
		return "在线"
	case minutes < 60:
		return fmt.Sprintf("%d分钟前", minutes)
	case minutes < 1440:
		return fmt.Sprintf("%d小时前", minutes/60)
	default:
		days := minutes / 1440
		if days > 30 {
			return "离线"
		}
		return fmt.Sprintf("%d天前", days)
	}
}

// SmartTime 智能时间显示 — 今天显示"今天 15:04"，否则显示"2006-01-02 15:04"
func SmartTime(t time.Time) string {
	now := time.Now()
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return "今天 " + t.Format("15:04")
	}
	return t.Format("2006-01-02 15:04")
}

// SmartDate 智能日期 — 今天显示"今天"，否则显示"01月02日"
func SmartDate(t time.Time) string {
	now := time.Now()
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return "今天"
	}
	return t.Format("01月02日")
}

// DaysInYear 计算两个日期之间跨越的年数和余天
func DaysInYear(startDate time.Time) (years int, days int) {
	now := time.Now()
	totalDays := int(now.Sub(startDate).Hours() / 24)
	years = totalDays / 365
	days = totalDays % 365
	return
}
