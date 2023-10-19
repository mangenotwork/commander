package check

import (
	"os"
	"strings"

	"gitee.com/mangenotework/commander/common/conf"
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"
)

// SlaveInitPath 检查配置地址数据持久文件
func SlaveInitPath(){
	confDBPathDataOk, _ := utils.PathExists(conf.SlaveConf.DBPath.Data)
	if !confDBPathDataOk {
		DBPathDataList := strings.Split(conf.SlaveConf.DBPath.Data, "/")
		DBPathDataDir := strings.Join(DBPathDataList[0:len(DBPathDataList)-1], "/")
		// logger.Info("DBPathDataDir = ", DBPathDataDir)
		err := os.MkdirAll(DBPathDataDir, 0777)
		if err != nil {
			logger.Error(err)
		}
		f, err := os.Create(conf.SlaveConf.DBPath.Data)
		defer f.Close()
		if err != nil {
			logger.Error(err)
		}
	}

	confProjectExeStoreHousePath, _ := utils.PathExists(conf.SlaveConf.ProjectExeStoreHousePath)
	if !confProjectExeStoreHousePath {
		err := os.MkdirAll(conf.SlaveConf.ProjectExeStoreHousePath, 0777)
		err = os.Chmod(conf.SlaveConf.ProjectExeStoreHousePath, 0777)
		if err != nil {
			logger.Error(conf.SlaveConf.ProjectExeStoreHousePath, err)
		}
	}

	confExeStoreHouseLogs, _ := utils.PathExists(conf.SlaveConf.ExeStoreHouseLogs)
	if !confExeStoreHouseLogs {
		err := os.MkdirAll(conf.SlaveConf.ExeStoreHouseLogs, 0777)
		err = os.Chmod(conf.SlaveConf.ExeStoreHouseLogs, 0777)
		if err != nil {
			logger.Error(conf.SlaveConf.ExeStoreHouseLogs, err)
		}
	}

	confExeStoreHousePath, _ := utils.PathExists(conf.SlaveConf.ExeStoreHousePath)
	if !confExeStoreHousePath {
		err := os.MkdirAll(conf.SlaveConf.ExeStoreHousePath, 0777)
		err = os.Chmod(conf.SlaveConf.ExeStoreHousePath, 0777)
		if err != nil {
			logger.Error(conf.SlaveConf.ExeStoreHousePath, err)
		}
	}
}

// MasterInitPath master 初始化配置文件并创建路径
func MasterInitPath(){
	confDBPathData, _ := utils.PathExists(conf.MasterConf.DBPath.Data)
	if !confDBPathData {
		pathList := strings.Split(conf.MasterConf.DBPath.Data, "/")
		dir := strings.Join(pathList[0:len(pathList)-1], "/")
 		//logger.Error("conf.MasterConf.DBPath.Data = ", conf.MasterConf.DBPath.Data, dir)
		mkdirErr := os.MkdirAll(dir, 0777)
		if mkdirErr != nil {
			logger.Error("mkdir Err = ", mkdirErr)
		}
		f, err := os.Create(conf.MasterConf.DBPath.Data)
		defer f.Close()
		if err != nil {
			logger.Error(err)
		}
	}

	confDBPathPerformance, _ := utils.PathExists(conf.MasterConf.DBPath.Performance)
	if !confDBPathPerformance {
		os.MkdirAll(conf.MasterConf.DBPath.Data, 0777)
		f, err := os.Create(conf.MasterConf.DBPath.Performance)
		defer f.Close()
		if err != nil {
			logger.Error(err)
		}
	}

	confExeStoreHousePath, _ := utils.PathExists(conf.MasterConf.ExeStoreHousePath)
	if !confExeStoreHousePath {
		os.Mkdir(conf.MasterConf.ExeStoreHousePath, 0777)
		os.Chmod(conf.MasterConf.ExeStoreHousePath, 0777)
	}

	confProjectPath, _ := utils.PathExists(conf.MasterConf.ProjectPath)
	if !confProjectPath {
		os.Mkdir(conf.MasterConf.ProjectPath, 0777)
		os.Chmod(conf.MasterConf.ProjectPath, 0777)
	}
}
