// Copyright 2018 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lib

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"frp_lib/client"
	"frp_lib/pkg/config"
	"frp_lib/pkg/util/log"
	"github.com/fatedier/golib/crypto"
)

var service *client.Service
var cfgPath string

func RunFrpc(cfgFilePath string) (err error) {
	if IsFrpRunning() {			
		return fmt.Errorf("frp already started")
	}
	crypto.DefaultSalt = "frp"
	return runClient(cfgFilePath)
}

func StopFrp() (err error) {
	if !IsFrpRunning() {
		return fmt.Errorf("frp not started")
	}

	service.Close()
	log.Info("frpc is stoped")
	service = nil
	return
}

func IsFrpRunning() bool {
	return service != nil && !service.IsClosed()
}

func ReloadFrpc() (err error) {
	if !IsFrpRunning() {
		return fmt.Errorf("frp not started")
	}
	_, pxyCfgs, visitorCfgs, err := config.ParseClientConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("reload frpc proxy config error: %s", err.Error())
	}

	if err = service.ReloadConf(pxyCfgs, visitorCfgs); err != nil {
		return fmt.Errorf("reload frpc proxy config error: %s", err.Error())
	}
	log.Info("success reload conf")
	return nil;
}

func SetServiceProxyFailedFunc(proxyFailedFunc func(err error)) {
	if service != nil {
		service.SetProxyFailedFunc(proxyFailedFunc)
	}
}

type ServiceClosedListener interface {
	OnClosed(msg string)
}

func SetServiceOnCloseListener(listener ServiceClosedListener) {
	if service != nil {
		service.SetOnCloseListener(listener)
	}
}

func SetServiceReConnectByCount(reConnectByCount bool) {
	if service != nil {
		service.ReConnectByCount = reConnectByCount
	}
}

func handleTermSignal(svr *client.Service) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	svr.GracefulClose(500 * time.Millisecond)
}

func runClient(cfgFilePath string) error {
	cfg, pxyCfgs, visitorCfgs, err := config.ParseClientConfig(cfgFilePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return startService(cfg, pxyCfgs, visitorCfgs, cfgFilePath)
}

func startService(
	cfg config.ClientCommonConf,
	pxyCfgs map[string]config.ProxyConf,
	visitorCfgs map[string]config.VisitorConf,
	cfgFile string,
) (err error) {
	log.InitLog(cfg.LogWay, cfg.LogFile, cfg.LogLevel,
		cfg.LogMaxDays, cfg.DisableLogColor)

	if cfgFile != "" {
		log.Info("start frpc service for config file [%s]", cfgFile)
		defer log.Info("frpc service for config file [%s] stopped", cfgFile)
	}
	svr, errRet := client.NewService(cfg, pxyCfgs, visitorCfgs, cfgFile)
	if errRet != nil {
		err = errRet
		return
	}
	service = svr
	cfgPath = cfgFile

	shouldGracefulClose := cfg.Protocol == "kcp" || cfg.Protocol == "quic"
	// Capture the exit signal if we use kcp or quic.
	if shouldGracefulClose {
		go handleTermSignal(svr)
	}

	_ = svr.Run(context.Background(), false)
	return
}
