package utils

import "time"

// NowTimeStr 获取当前时间
func NowTimeStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// NowTimeStrT2 获取当前时间
func NowTimeStrT2() string {
	return time.Now().Format("20060102150405")
}

// NowTimeStrHMS 获取当前时间
func NowTimeStrHMS() string {
	return time.Now().Format("15:04:05")
}

// NowUnix  当前时间的unix
func NowUnix() int64 {
	return time.Now().Unix()
}

// NowUnixStr 当前时间的unix string
func NowUnixStr() string {
	return StringValue(NowUnix())
}

// BeginDayUnix 获取当天凌晨的时间戳
func BeginDayUnix() int64 {
	timeStr := time.Now().Format("2006-01-02")
	t, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	return t.Unix()
}

// BeginDayUnixStr begin day unix string
func BeginDayUnixStr() string {
	return StringValue(BeginDayUnix())
}

// EndDayUnix 获取当天24点的时间戳
func EndDayUnix() int64 {
	timeStr := time.Now().Format("2006-01-02")
	t, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
	return t.Unix() + 86400
}

// EndDayUnixStr end day unix str
func EndDayUnixStr() string {
	return StringValue(EndDayUnix())
}

// MinuteAgo 获取多少分钟前的时间戳
func MinuteAgo(i int) int64 {
	return time.Now().Unix() - int64(i*60)
}

// HourAgo 获取多少小时前的时间戳
func HourAgo(i int) int64 {
	return time.Now().Unix() - int64(i*3600)
}

// DayAgo 获取多少天前的时间戳
func DayAgo(i int) int64 {
	return time.Now().Unix() - int64(i*3600*24)
}

// DayDiff 两个时间字符串的日期差
func DayDiff(beginDay string, endDay string) int {
	begin, _ := time.Parse("2006-01-02 15:04:05", beginDay+" 00:00:00")
	end, _ := time.Parse("2006-01-02 15:04:05", endDay+" 00:00:00")

	diff := end.Unix() - begin.Unix()
	return int(diff / (24 * 60 * 60))
}

// TickerRun 间隔运行
// t: 间隔时间， runFirst: 间隔前或者后执行  f: 运行的方法
func TickerRun(t time.Duration, runFirst bool, f func()) {
	if runFirst {
		f()
	}
	tick := time.NewTicker(t)
	for range tick.C {
		f()
	}
}

// Unix2Date 时间戳转日期
func Unix2Date(timeStamp int64) string {
	t := time.Unix(timeStamp, 0)
	return t.Format("2006-01-02 15:04:05")
}
