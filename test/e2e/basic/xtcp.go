package basic

import (
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"

	"frp_lib/test/e2e/framework"
	"frp_lib/test/e2e/framework/consts"
	"frp_lib/test/e2e/pkg/port"
	"frp_lib/test/e2e/pkg/request"
)

var _ = ginkgo.Describe("[Feature: XTCP]", func() {
	f := framework.NewDefaultFramework()

	ginkgo.It("Fallback To STCP", func() {
		serverConf := consts.DefaultServerConfig
		clientConf := consts.DefaultClientConfig

		bindPortName := port.GenName("XTCP")
		clientConf += fmt.Sprintf(`
			[foo]
			type = stcp
			local_port = {{ .%s }}

			[foo-visitor]
			type = stcp
			role = visitor
			server_name = foo
			bind_port = -1

			[bar-visitor]
			type = xtcp
			role = visitor
			server_name = bar
			bind_port = {{ .%s }}
			keep_tunnel_open = true
			fallback_to = foo-visitor
			fallback_timeout_ms = 200
			`, framework.TCPEchoServerPort, bindPortName)

		f.RunProcesses([]string{serverConf}, []string{clientConf})
		framework.NewRequestExpect(f).
			RequestModify(func(r *request.Request) {
				r.Timeout(time.Second)
			}).
			PortName(bindPortName).
			Ensure()
	})
})
