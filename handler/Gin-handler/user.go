package GinHandler

import (
	"FileCloud/common"
	"FileCloud/config"
	dblayer "FileCloud/db"
	nativeHandler "FileCloud/handler"
	"FileCloud/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

//SignupHandler: 返回注册页面
func SignupHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

//DoSignupHandler: 处理用户注册
func DoSignupHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	passwd := c.Request.FormValue("password")

	if len(username) < 3 || len(passwd) < 5 {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Invalid Parameter",
			"code": common.StatusParamInvalid,
		})
		return
	}

	//对密码进行加盐及取Sha1值加密
	encPasswd := util.Sha1([]byte(passwd + config.PwdSalt))
	ok := dblayer.UserSignup(username, encPasswd)
	if ok {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Signup succeeded",
			"code": common.StatusOK,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Signup failed",
			"code": common.StatusRegisterFailed,
		})
	}
}

//SignInHandler: 登录接口
func SignInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signin.html")
}

//DoSignInHandler: 处理登录请求
func DoSignInHandler(c *gin.Context) {
	var location string
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")

	//1.校验用户名与密码
	encPasswd := util.Sha1([]byte(password + config.PwdSalt))
	pwdChecked := dblayer.UserSignin(username, encPasswd)
	if !pwdChecked {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Name or password wrong",
			"code": common.StatusLoginFailed,
		})
		return
	}

	//2.生成用户访问凭证Token
	token := nativeHandler.GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "update token failed",
			"code": common.StatusLoginFailed,
		})
		return
	}
	//查询用户状态值
	user, err := dblayer.GetUserStatus(username)
	if err != nil {
		fmt.Println(err.Error())
	}
	//普通用户权限登录跳转
	if user.Status == 0 {
		location = "http://localhost:8080/static/view/home.html"
		//管理员权限登录跳转
	} else if user.Status == 7 {
		location = "http://localhost:8080/static/view/admin.html"
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
	c.Data(http.StatusOK, "text/plain", resp.JSONBytes())
}

//UserInfoHandler: 查询用户信息
func UserInfoHandler(c *gin.Context) {
	//1.解析请求参数
	username := c.Request.FormValue("username")

	//3.查询用户信息
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusOK, gin.H{
			"msg":  "查询用户信息失败",
			"code": common.StatusUserNotExists,
		})
		return
	}

	//4.组装并响应用户数据
	resp := util.RespMsg{
		Code: int(common.StatusOK),
		Msg:  "查询用户信息成功",
		Data: user,
	}
	c.Data(http.StatusOK, "text/plain", resp.JSONBytes())
}

//更新用户信息
func UpdateUserInfo(c *gin.Context) {
	var realpassword string
	//接收参数
	email := c.Request.FormValue("email")
	phone := c.Request.FormValue("phone")
	password := c.Request.FormValue("password")
	username := c.Request.FormValue("username")

	//如果密码有改动则调用更新密码的db方法
	//如果密码无改动则调用不更新密码的db方法
	if password != "" {
		realpassword = password
		enc_passwd := util.Sha1([]byte(realpassword + config.PwdSalt))
		//调用db模块
		res := dblayer.UpdateUserInfoIncludePWD(username, enc_passwd, phone, email)
		if res {
			fmt.Println("更新成功！")
			c.Redirect(http.StatusFound, "http://localhost:8080/static/view/home.html")
		}
	} else {
		res := dblayer.UpdateUserExceptPWD(username, phone, email)
		if res {
			fmt.Println("更新成功！")
			c.Redirect(http.StatusFound, "http://localhost:8080/static/view/home.html")
		}
	}
}

//查询所有注册用户
func UserQueryHandler(c *gin.Context) {
	if c.Request.Method == "GET" {
		//返回上传html页面
		c.Redirect(http.StatusFound, "http://localhost:8080/static/view/admin.html")
		//data, err := ioutil.ReadFile("src/FileCloud/static/view/admin.html")
	} else {
		users, err := dblayer.GetAllUser()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		resp := util.RespMsg{
			Data: users,
		}
		c.Data(http.StatusOK, "text/plain", resp.JSONBytes())
	}
}

//改变用户账户状态
func UpdateUserStatus(c *gin.Context) {
	username := c.Request.Form.Get("CurlUsername")
	status := c.Request.Form.Get("status")
	realStatus, _ := strconv.Atoi(status)
	dblayer.UpdateUserStatus(username, realStatus)
}

//新增管理员页面接口
func AddHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/AddAdmin.html")
}

//新增管理员
func AddAdmin(c *gin.Context) {

	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	status := c.Request.FormValue("status")
	fmt.Println(username + password + status)
	realStatus, _ := strconv.Atoi(status)
	if len(username) < 3 || len(password) < 5 {
		fmt.Println("设置错误")
		return
	}

	//对用户密码进行哈希的加密处理
	enc_passwd := util.Sha1([]byte(password + config.PwdSalt))
	suc := dblayer.AddAdmin(username, enc_passwd, realStatus)
	if suc {
		c.Writer.Write([]byte("SUCCESS"))
	} else {
		c.Writer.Write([]byte("FAILED"))
	}

}

//删除系统用户
func DeleteUserHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	if dblayer.DeleteUser(username) {
		c.Writer.WriteHeader(http.StatusOK)
	} else {
		c.Writer.Write([]byte("文件删除失败！"))
	}
}