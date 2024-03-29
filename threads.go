package threads

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"

	"github.com/pkg/errors"

	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	getUserDocID        = "23996318473300828"
	getUserThreadsDocID = "7119605668149033"
	getUserRepliesDocID = "6307072669391286"
	getPostDocID        = "6992290264212558"
	getPostLikersDocID  = "9360915773983802"
)

const (
	apiURL   = "https://www.threads.net/api/graphql"
	tokenURL = "https://www.threads.net/@instagram"
)

// Threads implements Threads.net API wrapper.
type Threads struct {
	token          string
	defaultHeaders http.Header
	proxy          *url.URL
}

type ThreadsOptFun func(*Threads)

func WithProxy(proxy string) ThreadsOptFun {
	return func(t *Threads) {
		if proxy == "" {
			return
		}
		if proxy, e := url.Parse(proxy); e == nil {
			t.proxy = proxy
		}
	}
}

// NewThreads constructs a Threads instance.
func NewThreads(opt ...ThreadsOptFun) (t *Threads, err error) {
	t = new(Threads)
	for _, o := range opt {
		o(t)
	}

	t.token, err = t.getToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token")
	}

	t.defaultHeaders = map[string][]string{
		"Authority":       {"www.threads.net"},
		"Accept":          {"*/*"},
		"Accept-Language": {"en-US,en;q=0.9"},
		"Cache-Control":   {"no-cache"},
		"Content-Type":    {"application/x-www-form-urlencoded"},
		"Connection":      {"keep-alive"},
		"Origin":          {"https://www.threads.net"},
		"Pragma":          {"no-cache"},
		"Sec-Fetch-Site":  {"same-origin"},
		"User-Agent":      {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"},
		"X-Asbd-Id":       {"129477"},
		"XX-Ig-App-Id":    {"238260118697367"},
		"X-Fb-Lsd":        {t.token},
	}

	return t, nil
}

func (t *Threads) getToken() (string, error) {
	req, err := http.NewRequest(http.MethodGet, tokenURL, nil)
	if err != nil {
		return "", errors.Wrap(err, "http.NewRequest")
	}

	req.Header = map[string][]string{
		"User-Agent": {"golang"},
	}

	client := t.httpClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "http Get request failed")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	bodyStr := string(body)

	token := findJsonStringValue(bodyStr, "token")
	if token != "" {
		return token, nil
	}
	return "", errors.New("token not found")
}

func findJsonStringValue(bodyStr string, key string) string {
	KeyPos := strings.Index(bodyStr, `"`+key+`"`)
	if KeyPos == -1 {
		return ""
	}
	valueStr := strings.ReplaceAll(bodyStr[KeyPos:], " ", "")
	valueStr = strings.ReplaceAll(valueStr, "\n", "")
	valueStr = strings.ReplaceAll(valueStr, "\r", "")
	valueStr = strings.ReplaceAll(valueStr, "\t", "")
	re := regexp.MustCompile(`"` + key + `"` + `:"([^"]+)"`)
	// 查找匹配项
	match := re.FindStringSubmatch(valueStr)
	if len(match) > 1 {
		value := match[1]
		return value
	}
	return ""
}

func (t *Threads) postRequest(variables map[string]any, docID string, headers http.Header, attrDatas ...map[string]string) ([]byte, error) {
	variablesStr, err := json.Marshal(variables)
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("lsd", t.token)
	data.Set("variables", string(variablesStr))
	data.Set("doc_id", docID)
	for _, attrData := range attrDatas {
		for k, v := range attrData {
			data.Set(k, v)
		}
	}
	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header = headers

	client := t.httpClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// GetPost fetches a post.
func (t *Threads) GetPost(id int64) ([]byte, error) {
	variables := map[string]any{
		"postID": fmt.Sprintf("%d", id),
		"__relay_internal__pv__BarcelonaIsLoggedInrelayprovider":                               false,
		"__relay_internal__pv__BarcelonaIsThreadContextHeaderEnabledrelayprovider":             false,
		"__relay_internal__pv__BarcelonaIsThreadContextHeaderFollowButtonEnabledrelayprovider": false,
		"__relay_internal__pv__BarcelonaUseCometVideoPlaybackEnginerelayprovider":              false,
		"__relay_internal__pv__BarcelonaOptionalCookiesEnabledrelayprovider":                   false,
		"__relay_internal__pv__BarcelonaIsViewCountEnabledrelayprovider":                       false,
		"__relay_internal__pv__BarcelonaShouldShowFediverseM075Featuresrelayprovider":          false,
	}

	headers := t.defaultHeaders.Clone()
	headers.Add("X-Fb-Friendly-Name", "BarcelonaPostPageQuery")
	return t.postRequest(variables, getPostDocID, headers)
}

// GetPostByURL fetches a post.
func (t *Threads) GetPostByURL(u string) ([]byte, error) {
	return t.GetPost(GetPostIDfromURL(u))
}

// GetPostLikers fetches all users who liked the post.
func (t *Threads) GetPostLikers(id int64) ([]byte, error) {
	variables := map[string]any{"mediaID": id}

	return t.postRequest(variables, getPostLikersDocID, t.defaultHeaders)
}

// GetUser fetches a user.
func (t *Threads) GetUser(id int64) ([]byte, error) {
	variables := map[string]any{"userID": id}

	headers := t.defaultHeaders.Clone()
	headers.Add("X-Fb-Friendly-Name", "BarcelonaProfileRootQuery")

	return t.postRequest(variables, getUserDocID, headers)
}

// GetUserThreads fetches a user's Threads.
func (t *Threads) GetUserThreads(id int64) ([]byte, error) {
	variables := map[string]any{"userID": id}

	headers := t.defaultHeaders.Clone()
	headers.Add("X-Fb-Friendly-Name", "BarcelonaProfileThreadsTabQuery")

	return t.postRequest(variables, getUserThreadsDocID, headers)
}

// GetUserReplies fetches a user's replies.
func (t *Threads) GetUserReplies(id int64) ([]byte, error) {
	variables := map[string]any{"userID": id}

	headers := t.defaultHeaders.Clone()
	headers.Add("X-Fb-Friendly-Name", "BarcelonaProfileRepliesTabQuery")

	return t.postRequest(variables, getUserRepliesDocID, headers)
}

// GetUserID fetches user's ID by username.
func (t *Threads) GetUserID(username string) (int64, error) {
	baseURL := fmt.Sprintf("https://www.threads.net/@%s", username)

	req, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		return -1, err
	}

	req.Header = t.defaultHeaders.Clone()

	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("Referer", baseURL)
	req.Header.Add("Sec-Fetch-Dest", "document")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Upgrade-Insecure-Requests", "1")

	req.Header.Del("X-Asbd-Id")
	req.Header.Del("X-Fb-Lsd")
	req.Header.Del("XX-Ig-App-Id")

	client := t.httpClient()
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}

	userIdRegex := regexp.MustCompile(`"user_id":"(\d+)"`)
	userIdStr := userIdRegex.FindStringSubmatch(string(body))[1]

	userId, err := strconv.ParseInt(userIdStr, 0, 63)
	if err != nil {
		return -1, err
	}

	return userId, nil
}

func (t *Threads) httpClient() *http.Client {
	if t.proxy == nil {
		return &http.Client{}
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(t.proxy),
	}
	return &http.Client{
		Transport: transport,
	}
}
