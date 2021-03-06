package db

import (
	mydb "FileCloud/db/mysql"
	"fmt"
)

// User : 用户表model
type User struct {
	Username     string
	Userpwd      string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

// UserSignup : 通过用户名及密码完成user表的注册操作
func UserSignup(username string, passwd string, email string, phone string, EmailValidate int, PhoneValidate int) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user (`user_name`,`user_pwd`,`email`,`phone`,`email_validated`,`phone_validated`) values (?,?,?,?,?,?)")
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, passwd, email, phone, EmailValidate, PhoneValidate)
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}
	if rowsAffected, err := ret.RowsAffected(); nil == err && rowsAffected > 0 {
		return true
	}
	return false
}

// UserSignin : 判断密码是否一致
func UserSignin(username string, encpwd string) bool {
	stmt, err := mydb.DBConn().Prepare("select * from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Println("username not found: " + username)
		return false
	}

	pRows := mydb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encpwd {
		return true
	}
	return false
}

//查询用户状态值
func GetUserStatus(username string) (User, error) {
	user := User{}

	stmt, err := mydb.DBConn().Prepare(
		"select status from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	// 执行查询的操作
	err = stmt.QueryRow(username).Scan(&user.Status)
	if err != nil {
		return user, err
	}
	return user, nil
}

// UpdateToken : 刷新用户登录的token
func UpdateToken(username string, token string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"replace into tbl_user_token (`user_name`,`user_token`) values (?,?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// 查询用户Token
func GetUserToken(username string) string {
	stmt, err := mydb.DBConn().Prepare(
		"select user_token from tbl_user_token where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return err.Error()
	}
	var token string
	err = stmt.QueryRow(username).Scan(&token)
	if err != nil {
		fmt.Println(err.Error())
		return err.Error()
	}
	fmt.Println("Token查询无差错，已正确返回...")
	return token
}

// GetUserInfo : 查询用户信息
func GetUserInfo(username string) (User, error) {
	user := User{}

	stmt, err := mydb.DBConn().Prepare(
		"select user_name,user_pwd,signup_at,email,phone,status from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	// 执行查询的操作
	err = stmt.QueryRow(username).Scan(&user.Username, &user.Userpwd, &user.SignupAt, &user.Email, &user.Phone, &user.Status)
	if err != nil {
		return user, err
	}
	return user, nil

}

// 获取用户属下文件总数
func GetFileNumByUserName(username string) int64 {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT COUNT(*) FROM tbl_user_file WHERE user_name = ?;")
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	defer stmt.Close()
	var count int64
	err = stmt.QueryRow(username).Scan(&count)
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	return count
}

// 获取系统所有用户数
func GetUserNum() int64 {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT COUNT(*) FROM tbl_user;")
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	defer stmt.Close()
	var count int64
	err = stmt.QueryRow().Scan(&count)
	if err != nil {
		fmt.Println(err.Error())
		return 0
	}
	return count
}

//更新用户信息（包含密码）
func UpdateUserInfoIncludePWD(username, userpwd, phone, email string) bool {
	stmt, err := mydb.DBConn().Prepare("update tbl_user set user_pwd=?,phone= ?,email=? where user_name= ? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	_, err = stmt.Exec(userpwd, phone, email, username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

//更新用户信息（不包含密码）
func UpdateUserExceptPWD(username, phone, email string) bool {
	stmt, err := mydb.DBConn().Prepare("update tbl_user set phone= ?,email=? where user_name= ? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	_, err = stmt.Exec(phone, email, username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

//查询所有用户，用于用户管理
//改造加入分页
func GetAllUser(pageIndex int, pageSize int) ([]User, error) {
	stmt, err := mydb.DBConn().Prepare("select user_name,user_pwd,email,phone,signup_at,last_active,status from tbl_user limit ?,?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query((pageIndex-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}

	var users []User
	for rows.Next() {
		user := User{}
		err = rows.Scan(&user.Username, &user.Phone, &user.Email, &user.Phone, &user.SignupAt, &user.LastActiveAt, &user.Status)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		users = append(users, user)
	}
	return users, nil
}

//更新用户账户状态
func UpdateUserStatus(username string, status int) bool {
	stmt, err := mydb.DBConn().Prepare("update tbl_user set status= ? where user_name= ? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	_, err = stmt.Exec(status, username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

//新增管理员
func AddAdmin(username string, passwd string, status int) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user (`user_name`,`user_pwd`,`status`) values (?,?,?)")
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, passwd, status)
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}
	if rowsAffected, err := ret.RowsAffected(); nil == err && rowsAffected > 0 {
		return true
	}
	return false
}

//删除系统用户
func DeleteUser(username string) bool {
	stmt, err := mydb.DBConn().Prepare(`DELETE FROM tbl_user WHERE user_name =?`)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	_, err = stmt.Exec(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
