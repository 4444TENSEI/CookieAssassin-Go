package service

import (
	"KazeFrame/internal/cache"
	"KazeFrame/internal/config"
	"KazeFrame/internal/dao"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	url1 = "https://sso.chaoxing.com/apis/login/userLogin4Uname.do"
	url2 = "https://sso.chaoxing.com/apis/login/userLogin.do"
)

// 更新后的cookie字段
type ExtendedCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Expires  time.Time `json:"expires,omitempty"`
	Path     string    `json:"path"`
	Domain   string    `json:"domain"`
	HttpOnly bool      `json:"httpOnly"`
	Secure   bool      `json:"secure"`
	Session  bool      `json:"session"`
}

// 将Cookie从json转header形式
func buildCookieHeader(cookies []ExtendedCookie) string {
	var cookieHeader strings.Builder
	for _, cookie := range cookies {
		cookieHeader.WriteString(fmt.Sprintf("%s=%s; ", cookie.Name, cookie.Value))
	}
	return cookieHeader.String()
}

// 解析响应头的Set-Cookie
func parseSetCookieHeader(header string) ExtendedCookie {
	var cookie ExtendedCookie
	parts := strings.Split(header, "; ")
	for _, part := range parts {
		if strings.Contains(part, "=") {
			cookieParts := strings.SplitN(part, "=", 2)
			switch cookieParts[0] {
			case "Domain":
				cookie.Domain = cookieParts[1]
			case "Expires":
				expireTime, err := time.Parse(time.RFC1123, cookieParts[1])
				if err == nil {
					cookie.Expires = expireTime
					cookie.Session = false
				}
			case "Path":
				cookie.Path = cookieParts[1]
			default:
				cookie.Name = cookieParts[0]
				cookie.Value = cookieParts[1]
			}
		} else {
			switch part {
			case "HttpOnly":
				cookie.HttpOnly = true
			case "Secure":
				cookie.Secure = true
			}
		}
	}
	return cookie
}

// 解析响应体并检查result键值对
func checkRenewalStatus(res *http.Response) (bool, error) {
	if res.Body == nil {
		return false, fmt.Errorf("response body is nil")
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}
	if result["result"] == float64(1) { // 注意：JSON unmarshal会将数字转换为float64
		return true, nil
	}
	return false, nil
}

// 使用请求头cookie发送请求，并检查是否可续期
func sendReq(url, basicCookie string, cookies []ExtendedCookie) ([]ExtendedCookie, bool, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Add("Cookie", basicCookie)
	req.Header.Add("Host", "sso.chaoxing.com")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	res, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()
	setCookieHeaders := res.Header["Set-Cookie"]
	for _, header := range setCookieHeaders {
		cookie := parseSetCookieHeader(header)
		found := false
		for i, c := range cookies {
			if c.Name == cookie.Name {
				cookies[i] = cookie
				found = true
				break
			}
		}
		if !found {
			cookies = append(cookies, cookie)
		}
	}
	renewalStatus, err := checkRenewalStatus(res)
	if err != nil {
		return nil, false, err
	}
	return cookies, renewalStatus, nil
}

// 传递一个包含ID和数据的切片，并发执行请求，保存数据到数据库
func ConcurrentRenewalCookie(paramType string, dataList []dao.DataWithID) error {
	redisClient := config.GetRedis()
	ctx := context.Background()
	var wg sync.WaitGroup
	errChan := make(chan error, len(dataList))
	// 数据库连接并发数20
	// 创建一个带缓冲的通道，限制数据库连接数
	semaphore := make(chan struct{}, 20)
	for _, data := range dataList {
		wg.Add(1)
		go func(id uint, data json.RawMessage) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量
			redisKey := fmt.Sprintf("%s%s:%d", cache.ChaoxingTask, paramType, id)
			if redisClient.Exists(ctx, redisKey).Val() > 0 {
				return
			}
			var cookies []ExtendedCookie
			if err := json.Unmarshal(data, &cookies); err != nil {
				updateList := map[string]interface{}{
					"update_status": "无法解析",
				}
				if updateErr := dao.CkDataRepo.UpdateByField("id", id, updateList); updateErr != nil {
					errChan <- fmt.Errorf("error updating status for id %d after unmarshalling error: %v", id, updateErr)
				} else {
					errChan <- fmt.Errorf("error unmarshalling data for id %d: %v", id, err)
				}
				redisClient.Set(ctx, redisKey, "failed", time.Hour)
				return
			}
			basicCookieStr := buildCookieHeader(cookies)
			urls := []string{url1, url2, url1}
			var lastRenewalStatus bool
			for _, url := range urls {
				var err error
				cookies, lastRenewalStatus, err = sendReq(url, basicCookieStr, cookies)
				if err != nil {
					errChan <- fmt.Errorf("error sending request for id %d: %v", id, err)
					redisClient.Set(ctx, redisKey, "failed", time.Hour)
					return
				}
				if lastRenewalStatus {
					break
				}
			}
			if lastRenewalStatus {
				cookiesJSON, _ := json.Marshal(cookies)
				updateList := map[string]interface{}{
					"data":          string(cookiesJSON),
					"update_status": "续期中",
				}
				if err := dao.CkDataRepo.UpdateByField("id", id, updateList); err != nil {
					errChan <- fmt.Errorf("error updating cookies for id %d: %v", id, err)
					redisClient.Set(ctx, redisKey, "failed", time.Hour)
					return
				}
			} else {
				updateList := map[string]interface{}{
					"update_status": "不可续期",
				}
				if err := dao.CkDataRepo.UpdateByField("id", id, updateList); err != nil {
					errChan <- fmt.Errorf("error updating status for id %d: %v", id, err)
					redisClient.Set(ctx, redisKey, "failed", time.Hour)
					return
				}
			}
			redisClient.Set(ctx, redisKey, "completed", time.Hour)
		}(data.ID, data.Data)
	}
	wg.Wait()
	close(errChan)
	var chanErr error
	for err := range errChan {
		if chanErr == nil {
			chanErr = err
		}
	}
	return chanErr
}
