package vmess

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"v2ray.com/core/infra/conf"
)

// Link2JSON converts vmess link to json string
func Link2JSON(v string) ([]byte, error) {
	var lk *VmessLink
	var err error
	lk, err = ParseVmess(v)
	out, err := Link2Outbound(lk, false)
	if err != nil {
		return nil, err
	}
	type outConfig struct {
		OutboundConfigs []conf.OutboundDetourConfig
	}
	return json.Marshal(outConfig{
		OutboundConfigs: []conf.OutboundDetourConfig{
			*out,
		},
	})
}

// Link2Outbound converts vmess link to *OutboundDetourConfig
func Link2Outbound(v *VmessLink, usemux bool) (*conf.OutboundDetourConfig, error) {
	out := &conf.OutboundDetourConfig{}
	out.Tag = asFileName(v.Ps)
	out.Protocol = "vmess"
	out.MuxSettings = &conf.MuxConfig{}
	if usemux {
		out.MuxSettings.Enabled = true
		out.MuxSettings.Concurrency = 8
	}

	p := conf.TransportProtocol(v.Net)
	s := &conf.StreamConfig{
		Network:  &p,
		Security: v.TLS,
	}

	switch v.Net {
	case "tcp":
		s.TCPSettings = &conf.TCPConfig{}
		if v.Type == "" || v.Type == "none" {
			s.TCPSettings.HeaderConfig = json.RawMessage([]byte(`{ "type": "none" }`))
		} else {
			pathb, _ := json.Marshal(strings.Split(v.Path, ","))
			hostb, _ := json.Marshal(strings.Split(v.Host, ","))
			s.TCPSettings.HeaderConfig = json.RawMessage([]byte(fmt.Sprintf(`
			{
				"type": "http",
				"request": {
					"path": %s,
					"headers": {
						"Host": %s
					}
				}
			}
			`, string(pathb), string(hostb))))
		}
	case "kcp":
		s.KCPSettings = &conf.KCPConfig{}
		s.KCPSettings.HeaderConfig = json.RawMessage([]byte(fmt.Sprintf(`{ "type": "%s" }`, v.Type)))
	case "ws":
		s.WSSettings = &conf.WebSocketConfig{}
		s.WSSettings.Path = v.Path
		s.WSSettings.Headers = map[string]string{
			"Host": v.Host,
		}
	case "h2", "http":
		s.HTTPSettings = &conf.HTTPConfig{
			Path: v.Path,
		}
		if v.Host != "" {
			h := conf.StringList(strings.Split(v.Host, ","))
			s.HTTPSettings.Host = &h
		}
	}

	if v.TLS == "tls" {
		s.TLSSettings = &conf.TLSConfig{
			Insecure: true,
		}
		if v.Host != "" {
			s.TLSSettings.ServerName = v.Host
		}
	}

	out.StreamSetting = s
	oset := json.RawMessage([]byte(fmt.Sprintf(`{
  "vnext": [
    {
      "address": "%s",
      "port": %v,
      "users": [
        {
          "id": "%s",
          "alterId": %v,
          "security": "auto"
        }
      ]
    }
  ]
}`, v.Add, v.Port, v.ID, v.Aid)))
	out.Settings = &oset
	return out, nil
}

func asFileName(ps string) string {
	reg := regexp.MustCompile(`([\\/:*?"<>|]|\s)+`)
	r := reg.ReplaceAll([]byte(ps), []byte(" "))
	return strings.Trim(string(r), " ")
}
