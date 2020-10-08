package vmess

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/qjebbs/v2tool/files"
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
	filesMap, err := getFilesMap(outdir)
	if err != nil {
		return err
	}
	writeFile := func(filename string, data []byte) error {
		if file, ok := filesMap[filename]; ok {
			// file exist
			rel, err := filepath.Rel(outdir, file)
			if err != nil {
				return err
			}
			hasher := md5.New()
			s, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}
			hasher.Write(s)
			fileMD5 := hex.EncodeToString(hasher.Sum(nil))
			hasher.Reset()
			hasher.Write(data)
			dataMD5 := hex.EncodeToString(hasher.Sum(nil))
			if fileMD5 != dataMD5 {
				fmt.Println("Updated:", rel)
				err = ioutil.WriteFile(file, data, 0644)
				if err != nil {
					return err
				}
			}
			delete(filesMap, filename)
			return nil
		}
		// file not exist
		file := filepath.Join(outdir, filename)
		fmt.Println("Added:", filename)
		return ioutil.WriteFile(file, data, 0644)

	}
	asFileName := func(ps string) string {
		reg := regexp.MustCompile(`([\\/:*?"<>|]|\s)+`)
		r := reg.ReplaceAll([]byte(ps), []byte(" "))
		return strings.TrimSpace(string(r))
	}
	subscriptionToJSONs := func(sub *Subscription) error {
		fmt.Println(sub)
		fmt.Println("Output:", outdir)
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
			filename := out.Tag + ".json"
			content, err := outbound2JSON(out, socketMark)
			if err != nil {
				return err
			}
			err = writeFile(filename, content)
			if err != nil {
				return err
			}
		}
		return nil
	}
	subscriptionsToJSONs := func(subs []*Subscription) error {
		for _, sub := range subs {
			err := subscriptionToJSONs(sub)
			if err != nil {
				return err
			}
		}
		return nil
	}

	c := &SubscriptionConfig{}
	data, err := ioutil.ReadFile(conf)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, c)
	if err != nil {
		return err
	}
	err = subscriptionsToJSONs(c.Subscriptions)
	if err != nil {
		return err
	}
	for _, file := range filesMap {
		rel, err := filepath.Rel(outdir, file)
		if err != nil {
			return err
		}
		fmt.Println("Removed:", rel)
	}
	return nil
}

func getFilesMap(dir string) (map[string]string, error) {
	files, err := files.GetFolderFiles(dir)
	if err != nil {
		return nil, err
	}
	filesMap := make(map[string]string)
	for _, f := range files {
		filesMap[filepath.Base(f)] = f
	}
	return filesMap, nil
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
