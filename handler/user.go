package handler

import (
	dblayer "FileCloud/db"
	"FileCloud/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	pwd_salt = "*#890"
)

//处理用户注册请求
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("src/FileCloud/static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(data)
		return
	}

	//表单参数解析
	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")

	if len(username) < 3 || len(passwd) < 5 {
		w.Write([]byte("invalid parameter"))
		return
	}

	//对用户密码进行哈希的加密处理
	enc_passwd := util.Sha1([]byte(passwd + pwd_salt))
	suc := dblayer.UserSignup(username, enc_passwd)
	if suc {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}
}

//用户登录接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {

	var location string

	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	encpasswd := util.Sha1([]byte(password + pwd_salt))

	pwdChecked := dblayer.UserSignin(username, encpasswd)
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}

	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("FAILED"))
		return
	}
	//查询用户状态值
	user, err := dblayer.GetUserStatus(username)
	if err != nil {
		fmt.Println(err.Error())
	}
	//普通用户权限登录跳转
	if user.Status == 0 {
		location = "http://" + r.Host + "/static/view/home.html"
		//管理员权限登录跳转
	} else if user.Status == 7 {
		location = "http://" + r.Host + "/static/view/admin.html"
	}

	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Status   int
			Token    string
		}{
			Location: location,
			Username: username,
			Status:   user.Status,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())

}

// UserInfoHandler ： 查询用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	//	token := r.Form.Get("token")

	//以下逻辑已在拦截器中重写
	// // 2. 验证token是否有效
	// isValidToken := IsTokenValid(token)
	// if !isValidToken {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	return
	// }

	// 3. 查询用户信息
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 4. 组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

//更新用户信息
func UpdateUserInfo(w http.ResponseWriter, r *http.Request) {
	var realpassword string
	//接收参数
	r.ParseForm()
	email := r.Form.Get("email")
	phone := r.Form.Get("phone")
	password := r.Form.Get("password")
	username := r.Form.Get("username")
	//如果密码有改动则调用更新密码的db方法
	//如果密码无改动则调用不更新密码的db方法
	if password != "" {
		realpassword = password
		enc_passwd := util.Sha1([]byte(realpassword + pwd_salt))
		//调用db模块
		res := dblayer.UpdateUserInfoIncludePWD(username, enc_passwd, phone, email)
		if res {
			fmt.Println("更新成功！")
			http.Redirect(w, r, "http://"+r.Host+"/static/view/home.html", http.StatusFound)
		}
	} else {
		res := dblayer.UpdateUserExceptPWD(username, phone, email)
		if res {
			fmt.Println("更新成功！")
			http.Redirect(w, r, "http://"+r.Host+"/static/view/home.html", http.StatusFound)
		}
	}

}

// GenToken : 生成token
func GenToken(username string) string {
	// 40位字符:md5(username+timestamp+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

// IsTokenValid : token是否有效
func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断token的时效性，是否过期
	// TODO: 从数据库表tbl_user_token查询username对应的token信息
	// TODO: 对比两个token是否一致
	return true
}

//查询所有注册用户
func UserQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//返回上传html页面
		data, err := ioutil.ReadFile("src/FileCloud/static/view/admin.html")
		if err != nil {
			io.WriteString(w, "internal ser ver error")
			return
		}
		io.WriteString(w, string(data))
	} else {
		users, err := dblayer.GetAllUser()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp := util.RespMsg{
			Code: 0,
			Msg:  "OK",
			Data: users,
		}
		w.Write(resp.JSONBytes())
	}
}

//改变用户账户状态
func UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("CurlUsername")
	status := r.Form.Get("status")
	realStatus, _ := strconv.Atoi(status)
	dblayer.UpdateUserStatus(username, realStatus)
}

//新增管理员
func AddAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("src/FileCloud/static/view/AddAdmin.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(data)
		return
	}
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	status := r.Form.Get("status")
	fmt.Println(username + password + status)
	realStatus, _ := strconv.Atoi(status)
	if len(username) < 3 || len(password) < 5 {
		w.Write([]byte("invalid parameter"))
		return
	}

	//对用户密码进行哈希的加密处理
	enc_passwd := util.Sha1([]byte(password + pwd_salt))
	suc := dblayer.AddAdmin(username, enc_passwd, realStatus)
	if suc {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}

}

//删除系统用户
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	if dblayer.DeleteUser(username) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.Write([]byte("用户删除失败！"))
	}
}
