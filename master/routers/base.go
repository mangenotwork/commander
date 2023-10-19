package routers

import (
	"gitee.com/mangenotework/commander/common/logger"
	"gitee.com/mangenotework/commander/common/utils"
	"gitee.com/mangenotework/commander/master/dao"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var Router *gin.Engine

func init() {
	Router = gin.Default()
}

func Routers() *gin.Engine {
	Router.StaticFS("/static", http.Dir("./static"))

	//模板
	// 自定义模板方法
	//Router.SetFuncMap(template.FuncMap{
	//	//"formatAsDate": FormatAsDate,
	//})

	Router.Delims("{[", "]}")

	//Router.LoadHTMLGlob("views/*")
	Router.LoadHTMLGlob("views/**/*")

	V1()

	return Router
}

// HttpMiddleware http中间件
func HttpMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		go func() {
			// 获取客户端ip
			ip := GetIP(ctx.Request)
			logger.Info("[HTTP]", ip, " | ", ctx.Request.URL.Path)
			_ = new(dao.DaoOperate).Set(ip, ctx.Request.URL.Path)
		}()
		ctx.Next()
	}
}

// GetIP 获取ip
// - X-Real-IP：只包含客户端机器的一个IP，如果为空，某些代理服务器（如Nginx）会填充此header。
// - X-Forwarded-For：一系列的IP地址列表，以,分隔，每个经过的代理服务器都会添加一个IP。
// - RemoteAddr：包含客户端的真实IP地址。 这是Web服务器从其接收连接并将响应发送到的实际物理IP地址。 但是，如果客户端通过代理连接，它将提供代理的IP地址。
//
// RemoteAddr是最可靠的，但是如果客户端位于代理之后或使用负载平衡器或反向代理服务器时，它将永远不会提供正确的IP地址，因此顺序是先是X-REAL-IP，
// 然后是X-FORWARDED-FOR，然后是 RemoteAddr。 请注意，恶意用户可以创建伪造的X-REAL-IP和X-FORWARDED-FOR标头。
func GetIP(r *http.Request) (ip string) {
	for _, ip := range strings.Split(r.Header.Get("X-Forward-For"), ",") {
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	if ip = r.Header.Get("X-Real-IP"); net.ParseIP(ip) != nil {
		return ip
	}
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	return "0.0.0.0"
}

// CorsHandler 跨域中间件
func CorsHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "*")
		ctx.Header("Access-Control-Allow-Headers", "*")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, "+
			"Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,"+
			"Pragma,FooBar") // 跨域关键设置 让浏览器可以解析
		ctx.Header("Access-Control-Max-Age", "172800")          // 缓存请求信息 单位为秒
		ctx.Header("Access-Control-Allow-Credentials", "false") //  跨域请求是否需要带cookie信息 默认设置为true
		ctx.Header("content-type", "application/json")          // 设置返回格式是json
		//Release all option pre-requests
		if ctx.Request.Method == http.MethodOptions {
			ctx.JSON(http.StatusOK, "Options Request!")
		}
		ctx.Next()
	}
}

// AuthPG 认证  权限验证中间件
func AuthPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie("authenticate")
		if token != "" {
			j := utils.NewJWT()
			err := j.ParseToken(token)
			if err == nil {
				account := j.GetString("account")
				password := j.GetString("password")
				user, uErr := new(dao.DaoUser).Get()
				if uErr == nil && account == user.Account && password == user.Password {
					c.Next()
					return
				}
			}
		}
		c.Redirect(http.StatusFound, "/")
		return
	}
}

// TODO 请求记录
