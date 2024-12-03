package cookie

import (
	"KazeFrame/internal/dao"
	"KazeFrame/pkg/util"

	"github.com/gin-gonic/gin"
)

// 获取超星用户“续期中”的Cookie
func SearchChaoxing(c *gin.Context) {
	paramID := c.Param("chaoxing_id")
	if paramID == "" {
		util.Rsp(c, 400, "超星ID不能为空")
		return
	}
	chaoxingUserList, err := dao.CkInfoRepo.FindChaoxingList(paramID)
	if err != nil {
		util.Rsp(c, 500, err.Error())
		return
	}
	renewableData, _ := dao.CkDataRepo.FindRenewableDataByID(chaoxingUserList)
	if len(chaoxingUserList) == 0 || len(renewableData) == 0 {
		c.JSON(404, gin.H{"chaoxing_id": paramID, "message": "该用户不存在处于“续期中”的Cookie"})
	} else {
		c.JSON(200, renewableData)
	}
}
