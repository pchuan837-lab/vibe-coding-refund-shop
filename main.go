// refund-shop 服务入口。
//
// 本地启动：go run .   （goproxy 已默认国内镜像）
// 自动用 go:embed 打包 public/ 三页静态前端，与后端共用 3000 端口，无需单独起前端。
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"refund-shop/internal/db"
	"refund-shop/internal/routes"
)

//go:embed all:public
var publicFS embed.FS

func main() {
	database, err := db.NewDB("data.db")
	if err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer database.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// API 分组
	api := r.Group("/api")
	routes.RegisterOrders(api, database)
	routes.RegisterRefunds(api, database)

	// 静态前端（go:embed all:public，Sub 掉前缀目录）
	// 说明：不能用 r.StaticFS("/", ...)：其 `/*filepath` 通配与 /api 前缀路由冲突。
	// 改用 NoRoute + http.FileServer 兜底：已注册的 /api 路由优先命中，其余交给静态文件服务。
	staticFS, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatalf("embed public: %v", err)
	}
	fileServer := http.FileServer(http.FS(staticFS))
	r.NoRoute(func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	log.Println("✅ refund-shop 启动成功 → http://localhost:3000")
	if err := r.Run(":3000"); err != nil {
		log.Fatalf("server run: %v", err)
	}
}