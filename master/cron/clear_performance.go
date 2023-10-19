package cron

import (
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"
	"time"
)

// 由于BoltDB的设计使然，即使明明已删除了最旧的日志条目，BoltDB在磁盘上使用的空间也不会缩小。相反，
// 所有用于存储已删除数据的页面（文件中的4kb段）而是都被标记为“空闲”，并重新用于后续的写入。
// BoltDB在一种名为“空闲链表”（freelist）的结构中跟踪这些空闲页面。通常，写入延迟并不受更新空闲链表所需的时间的显著影响

// ClearPerformance 清理旧的 performance 数据
func ClearPerformance() {
	for {
		time.Sleep(100 * time.Second)
		logger.Info("开始清理旧的 performance 数据 ")
		// 按梯度清理
		logger.Info("清理 一年前的数据 每天保留一条")
		Level1()
		time.Sleep(10 * time.Second)

		logger.Info("清理 180天前到365天前的数据 每3小时保留1条")
		Level2()
		time.Sleep(10 * time.Second)

		logger.Info("清理 90天前到180天前的数据 每小时保留1条")
		Level3()
		time.Sleep(10 * time.Second)

		logger.Info("清理 30天前到90天前的数据 每10分钟保留1条")
		Level4()
		time.Sleep(10 * time.Second)

		logger.Info("清理 1天前到30天前 每分钟保留1条")
		Level5()
		time.Sleep(240 * time.Hour)
	}

}

// Level1 一年前的数据 每天保留一条
func Level1() {
	data, _ := new(dao.DaoSlave).GetALL()
	for _, slave := range data {
		max := 500
		min := 365
		for max > min {
			section, _ := new(dao.DaoPerformance).GetPerformanceMinuteSection(slave.Slave,
				utils.Unix2Date(utils.DayAgo(max)),
				utils.Unix2Date(utils.DayAgo(max-1)))
			n := 0
			for k, _ := range section {
				if n > 1 {
					logger.Info("清理 ==>  ", k)
					new(dao.DaoPerformance).Del(slave.Slave, k)
				}
				n++
			}
			max--
		}
	}
}

// Level2 180天前到365天前的数据 每3小时保留1条
func Level2() {
	data, _ := new(dao.DaoSlave).GetALL()
	for _, slave := range data {
		max := 365 * 24
		min := 180 * 24
		for max > min {
			section, _ := new(dao.DaoPerformance).GetPerformanceMinuteSection(slave.Slave,
				utils.Unix2Date(utils.HourAgo(max)),
				utils.Unix2Date(utils.HourAgo(max-3)))
			n := 0
			for k, _ := range section {
				if n > 1 {
					logger.Info("清理 ==>  ", k)
					new(dao.DaoPerformance).Del(slave.Slave, k)
				}
				n++
			}
			max -= 3
		}
	}
}

// Level3 90天前到180天前的数据 每小时保留1条
func Level3() {
	data, _ := new(dao.DaoSlave).GetALL()
	for _, slave := range data {
		max := 180 * 24
		min := 90 * 24
		for max > min {
			section, _ := new(dao.DaoPerformance).GetPerformanceMinuteSection(slave.Slave,
				utils.Unix2Date(utils.HourAgo(max)),
				utils.Unix2Date(utils.HourAgo(max-1)))
			n := 0
			for k, _ := range section {
				if n > 1 {
					logger.Info("清理 ==>  ", k)
					new(dao.DaoPerformance).Del(slave.Slave, k)
				}
				n++
			}
			max--
		}
	}
}

// Level4 30天前到90天前的数据 每10分钟保留1条
func Level4() {
	data, _ := new(dao.DaoSlave).GetALL()
	for _, slave := range data {
		max := 90 * 24 * 60
		min := 30 * 24 * 60
		for max > min {
			section, _ := new(dao.DaoPerformance).GetPerformanceMinuteSection(slave.Slave,
				utils.Unix2Date(utils.MinuteAgo(max)),
				utils.Unix2Date(utils.MinuteAgo(max-10)))
			n := 0
			for k, _ := range section {
				if n > 1 {
					logger.Info("清理 ==>  ", k)
					new(dao.DaoPerformance).Del(slave.Slave, k)
				}
				n++
			}
			max -= 10
		}
	}
}

// Level5 7天前到30天前 每分钟保留1条
func Level5() {
	logger.Info("7天前到30天前 每分钟保留1条")
	data, _ := new(dao.DaoSlave).GetALL()
	for _, slave := range data {
		max := 30 * 24 * 60
		min := 1 * 24 * 60
		for max > min {
			maxDate := utils.Unix2Date(utils.MinuteAgo(max))
			minDate := utils.Unix2Date(utils.MinuteAgo(max - 1))
			section, _ := new(dao.DaoPerformance).GetPerformanceMinuteSection(slave.Slave,
				minDate,
				maxDate)
			n := 0
			for k, v := range section {
				logger.Info(k, " --> ", v)
				if n > 1 {
					logger.Info("清理 ==>  ", k)
					new(dao.DaoPerformance).Del(slave.Slave, k)
				}
				n++
			}
			max--
			break
		}
	}
}
