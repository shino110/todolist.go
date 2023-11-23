package service

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

//all input datetime-local is initiallized with either of these values
//const inputTimeMin := time.Date(1000, time.Month(time.January), 1, 0, 0, 0, 0, time.UTC)
//const inputTimeMax := time.Date(9999, time.Month(time.December), 31, 23, 59, 0, 0, time.UTC)

// 12000-12-12 00:00:00 <- input_min.Format("1000-01-01 00:00:00")
// Formatが動いて欲しいように動かない
func PutTimeinSQLdatetime(tm time.Time) string {
	year, month, day := tm.Date()
	return fmt.Sprintf("%04d", year) + "-" + fmt.Sprintf("%02d", int(month)) + "-" + fmt.Sprintf("%02d", day) + " " + fmt.Sprintf("%02d", tm.Hour()) + ":" + fmt.Sprintf("%02d", tm.Minute()) + ":" + fmt.Sprintf("%02d", tm.Second())
}

// input: YYYY-mm-ddTHH%3AMM   %3A is ":"
func DateTimeInput2Time(ctx *gin.Context, str string) time.Time {
	arr1 := strings.Split(str, "-")
	year, err := strconv.Atoi(arr1[0])
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
	}
	mon, err := strconv.Atoi(arr1[1])
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
	}
	arr2 := strings.Split(arr1[2], "T")
	day, err := strconv.Atoi(arr2[0])
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
	}
	arr3 := strings.Split(arr2[1], ":")
	hour, err := strconv.Atoi(arr3[0])
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
	}
	min, err := strconv.Atoi(arr3[1])
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
	}
	return time.Date(year, time.Month(mon), day, hour, min, 0, 0, time.UTC)
}

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	inputTimeMin := time.Date(1000, time.Month(time.January), 1, 0, 0, 0, 0, time.UTC)
	inputTimeMax := time.Date(9999, time.Month(time.December), 31, 23, 59, 0, 0, time.UTC)

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// Get query parameter
	kw := ctx.Query("kw")
	is_done := ctx.Query("is_done")
	input_start := ctx.Query("deadline-start")
	input_end := ctx.Query("deadline-end")
	var deadline_start, deadline_end time.Time
	datetime_default := false //check if two datetime input is default

	// putting input string into SQL results in SQL error
	// below parse intput string to time.Time
	// datetime in MySQL does nothing with timezone but deal LITERALLY so using UTC

	//convert input dates string to time.Time
	if input_start != "" {
		deadline_start = DateTimeInput2Time(ctx, input_start)
	}
	if input_end == "" {
		if input_start == "" {
			datetime_default = true
		}
	} else {
		deadline_end = DateTimeInput2Time(ctx, input_end)
	}
	if deadline_start.Before(deadline_end) {
		Error(http.StatusBadRequest, "Put appropriate datetime")
	}
	if deadline_start == inputTimeMin && deadline_end == inputTimeMax {
		datetime_default = true
	}

	//set var done_selected and done_bool like func UpdateTask
	done_selected := true
	if is_done == "" {
		done_selected = false
	}
	var done_bool bool
	if done_selected {
		done_bool, err = strconv.ParseBool(is_done)
		if err != nil {
			Error(http.StatusBadRequest, err.Error())(ctx)
			return
		}
	}

	// //make arrays of conditions
	var conditions []string
	if done_selected {
		conditions = append(conditions, "is_done="+strconv.FormatBool(done_bool))
	}
	if !datetime_default {
		conditions = append(conditions,
			"deadline BETWEEN '"+PutTimeinSQLdatetime(deadline_start)+"' AND '"+PutTimeinSQLdatetime(deadline_end)+"'")
	}

	// var query string
	var query string
	if len(conditions) > 0 {
		for i := 0; i < len(conditions)-1; i++ {
			query = query + conditions[i] + " AND "
		}
		query = query + conditions[len(conditions)-1]
	}

	// Get tasks in DB
	var tasks []database.Task
	base_query := "SELECT id, title, deadline, created_at, is_done, memo FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ? AND deleted=false"
	if len(conditions) > 0 && kw != "" {
		err = db.Select(&tasks, base_query+" AND "+query+" AND title LIKE ? ", userID, "%"+kw+"%")
	} else if kw != "" {
		err = db.Select(&tasks, base_query+" AND title LIKE ? ", userID, "%"+kw+"%")
	} else if len(conditions) > 0 {
		err = db.Select(&tasks, base_query+" AND "+query, userID)
	} else {
		err = db.Select(&tasks, base_query, userID)
	}

	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Render tasks
	ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "UserID": userID, "LoggedIn": true, "Tasks": tasks})
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	//will not access to deleted task because of CorrectUserCheck() on user.go
	userID := sessions.Default(ctx).Get("user") //will be used to access mypage

	// parse ID given as a parameter
	task_id, err := strconv.Atoi(ctx.Param("id"))
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

	// Get a task with given ID
	var task database.Task
	err = db.Get(&task, "SELECT id, title, deadline, created_at, memo, is_done FROM tasks WHERE id=? ", task_id) // Use DB#Get for one entry
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Render task
	ctx.HTML(http.StatusOK, "task.html", gin.H{"UserID": userID, "LoggedIn": true, "Task": task})
}

func NewTaskForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"UserID": userID, "LoggedIn": true, "Title": "Task registration"})
}

func RegisterTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")

	// Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist {
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}
	memo, exist := ctx.GetPostForm("memo")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Create new data with given title on DB
	tx := db.MustBegin()
	var result sql.Result
	if exist {
		result, err = tx.Exec("INSERT INTO tasks (title, memo) VALUES (?, ?)", title, memo)
	} else {
		result, err = tx.Exec("INSERT INTO tasks (title) VALUES (?)", title)
	}
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Render status
	taskID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	_, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userID, taskID)
	if err != nil {
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx.Commit()
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
}

func UpdateTask(ctx *gin.Context) {
	// Get task data
	is_done, done_selected := ctx.GetPostForm("is_done")
	if !done_selected {
		Error(http.StatusBadRequest, "No situation is given")(ctx)
		return
	}
	done_bool, err := strconv.ParseBool(is_done)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	deadline, deadline_selected := ctx.GetPostForm("deadline")
	if !deadline_selected {
		Error(http.StatusBadRequest, "Deadline not selected")(ctx) //should not happen
		return
	}
	deadlineTime := DateTimeInput2Time(ctx, deadline)
	inputTimeMin := time.Date(1000, time.Month(time.January), 1, 0, 0, 0, 0, time.UTC)

	memo, memo_written := ctx.GetPostForm("memo")
	if !memo_written {
		Error(http.StatusBadRequest, "Deadline not selected")(ctx) //should not happen
		return
	}

	updateValues := "is_done=" + strconv.FormatBool(done_bool)
	if deadlineTime.After(inputTimeMin) { //initial value changed 1000-01-01T00:00
		updateValues = updateValues + ", deadline='" + PutTimeinSQLdatetime(deadlineTime) + "'"
	}
	if memo != "" {
		updateValues = updateValues + ", memo='" + memo + "'"
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
	if _, err := db.Exec("UPDATE tasks SET "+updateValues+" WHERE id=?", id); err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//go back to task page
	path := fmt.Sprintf("/task/%d", id)
	ctx.Redirect(http.StatusFound, path)
}

func EditTaskForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
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
	err = db.Get(&task, "SELECT id, title, deadline, created_at, memo, is_done FROM tasks WHERE id=?", id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	// Render edit form
	ctx.HTML(http.StatusOK, "form_edit_task.html",
		gin.H{"UserID": userID, "LoggedIn": true, "Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task})
}

func DeleteTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
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
	// change Delete flag
	_, err = db.Exec("UPDATE ownership SET deleted=true WHERE task_id=? AND user_id=?", id, userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// Redirect to /list
	ctx.Redirect(http.StatusFound, "/list")
}

func DeleteTaskAll(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// change Delete flag
	_, err = db.Exec("UPDATE ownership SET deleted=true WHERE user_id=?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	path := fmt.Sprintf("/user/%d", userID)
	ctx.Redirect(http.StatusFound, path)
}

func DeleteUser(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	// change Delete flag
	_, err = db.Exec("UPDATE ownership SET deleted=true WHERE user_id=?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	_, err = db.Exec("UPDATE users SET deleted=true WHERE id=?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.Redirect(http.StatusFound, "/logout")
}

func DashboardForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	var user database.User

	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	err = db.Get(&user, "SELECT id, name FROM users WHERE id=?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.HTML(http.StatusFound, "dashboard.html", gin.H{"UserID": userID, "User": user, "LoggedIn": true})
	// path := fmt.Sprintf("/user/%d", userID)
	// ctx.Redirect(http.StatusFound, path)
}
