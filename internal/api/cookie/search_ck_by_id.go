package cookie

import (
	"KazeFrame/internal/dao"
	"KazeFrame/pkg/util"

	"github.com/gin-gonic/gin"
)

// 通过URL中的ID精准获取数据接口
func SearchCookieById(c *gin.Context) {
	paramID := c.Param("id")
	if paramID == "" {
		util.Rsp(c, 400, "查询目标id不能为空")
		return
	}
	cookieData, err := dao.CkDataRepo.FindByFieldExact("id", paramID)
	if err != nil {
		util.Rsp(c, 500, "查询用户数据时发生错误: "+err.Error())
		return
	}
	if len(cookieData) == 0 {
		util.Rsp(c, 404, "查询不到目标数据")
		return
	}
	c.JSON(200, cookieData[0].Data)
}
