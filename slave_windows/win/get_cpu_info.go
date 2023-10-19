package win

/*
#ifdef WIN32
#include "cpu_info.h"
#include "cpu_use.h"
#endif
*/
import "C"

import (
	"gitee.com/mangenotework/commander/common/cmd"
	"log"
	"os"
	"unsafe"
)

// GetCpuVendorId 获取cpu的VendorId
func GetCpuVendorId() string {
	//定义一个字符串指针接收C 函数返回值
	var out *C.char = C.WindowsGetCpuVendorId()
	//释放这个指针  前提是使用了malloc
	defer C.free(unsafe.Pointer(out))
	return C.GoString(out)
}

func GetCpuVendorId1() string {
	return C.GoString(C.WindowsGetCpuVendorId())
}

func GetCpuVendorId2() string {
	return "aaa"
}

// GetCpuId 获取cpu的 CpuId
func GetCpuId() string {
	return C.GoString(C.WindowsGetCpuId())
}

// GetCpuName 获取cup的 CpuName
func GetCpuName() string {

	//return C.GoString(C.WindowsGetCpuName())
	dir,_ := os.Getwd()
	cmds := []string{dir+"\\win_mange.exe"}
	boardIdStr := cmd.WindowsSendCommand(cmds)
	log.Println("boardIdStr = ", boardIdStr)
	return boardIdStr
}

// GetCPUUse 获取cpu使用率
func GetCPUUse() int {
	return int(C.cpu())
}
