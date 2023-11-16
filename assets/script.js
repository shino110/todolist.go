/* placeholder file for JavaScript */
const confirm_delete = (id) => {
    if(window.confirm(`Task ${id} を削除します．よろしいですか？`)) {
        location.href = `/task/delete/${id}`;
    }
}
 
const confirm_deleteAll = (id) => {
    if(window.confirm(`全てのタスクを削除します．よろしいですか？`)) {
        location.href = `/user/delete_task/${id}`;
    }
}

const confirm_deleteUser = (id) => {
    if(window.confirm(`アカウントを削除します．よろしいですか？`)) {
        location.href = `/user/delete/${id}`;
    }
}

const confirm_update = (id) => {
    // 練習問題 7-2
    if(window.confirm(`Task ${id} を更新します．よろしいですか？`)) {
        location.href = `/task/${id}`;
    }
}