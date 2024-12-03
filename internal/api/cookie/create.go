package cookie

import (
	"KazeFrame/internal/config"
	"KazeFrame/internal/dao"
	"KazeFrame/internal/model"
	"KazeFrame/pkg/util"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ReqPayload struct {
	Cookies      json.RawMessage `json:"cookies"`
	Url          string          `json:"url" comment:"网址"`
	UrlTitle     string          `json:"url_title" comment:"网址标题"`
	Ip           string          `json:"ip" comment:"IP地址"`
	UpdateStatus string          `json:"update_status" comment:"续期状态"`
	ChaoxingId   string          `json:"chaoxing_name" comment:"超星用户名"`
	Tag          string          `json:"tag" comment:"Cookie自定义标识"`
}

type CookieAnatomy struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

var DomainTypeMapping = map[string]string{
	"chaoxing.com": "超星",
	"bilibili.com": "B站",
	"bing.com":     "必应",
	"baidu.com":    "百度",
	"so.com":       "360",
	"localhost":    "本地",
}

// 检测url中是否包含了关键词, 进行分类
func GetTypeFromURL(url string) string {
	for domain, typeName := range DomainTypeMapping {
		if strings.Contains(url, domain) {
			return typeName
		}
	}
	return "未知分类"
}

// 触发了"超星"类型，并且前端没有传来chaoxing_name，那么尝试从cookies字段中的内容中，提取UID作为chaoxing_id
func extractChaoxingUID(cookies json.RawMessage) (string, error) {
	var cookieList []CookieAnatomy
	if err := json.Unmarshal(cookies, &cookieList); err != nil {
		return "", err
	}
	for _, cookie := range cookieList {
		if strings.EqualFold(cookie.Name, "UID") {
			return cookie.Value, nil
		}
	}
	return "", errors.New("在Cookie中找不到“UID”")
}

// 储存数据接口
func SetCookie(c *gin.Context) {
	reqData := ReqPayload{}
	if err := c.ShouldBindJSON(&reqData); err != nil {
		util.Rsp(c, 400, "请求参数或值错误")
		return
	}
	// 检查Type是否为"超星"并且chaoxing_name为空
	if GetTypeFromURL(reqData.Url) == "超星" && reqData.ChaoxingId == "" {
		uid, err := extractChaoxingUID(reqData.Cookies)
		if err != nil {
			return
		}
		reqData.ChaoxingId = uid
	}
	// 开始事务
	tx := dao.CkInfoRepo.DB.Begin()
	// 创建CkInfo记录
	infoData := model.CkInfo{
		Url:        reqData.Url,
		UrlTitle:   reqData.UrlTitle,
		Ip:         c.ClientIP(),
		ChaoxingId: reqData.ChaoxingId,
		Type:       GetTypeFromURL(reqData.Url),
		Tag:        reqData.Tag,
	}
	// 创建infoData并获取其ID
	if err := tx.Create(&infoData).Error; err != nil {
		tx.Rollback()
		util.Rsp(c, 500, "创建失败")
		return
	}
	// 创建CkData记录，确保其ID与infoData的ID匹配
	ckData := model.CkData{
		ID:   infoData.ID,
		Data: reqData.Cookies,
		Type: GetTypeFromURL(reqData.Url),
	}
	// 创建CkData记录
	if err := tx.Create(&ckData).Error; err != nil {
		tx.Rollback()
		util.Rsp(c, 500, "创建失败")
		return
	}
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		util.Rsp(c, 500, "创建失败")
		return
	}
	util.Rsp(c, 200, "创建成功")
	// 请求一次超星数据续期接口
	renewalNewApi := fmt.Sprintf("http://localhost:%s/cookie/update/cx/new", config.GetConfig().Server.Port)
	http.Get(renewalNewApi)
}
