package vmess

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"v2ray.com/core/infra/conf"
)

// Subscription represents a subscription config
type Subscription struct {
	Tag    string `json:"tag"`
	URL    string `json:"url"`
	Ignore string `json:"ignore"`
	Match  string `json:"match"`
}

// SubscriptionConfig represents a subscription json
type SubscriptionConfig struct {
	Subscriptions []*Subscription `json:"subscriptions"`
}

func (s *Subscription) String() string {
	return fmt.Sprintf(`Tag: %s
URL: %s
Ignore: %s
Match: %s`,
		s.Tag, s.URL, s.Ignore, s.Match)
}

// FetchSubscriptions fetches subscription specified by "conf", and generating json files to "outdir"
func FetchSubscriptions(conf string, outdir string, socketMark int32) error {
	c := &SubscriptionConfig{}
	data, err := ioutil.ReadFile(conf)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, c)
	if err != nil {
		return err
	}
	return SubscriptionsToJSONs(c.Subscriptions, outdir, socketMark)
}

// SubscriptionsToJSONs fetch multiple subscriptions and generating json files
func SubscriptionsToJSONs(subs []*Subscription, dir string, socketMark int32) error {
	for _, sub := range subs {
		err := SubscriptionToJSONs(sub, dir, socketMark)
		if err != nil {
			return err
		}
	}
	return nil
}

// SubscriptionToJSONs fetch subscription and generating json files
func SubscriptionToJSONs(sub *Subscription, dir string, socketMark int32) error {
	fmt.Println(sub)
	fmt.Println("Output:", dir)
	if socketMark != 0 {
		fmt.Println("Sokect mark:", socketMark)
	}
	fmt.Println("Downloading...")
	links, err := LinksFromSubscription(sub.URL)
	if err != nil {
		return err
	}
	fmt.Printf("%v link(s) found...\n", len(links))

	links, err = filterLinks(links, sub.Ignore, sub.Match)
	if err != nil {
		return err
	}
	for _, link := range links {
		out, err := Link2Outbound(link, false)
		if err != nil {
			return err
		}
		out.Tag = asFileName(sub.Tag + " - " + link.Ps)
		file := path.Join(dir, out.Tag+".json")
		content, err := outbound2JSON(out, socketMark)
		if err != nil {
			return err
		}
		err = writeFile(file, content)
		if err != nil {
			return err
		}
	}
	return nil
}

// outbound2JSON converts vmess link to json string
func outbound2JSON(out *conf.OutboundDetourConfig, socketMark int32) ([]byte, error) {
	if socketMark != 0 {
		if out.StreamSetting == nil {
			out.StreamSetting = &conf.StreamConfig{}
		}
		out.StreamSetting.SocketSettings = &conf.SocketConfig{
			Mark: socketMark,
		}
	}
	type outConfig struct {
		OutboundConfigs []conf.OutboundDetourConfig `json:"outbounds"`
	}
	return json.Marshal(outConfig{
		OutboundConfigs: []conf.OutboundDetourConfig{
			*out,
		},
	})
}

func asFileName(ps string) string {
	reg := regexp.MustCompile(`([\\/:*?"<>|]|\s)+`)
	r := reg.ReplaceAll([]byte(ps), []byte(" "))
	return strings.TrimSpace(string(r))
}

func writeFile(filename string, data []byte) error {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Added:", path.Base(filename))
			return ioutil.WriteFile(filename, data, 0644)
		}
		return err
	}
	// file exist
	hasher := md5.New()
	s, err := ioutil.ReadFile(filename)
	hasher.Write(s)
	fileMD5 := hex.EncodeToString(hasher.Sum(nil))
	hasher.Reset()
	hasher.Write(data)
	dataMD5 := hex.EncodeToString(hasher.Sum(nil))
	if fileMD5 != dataMD5 {
		fmt.Println("Updated:", path.Base(filename))
		return ioutil.WriteFile(filename, data, 0644)
	}
	return nil
}

// LinksFromSubscription downloads and parses links from a subscription URL
func LinksFromSubscription(url string) ([]*Link, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	decoded, err := base64Decode(string(body))
	if err != nil {
		return nil, err
	}
	content := string(decoded)
	links := make([]*Link, 0)
	for _, line := range strings.Split(content, "\n") {
		line = strings.Trim(line, " ")
		if !strings.HasPrefix(line, "vmess://") {
			continue
		}
		link, err := ParseVmess(line)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func filterLinks(links []*Link, exclude string, include string) ([]*Link, error) {
	lks := make([]*Link, 0)
	var (
		err        error
		regExclude *regexp.Regexp
		regInclude *regexp.Regexp
	)
	if exclude != "" {
		regExclude, err = regexp.Compile(exclude)
		if err != nil {
			return nil, err
		}
	}
	if include != "" {
		regInclude, err = regexp.Compile(include)
		if err != nil {
			return nil, err
		}
	}
	for _, l := range links {
		if regExclude != nil && regExclude.Match([]byte(l.Ps)) {
			fmt.Printf("Ignored: %s\n", l.Ps)
			continue
		}
		if regInclude != nil && !regInclude.Match([]byte(l.Ps)) {
			continue
		}
		lks = append(lks, l)
	}
	return lks, nil
}
