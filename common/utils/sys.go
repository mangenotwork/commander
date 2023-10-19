package utils

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"gitee.com/mangenotework/commander/common/logger"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// GetMyIP 获取本机ip
func GetMyIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		logger.Error("[Error] :", err)
		return ""
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.String()
}

// GetSysType 获取host 系统类型
func GetSysType() string {
	return runtime.GOOS
}

// GetSysArch 获取系统架构
func GetSysArch() string {
	return runtime.GOARCH
}

// GetHostName 获取host 命名
func GetHostName() string {
	name, err := os.Hostname()
	if err != nil {
		name = "null"
	}
	return name
}

// GetCpuCoreNumber 获取cpu核心数
func GetCpuCoreNumber() string {
	return fmt.Sprintf("%d核", runtime.GOMAXPROCS(0))
}

// SysInfo 获取系统信息
func SysInfo() {
	logger.Info(`系统类型：`, runtime.GOOS)
	logger.Info(`系统架构：`, runtime.GOARCH)
	logger.Info(`CPU 核数：`, runtime.GOMAXPROCS(0))
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	logger.Info(`电脑名称：`, name)
}

// GetNetInfo 获取网卡信息
func GetNetInfo() error {
	intf, err := net.Interfaces()
	if err != nil {
		logger.Error("get network info failed: ", err)
		return err
	}
	for _, v := range intf {
		logger.Info("最大传输单元 = ", v.MTU)
		logger.Info("Name = ", v.Name)
		logger.Info("硬件地址 = ", v.HardwareAddr)
		logger.Info("接口的属性 = ", v.Flags)
		/*
			"up",  接口在活动状态
				"broadcast",   接口支持广播
				"loopback",  接口是环回的
				"pointtopoint",  接口是点对点的
				"multicast",  接口支持组播
		*/
		ips, err := v.Addrs()
		if err != nil {
			logger.Error("get network addr failed: ", err)
			return err
		}
		logger.Info("ips = ", ips)
		mips, err := v.MulticastAddrs()
		if err != nil {
			logger.Error("get network addr failed: ", err)
			return err
		}
		logger.Info("mips = ", mips)
	}
	return nil
}

// DecompressionZipFile zip压缩文件
func DecompressionZipFile(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		filePath := path.Join(dest, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
				return err
			}
			inFile, err := file.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Compress 压缩文件
// files 文件数组，可以是不同dir下的文件或者文件夹
// dest 压缩文件存放地址
func Compress(files []string, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	w := zip.NewWriter(d)
	defer w.Close()
	for _, file := range files {
		err := compress(file, "", w)
		if err != nil {
			return err
		}
	}
	return nil
}

func compress(filePath string, prefix string, zw *zip.Writer) error {
	file, err := os.Open(filePath)
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f := file.Name() + "/" + fi.Name()
			if err != nil {
				return err
			}
			err = compress(f, prefix, zw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}


// DeCompressZIP zip解压文件
func DeCompressZIP(zipFile, dest string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		filename := dest + file.Name
		err = os.MkdirAll(getDir(filename), 0755)
		if err != nil {
			return err
		}
		w, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = io.Copy(w, rc)
		if err != nil {
			return err
		}
		w.Close()
		rc.Close()
	}
	return nil
}

// DeCompressTAR tar 解压文件
func DeCompressTAR(tarFile, dest string) error {
	file, err := os.Open(tarFile)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	// tar对象读取文件内容, 遍历输出文件内容
	tr := tar.NewReader(file)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s文件内容:\n", hdr.Name)
		filename := dest + hdr.Name
		err = os.MkdirAll(getDir(filename), 0755)
		if err != nil {
			return err
		}
		w, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = io.Copy(w, tr)
		if err != nil {
			return err
		}
		w.Close()
	}
	return nil
}

// getDir
func getDir(path string) string {
	return subString(path, 0, strings.LastIndex(path, "/"))
}

// subString
func subString(str string, start, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		panic("start is wrong")
	}

	if end < start || end > length {
		panic("end is wrong")
	}

	return string(rs[start:end])
}

// GetAllFile 获取目录下的所有文件
func GetAllFile(pathname string) ([]string, error) {
	s := make([]string, 0)
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return s, err
	}

	for _, fi := range rd {
		if !fi.IsDir() {
			fullName := pathname + "/" + fi.Name()
			s = append(s, fullName)
		}
	}
	return s, nil
}
