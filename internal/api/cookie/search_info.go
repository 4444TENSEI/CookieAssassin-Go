package cookie

import (
	"KazeFrame/internal/dao"
	"KazeFrame/pkg/util"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 默认分页参数
const (
	defaultPage  = 1
	defaultLimit = 9999
)

// 获取所有用户信息、在线状态接口
func SearchInfoData(c *gin.Context) {
	// 获取page参数，如果为空或不是数字，则使用默认值defaultPage
	pageStr := c.Query("page")
	if pageStr == "" {
		pageStr = strconv.Itoa(defaultPage)
	}
	// 获取limit参数，如果为空或不是数字，则使用默认值defaultLimit
	pageSizeStr := c.Query("limit")
	if pageSizeStr == "" {
		pageSizeStr = strconv.Itoa(defaultLimit)
	}
	// 转换为数值用于计算总页数
	curPage, _ := strconv.Atoi(pageStr)
	if curPage < 1 {
		curPage = defaultPage
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 {
		pageSize = defaultLimit
	}
	// 查询
	allData, dataCount, err := dao.CkInfoRepo.FindTableData(curPage, pageSize)
	if err != nil {
		util.Rsp(c, 500, "数据查询失败, "+err.Error())
		return
	}
	// 计算总页数
	pageSizeInt64 := int64(pageSize)
	pageCount := dataCount / pageSizeInt64
	if dataCount%pageSizeInt64 != 0 {
		pageCount++
	}
	c.JSON(200, gin.H{
		"data_count": dataCount,
		"data":       allData,
		"page":       curPage,
		"limit":      pageSize,
		"page_count": pageCount,
	})
}
