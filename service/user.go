package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"regexp"
	"strconv"
	"unicode/utf8"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

func NewUserForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user"})
}

func NewPasswordForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)

	var user database.User
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.HTML(http.StatusOK, "new_password_form.html", gin.H{"Title": "Change password", "User": user, "LoggedIn": true})
}

func NewUserNameForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)

	var user database.User
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.HTML(http.StatusOK, "new_username_form.html", gin.H{"Title": "Change Username", "User": user, "LoggedIn": true})
}

func hash(pw string) []byte {
	const salt = "todolist.go#"
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(pw))
	return h.Sum(nil)
}

func passwordFirmChecker(ctx *gin.Context, title string, html string, pwd string) bool {
	if utf8.RuneCountInString(pwd) < 8 {
		ctx.HTML(http.StatusBadRequest, html, gin.H{"Title": title, "Error": "password is too short"})
		return false
	}
	if !regexp.MustCompile(`[0-9]`).Match([]byte(pwd)) {
		ctx.HTML(http.StatusBadRequest, html, gin.H{"Title": title, "Error": "password must include at least one number: [0-9]"})
		return false
	}
	if !regexp.MustCompile(`[a-z]`).Match([]byte(pwd)) {
		ctx.HTML(http.StatusBadRequest, html, gin.H{"Title": title, "Error": "password must include at least one lower-case alphabet: [a-z]"})
		return false
	}
	if !regexp.MustCompile(`[A-Z]`).Match([]byte(pwd)) {
		ctx.HTML(http.StatusBadRequest, html, gin.H{"Title": title, "Error": "password must include at least one upper-case alphabet: [A-Z]"})
		return false
	}
	return true
}

func check_regexp(reg, str string) {
	regexp.MustCompile(reg).Match([]byte(str))
}

func RegisterUser(ctx *gin.Context) {
	// フォームデータの受け取り
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	password_confirm := ctx.PostForm("password_confirm")
	switch {
	case username == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Usernane is not provided", "Username": username})
		return
	case password == "" || password_confirm == "":
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Password": password})
		return
	}
	if password != password_confirm {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password does not match", "Password": password})
		return
	}

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 重複チェック
	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if duplicate > 0 {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password})
		return
	}
	// DB への保存
	result, err := db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 保存状態の確認
	id, _ := result.LastInsertId()
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func RegisterPassword(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get(userkey)

	// フォームデータの受け取り
	password_old := ctx.PostForm("password_old")
	password := ctx.PostForm("password")
	password_confirm := ctx.PostForm("password_confirm")
	switch {
	case password_old == "":
		ctx.HTML(http.StatusBadRequest, "new_password_form.html", gin.H{"Title": "Change password", "Error": "Usernane is not provided", "Password_Old": password_old, "LoggedIn": true})
		return
	case password == "" || password_confirm == "":
		ctx.HTML(http.StatusBadRequest, "new_password_form.html", gin.H{"Title": "Change password", "Error": "Password is not provided", "Password": password, "LoggedIn": true})
		return
	}
	if password != password_confirm {
		ctx.HTML(http.StatusBadRequest, "new_password_form.html", gin.H{"Title": "Change password", "Error": "Password does not match", "Password": password, "LoggedIn": true})
		return
	}

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// old password check
	var userNum int
	err = db.Get(&userNum, "SELECT COUNT(*) FROM users WHERE id=? and password=?", userID, password_old)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if userNum > 0 {
		ctx.HTML(http.StatusBadRequest, "new_password_form.html", gin.H{"Title": " Change password", "Error": "Incorrect Password", "User": user, "LoggedIn": true})
		return
	}
	// DB への保存
	_, err = db.Exec("UPDATE users SET password=? WHERE id=?", hash(password), userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 保存状態の確認
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func RegisterUserName(ctx *gin.Context) {
	// フォームデータの受け取り
	username_new := ctx.PostForm("username_new")
	password := ctx.PostForm("password")
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	userid := sessions.Default(ctx).Get(userkey)

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userid)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// old password check
	var userNum int
	err = db.Get(&userNum, "SELECT COUNT(*) FROM users WHERE id=? and password=?", userid, password)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if userNum > 0 {
		ctx.HTML(http.StatusBadRequest, "new_username_form.html", gin.H{"Title": " Change password", "Error": "Incorrect Password", "User": user, "LoggedIn": true})
		return
	}
	if username_new == "" {
		ctx.HTML(http.StatusBadRequest, "new_username_form.html", gin.H{"Title": "title", "User": user, "Error": "input new username", "LoggedIn": true})
		return
	}

	// 重複チェック
	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username_new) //復活のことも考えて、削除されたユーザーのユーザー名も弾く
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if duplicate > 0 {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "User": user, "LoggedIn": true})
		return
	}
	// DB への保存
	_, err = db.Exec("UPDATE users SET name=? WHERE id=?", username_new, userid)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 保存状態の確認
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.JSON(http.StatusOK, user)
}

const userkey = "user"

func Login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// ユーザの取得
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE name = ? AND deleted=false", username)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
		return
	}

	// パスワードの照合
	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
		return
	}

	// セッションの保存
	session := sessions.Default(ctx)
	session.Set(userkey, user.ID)
	session.Save()

	ctx.Redirect(http.StatusFound, "/list")
}

func LoginForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "Task registration"})
}

func LoginCheck(ctx *gin.Context) {
	if sessions.Default(ctx).Get(userkey) == nil {
		ctx.Redirect(http.StatusFound, "/login")
		ctx.Abort()
	} else {
		ctx.Next()
	}
}

func Logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()
	ctx.Redirect(http.StatusFound, "/")
}

func CorrectUserCheck(ctx *gin.Context) {
	// Deleted task also goes to login menu
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	userid := sessions.Default(ctx).Get(userkey)

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get tasks in DB
	var owner []database.Owner
	err = db.Select(&owner, "SELECT user_id, task_id FROM ownership WHERE user_id = ? AND task_id = ? AND deleted=false", userid, id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	if owner == nil {
		ctx.Redirect(http.StatusFound, "/login")
		ctx.Abort()
	} else {
		ctx.Next()
	}
}
