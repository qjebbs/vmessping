package miniv2ray

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"context"
	"errors"
	"net"
	"net/http"
	"net/url"

	"github.com/qjebbs/v2tool/vmess"
	"v2ray.com/core"
	"v2ray.com/core/app/dispatcher"
	applog "v2ray.com/core/app/log"
	"v2ray.com/core/app/proxyman"
	commlog "v2ray.com/core/common/log"
	v2net "v2ray.com/core/common/net"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/infra/conf"
)

func JSON2Outbound(f string, usemux bool) (*core.OutboundHandlerConfig, error) {
	c := &conf.Config{}
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(data, c)
	if err != nil {
		return nil, err
	}
	if c.OutboundConfigs == nil || len(c.OutboundConfigs) == 0 {
		return nil, fmt.Errorf("no valid outbound found in %s", f)
	}
	out := c.OutboundConfigs[0]
	out.Tag = "proxy"
	out.MuxSettings = &conf.MuxConfig{}
	if usemux {
		out.MuxSettings.Enabled = true
		out.MuxSettings.Concurrency = 8
	}
	return out.Build()
}

func Vmess2Outbound(v *vmess.VmessLink, usemux bool) (*core.OutboundHandlerConfig, error) {
	out, err := vmess.Link2Outbound(v, usemux)
	out.Tag = "proxy"
	if err != nil {
		return nil, err
	}
	return out.Build()
}

func StartV2Ray(vm string, verbose, usemux bool) (*core.Instance, error) {

	loglevel := commlog.Severity_Error
	if verbose {
		loglevel = commlog.Severity_Debug
	}
	var (
		ob  *core.OutboundHandlerConfig
		err error
	)
	var u *url.URL
	if u, err = url.Parse(vm); err == nil && u.Scheme != "" {
		lk, err := vmess.ParseVmess(vm)
		if err != nil {
			return nil, err
		}

		fmt.Println("\n" + lk.DetailStr())
		ob, err = Vmess2Outbound(lk, usemux)
	} else {
		ob, err = JSON2Outbound(vm, usemux)
	}
	if err != nil {
		return nil, err
	}

	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&applog.Config{
				ErrorLogType:  applog.LogType_Console,
				ErrorLogLevel: loglevel,
			}),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		},
	}

	commlog.RegisterHandler(commlog.NewLogger(commlog.CreateStderrLogWriter()))
	config.Outbound = []*core.OutboundHandlerConfig{ob}
	server, err := core.New(config)
	if err != nil {
		return nil, err
	}

	return server, nil
}

func MeasureDelay(inst *core.Instance, timeout time.Duration, dest string) (int64, error) {
	start := time.Now()
	code, _, err := CoreHTTPRequest(inst, timeout, "GET", dest)
	if err != nil {
		return -1, err
	}
	if code > 399 {
		return -1, fmt.Errorf("status incorrect (>= 400): %d", code)
	}
	return time.Since(start).Milliseconds(), nil
}

func CoreHTTPClient(inst *core.Instance, timeout time.Duration) (*http.Client, error) {

	if inst == nil {
		return nil, errors.New("core instance nil")
	}

	tr := &http.Transport{
		DisableKeepAlives: true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dest, err := v2net.ParseDestination(fmt.Sprintf("%s:%s", network, addr))
			if err != nil {
				return nil, err
			}
			return core.Dial(ctx, inst, dest)
		},
	}

	c := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	return c, nil
}

func CoreHTTPRequest(inst *core.Instance, timeout time.Duration, method, dest string) (int, []byte, error) {

	c, err := CoreHTTPClient(inst, timeout)
	if err != nil {
		return 0, nil, err
	}

	req, _ := http.NewRequest(method, dest, nil)
	resp, err := c.Do(req)
	if err != nil {
		return -1, nil, err
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, b, nil
}

func CoreVersion() string {
	return core.Version()
}
