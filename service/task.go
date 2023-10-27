package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get query parameter
	kw := ctx.Query("kw")
	is_done := ctx.Query("is_done")

	//set var exist and done_bool like func UpdateTast
	var exist bool
	if is_done == "" {
		exist = false
	} else {
		exist = true
	}
	var done_bool bool
	if exist {
		done_bool, err = strconv.ParseBool(is_done)
		if err != nil {
			Error(http.StatusBadRequest, err.Error())(ctx)
			return
		}
	}

	// Get tasks in DB
	var tasks []database.Task
	switch {
	case kw != "":
		if exist {
			err = db.Select(&tasks, "SELECT * FROM tasks WHERE title LIKE ? AND is_done=?", "%"+kw+"%", done_bool)
		} else {
			err = db.Select(&tasks, "SELECT * FROM tasks WHERE title LIKE ?", "%"+kw+"%")
		}
	default:
		if exist {
			err = db.Select(&tasks, "SELECT * FROM tasks WHERE is_done=?", done_bool)
		} else {
			err = db.Select(&tasks, "SELECT * FROM tasks")
		}
	}

	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// parse ID given as a parameter
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get a task with given ID
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Render task
	//ctx.String(http.StatusOK, task.Title)  // Modify it!!
	ctx.HTML(http.StatusOK, "task.html", task)
}

func NewTaskForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registration"})
}

func RegisterTask(ctx *gin.Context) {
	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Create new data with given title on DB
	result, err := db.Exec("INSERT INTO tasks (title) VALUES (?)", title)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Render status
	path := "/list" // デフォルトではタスク一覧ページへ戻る
	if id, err := result.LastInsertId(); err == nil {
		path = fmt.Sprintf("/task/%d", id) // 正常にIDを取得できた場合は /task/<id> へ戻る
	}
	ctx.Redirect(http.StatusFound, path)
}

func UpdateTask(ctx *gin.Context) {
	// Get task data
	is_done, exist := ctx.GetPostForm("is_done")
	if !exist {
		Error(http.StatusBadRequest, "No situation is given")(ctx)
		return
	}
	done_bool, err := strconv.ParseBool(is_done)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Update data with given title on DB
	if _, err := db.Exec("UPDATE tasks SET is_done=? WHERE id=?", done_bool, id); err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//go back to task page
	path := fmt.Sprintf("/task/%d", id)
	ctx.Redirect(http.StatusFound, path)
}

func EditTaskForm(ctx *gin.Context) {
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Get target task
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Render edit form
	ctx.HTML(http.StatusOK, "form_edit_task.html",
		gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task})
}

func DeleteTask(ctx *gin.Context) {
	// ID の取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks")
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Delete the task from DB
	_, err = db.Exec("DELETE FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Redirect to /list
	ctx.Redirect(http.StatusFound, "/list")
}
