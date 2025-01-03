// 接口路由
package router

import (
	"KazeFrame/internal/api/cookie"
	"KazeFrame/internal/api/email"
	"KazeFrame/internal/api/monitor"
	"KazeFrame/internal/api/user"
	"KazeFrame/internal/middleware"
	"KazeFrame/pkg/util"
	"time"

	"github.com/gin-gonic/gin"
)

func RunServer() *gin.Engine {
	// 使用gin默认配置
	r := gin.Default()
	// 跨域中间件
	r.Use(middleware.CORS())
	// 请求/访问日志记录中间件
	r.Use(middleware.LoggerMiddleware())
	// 频繁访问限制中间件, 需要解除initRedis注释，制定同一IP在一定时间内仅允许一定数量的请求, 例如这里限制所有路由60分钟内200次请求
	// r.Use(middleware.IPRateLimiter(200, 60*time.Minute))
	// 公开可访问的静态资源目录, 不能定义为跟路由因为和其他路由会冲突, 故这里定义为/ui
	r.Static("/ui", "./static/ui")
	// 跟路由302重定向到首页
	r.GET("/doc", func(c *gin.Context) {
		c.Redirect(302, "https://apifox.com/apidoc/shared-28e22c5c-ee2f-41eb-8be0-a77f618ca03b")
	})
	// 404响应
	r.NoRoute(func(c *gin.Context) {
		util.Rsp(c, 404, "你跑丢了~")
	})
	cookieAPI := r.Group("/cookie")
	{
		// 存数据
		cookieAPI.POST("/set", cookie.SetCookie)
		// 执行超星续期任务
		cookieAPI.GET("/update/cx/:type", cookie.ChaoxingUpdate)
		// 获取基础Cookie列表内容
		cookieAPI.GET("/search/info", middleware.ParamAuth(), cookie.SearchInfoData)
		// 根据id精准查找对应的cookie内容
		cookieAPI.GET("/search/id/:id", middleware.ParamAuth(), cookie.SearchCookieById)
		// 根据chaoxing_id精准查找对应的cookie内容
		cookieAPI.GET("/search/cx/:chaoxing_id", middleware.ParamAuth(), cookie.SearchChaoxing)
		cookieAPI.DELETE("/delete", middleware.ParamAuth(), cookie.DeleteCookie)
	}
	// 用户服务接口 注意加RoleAuth中间件不然获取不到Token中的用户ID, 目的是防止越权
	// 同时登录后Cookie中还会额外有一个明文的用户ID, 但是请慎用因为不做校验有可能会被客户端修改
	userAPI := r.Group("/user")
	{
		// 登录
		userAPI.POST("/login", user.UserLogin)
		// 发送注册验证码邮件
		userAPI.POST("/captcha/register", middleware.IPRateLimiter(4, 60*time.Minute), email.SendRegisterCaptcha)
		// 注册
		userAPI.POST("/register", middleware.IPRateLimiter(20, 60*time.Minute), user.UserRegisterPayload)
		// 发送找回密码邮件
		userAPI.POST("/captcha/forget", middleware.IPRateLimiter(4, 60*time.Minute), email.SendForgetPswCaptcha)
		// 找回密码
		userAPI.POST("/forget", middleware.IPRateLimiter(20, 60*time.Minute), user.UserForgetPassword)
		// 退出登录, 并清除redis登录状态
		userAPI.GET("/logout", middleware.RoleAuth(2), user.UserLogout)
		// 维持登录状态, 使用刷新令牌重新生成登录令牌, 并记录用户在线状态到数据库实现用户在线状态统计
		userAPI.GET("/keep", middleware.RoleAuth(2), user.KeepOnline)
		// 获取个人资料, 并将可公开的信息存入客户端Cookie用于客户端UI展示
		userAPI.GET("/me", middleware.RoleAuth(2), user.GetProfile)
		// 更新个人资料
		userAPI.PUT("/me", middleware.RoleAuth(2), user.UpdateProfile)
		// 获取全部用户资料
		userAPI.GET("/list", middleware.RoleAuth(3), user.GetAllUser)
		// 删除用户
		userAPI.DELETE("/delete", middleware.RoleAuth(3), user.DeleteUser)
		// 精准搜索用户-直接在URL传用户id
		userAPI.GET("/find/id/:param", middleware.RoleAuth(2), user.GetUserByID)
		// 模糊搜索用户-直接在URL传用户用户名
		userAPI.GET("/find/nickname/:param", user.FinUser)
		// 获取在线用户列表, 仅返回用户id
		userAPI.GET("/online", user.GetOnlineUserList)
		// 获取服务端总用户数和当前在线人数
		userAPI.GET("/online/count", user.GetUserOnlineCount)
	}
	// 缓存清理接口(统一应用权限中间件）
	cacheAPI := r.Group("/cache", middleware.ParamAuth())
	{
		// 传递键名单独或多个删除
		cacheAPI.GET("/clear", monitor.ClearCacheByKey)
		// 全部删除, 这里为了方便直接调用就懒得用DELETE请求方式了
		cacheAPI.GET("/clear/all", monitor.ClearAllCache)
	}
	// 站点访问日志查询和清理接口(统一应用权限中间件）
	logAPI := r.Group("/log", middleware.RoleAuth(3))
	{
		// 清理数据库"接口请求日志"表数据(站点访问日志表)
		logAPI.GET("", monitor.GetRequestLog)
		// 清理数据库"接口请求日志"表数据, 这里图省事直接GET请求传日期和时间即可
		logAPI.DELETE("/clear", monitor.ClearRequestLogBytime)
		// 清理数据库"接口请求日志"表数据全部数据
		logAPI.DELETE("/clear/all", monitor.ClearAllRequestLog)
	}
	return r
}
