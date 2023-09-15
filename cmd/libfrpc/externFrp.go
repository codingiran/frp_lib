package main

/*
 */
import "C"
import (
	"frp_lib/cmd/frpc/lib"
)

var stListeners LibStateListeners

//export RunFrpc
func RunFrpc(cfgFilePath *C.char) C.int {
	path := C.GoString(cfgFilePath)

	if err := lib.RunFrpc(path); err != nil {
		LogPrint(err)
		return C.int(0)
	}
	lib.SetServiceOnCloseListener(&stListeners)
	lib.SetServiceProxyFailedFunc(stListeners.OnProxyFailed)
	return C.int(1)
}

//export ReloadFrpc
func ReloadFrpc() C.int {
	if err := lib.ReloadFrpc(); err != nil {
		LogPrint(err)
		return C.int(0)
	}
	lib.SetServiceOnCloseListener(&stListeners)
	lib.SetServiceProxyFailedFunc(stListeners.OnProxyFailed)
	return C.int(1)
}

//export SetReConnectByCount
func SetReConnectByCount(reConnectByCount bool) {
	lib.SetServiceReConnectByCount(reConnectByCount)
}

func init() {
	stListeners = LibStateListeners{}
}
