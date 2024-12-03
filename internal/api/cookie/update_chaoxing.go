package cookie

import (
	"KazeFrame/internal/cache"
	"KazeFrame/internal/config"
	"KazeFrame/internal/dao"
	"KazeFrame/internal/service"
	"KazeFrame/pkg/util"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// 超星续期任务接口
func ChaoxingUpdate(c *gin.Context) {
	paramType := c.Param("type")
	if paramType != "all" && paramType != "new" && paramType != "day" {
		util.Rsp(c, 400, "请指定正确的更新方式all/new/day")
		return
	}
	ctx := context.Background()
	redisClient := config.GetRedis()
	// 如果是新数据则不做redis检查每次都创建任务
	if paramType == "new" {
		cookieData, _ := dao.CkDataRepo.FindChaoxingTaskIds(paramType)
		service.ConcurrentRenewalCookie(paramType, cookieData)
	}
	// 检查Redis中是否存在正在进行的任务
	redisKey := fmt.Sprintf("%s%s:*", cache.ChaoxingTask, paramType)
	keys, err := redisClient.Keys(ctx, redisKey).Result()
	if err != nil {
		util.Rsp(c, 500, "获取Redis任务状态时发生错误: "+err.Error())
		return
	}
	// 如果存在任务，则统计总数和完成数
	if len(keys) > 0 {
		completedTaskCount := 0
		failedTaskCount := 0
		for _, key := range keys {
			status, err := redisClient.Get(ctx, key).Result()
			if err != nil && err != redis.Nil {
				util.Rsp(c, 500, "获取Redis状态时发生错误: "+err.Error())
				return
			}
			if status == "completed" {
				completedTaskCount++
			}
			if status == "failed" {
				failedTaskCount++
			}
		}
		bodyMessage := ""
		var bodyCode int
		if len(keys) == completedTaskCount || len(keys) == failedTaskCount+completedTaskCount {
			bodyCode = 204
			bodyMessage = "最近一小时内进行过超星续期任务"
		} else {
			bodyCode = 202
			bodyMessage = "超星续期任务进行中"
		}
		c.JSON(200, gin.H{
			"code":            bodyCode,
			"message":         bodyMessage,
			"task_count":      len(keys),
			"completed_count": completedTaskCount,
			"failed_count":    failedTaskCount,
		})
		return
	} else {
		// 如果不存在任务，则开始Cookie续期执行任务
		cookieData, err := dao.CkDataRepo.FindChaoxingTaskIds(paramType)
		if err != nil {
			util.Rsp(c, 500, "服务端缓存服务错误: "+err.Error())
			return
		}
		errors := service.ConcurrentRenewalCookie(paramType, cookieData)
		if errors != nil {
			util.Rsp(c, 200, "续期任务已开始, 但产生了错误: "+errors.Error())
			return
		} else {
			c.JSON(200, gin.H{"code": 200, "message": "续期任务已开始, 一小时内请勿重复操作"})
		}
	}
}
