package cookie

import (
	"KazeFrame/internal/dao"
	"KazeFrame/internal/model"
	"KazeFrame/pkg/util"

	"github.com/gin-gonic/gin"
)

// 执行两个数据库表删除操作，但只响应第一个操作的结果
func DeleteCookie(c *gin.Context) {
	var userDeletePayload model.DeletePayload
	if err := c.ShouldBindJSON(&userDeletePayload); err != nil {
		util.Rsp(c, 400, "请求参数错误: "+err.Error())
		return
	}
	// 删除ck_info表的数据
	response1, err := dao.CkInfoRepo.QuickHardDelete(userDeletePayload.Field, userDeletePayload.Value)
	// 删除ck_data表的数据
	dao.CkDataRepo.QuickHardDelete(userDeletePayload.Field, userDeletePayload.Value)
	if err != nil {
		util.Rsp(c, 500, "删除操作失败: "+err.Error())
		return
	}
	if response1.OkCount == 0 {
		util.Rsp(c, 404, "没找到你想删除的数据捏")
		return
	}
	c.JSON(200, response1)
}
