package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"todolist.go/db"
	"todolist.go/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

const port = 8000

func main() {
	// initialize DB connection
	dsn := db.DefaultDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err := db.Connect(dsn); err != nil {
		log.Fatal(err)
	}

	// initialize Gin engine
	engine := gin.Default()
	engine.LoadHTMLGlob("views/*.html")

	// prepare session
	store := cookie.NewStore([]byte("my-secret"))
	engine.Use(sessions.Sessions("user-session", store))

	// routing
	engine.Static("/assets", "./assets")
	engine.GET("/", service.Home)
	engine.GET("/list", service.LoginCheck, service.TaskList)

	taskGroup := engine.Group("/task")
	taskGroup.Use(service.LoginCheck)
	{
		taskGroup.GET("/:id", service.CorrectUserCheck, service.ShowTask) // ":id" is a parameter
		// タスクの新規登録
		taskGroup.GET("/new", service.NewTaskForm)
		taskGroup.POST("/new", service.RegisterTask)
		// 既存タスクの編集
		taskGroup.GET("/edit/:id", service.CorrectUserCheck, service.EditTaskForm)
		taskGroup.POST("/edit/:id", service.CorrectUserCheck, service.NotImplemented)
		// 既存タスクの削除
		taskGroup.GET("/delete/:id", service.CorrectUserCheck, service.DeleteTask)
	}

	// ユーザ登録
	engine.GET("/user/new", service.NewUserForm)
	engine.POST("/user/new", service.RegisterUser)

	engine.GET("/login", service.LoginForm)
	engine.POST("/login", service.Login)

	engine.GET("/logout", service.LogoutForm)
	engine.POST("/logout", service.Logout)

	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}
