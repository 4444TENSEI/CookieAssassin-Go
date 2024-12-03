package cookie

import (
	"KazeFrame/internal/dao"
	"KazeFrame/pkg/util"

	"github.com/gin-gonic/gin"
)

// 通过URL中的用户ID精准获取用户资料接口
func SearchChaoxing(c *gin.Context) {
	paramID := c.Param("chaoxing_id")
	if paramID == "" {
		util.Rsp(c, 400, "超星ID不能为空")
		return
	}
	renewableData, overdueData, err := dao.CkInfoRepo.FindChaoxingList(paramID)
	if err != nil {
		util.Rsp(c, 500, err.Error())
		return
	}
	if len(renewableData) == 0 && len(overdueData) == 0 {
		c.JSON(404, gin.H{"chaoxing_id": paramID, "message": "未找到该超星用户相关数据"})
	} else {
		c.JSON(200, gin.H{
			"renewable_data":  renewableData,
			"renewable_count": len(renewableData),
			"overdue_data":    overdueData,
			"overdue_count":   len(overdueData),
		})
	}
}
