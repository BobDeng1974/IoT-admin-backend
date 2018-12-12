package handler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	myjwt "IoT-admin-backend/middleware"
	"IoT-admin-backend/models"

	jwtgo "github.com/dgrijalva/jwt-go"

	"github.com/gin-gonic/gin"
)

// List all users
func ListUser(c *gin.Context) {
	db := c.MustGet("db").(*mgo.Database)
	var users []models.User
	err := db.C(models.CollectionUser).Find(nil).All(&users)
	if err != nil {
		c.Error(err)
	}

	c.JSON(http.StatusOK, users)
}

// Get a user
func GetUser(c *gin.Context) {
	db := c.MustGet("db").(*mgo.Database)
	var user models.User

	err := db.C(models.CollectionUser).
		FindId(bson.ObjectIdHex(c.Param("_id"))).
		One(&user)
	if err != nil {
		c.Error(err)
	}

	c.JSON(http.StatusOK, user)
}

// Create a user
func CreateUser(c *gin.Context) {
	db := c.MustGet("db").(*mgo.Database)

	var user models.User
	err := c.BindJSON(&user)
	if err != nil {
		c.Error(err)
		return
	}

	err = db.C(models.CollectionUser).Insert(user)
	if err != nil {
		c.Error(err)
	}
	c.JSON(http.StatusOK, user)
}

// Delete user
func DeleteUser(c *gin.Context) {
	db := c.MustGet("db").(*mgo.Database)

	query := bson.M{"_id": bson.ObjectIdHex(c.Param("_id"))}
	err := db.C(models.CollectionUser).Remove(query)
	if err != nil {
		c.Error(err)
	}

	c.JSON(http.StatusOK, nil)
}

// Update user
func UpdateUser(c *gin.Context) {
	db := c.MustGet("db").(*mgo.Database)

	var user models.User
	err := c.BindJSON(&user)
	if err != nil {
		c.Error(err)

		return
	}
	fmt.Print(user)

	// 查找原来的文档
	query := bson.M{
		"_id": bson.ObjectIdHex(c.Param("_id")),
	}

	// 更新
	err = db.C(models.CollectionUser).Update(query, user)
	if err != nil {
		c.Error(err)
	}

	c.JSON(http.StatusOK, user)
}

// List all organization users
func ListOrgUsers(c *gin.Context) {
	db := c.MustGet("db").(*mgo.Database)
	var users []models.User
	query := bson.M{
		"orgnizationId": bson.ObjectIdHex(c.Param("_id")),
		"type":          1,
	}
	err := db.C(models.CollectionUser).Find(query).All(&users)
	if err != nil {
		c.Error(err)
	}

	c.JSON(http.StatusOK, users)
}

// List all customer users
func ListCustomerUsers(c *gin.Context) {
	db := c.MustGet("db").(*mgo.Database)
	var users []models.User
	query := bson.M{
		"customerId": bson.ObjectIdHex(c.Param("_id")),
		"type":       2,
	}
	err := db.C(models.CollectionUser).Find(query).All(&users)
	if err != nil {
		c.Error(err)
	}

	c.JSON(http.StatusOK, users)
}

// LoginResult 登录结果结构
type LoginResult struct {
	Token string `json:"token"`
	models.User
}

// Login 登录
func Login(c *gin.Context) {
	var loginReq models.LoginReq
	if c.BindJSON(&loginReq) == nil {

		db := c.MustGet("db").(*mgo.Database)
		var user models.User
		query := bson.M{
			"email": loginReq.Email,
		}
		err := db.C(models.CollectionUser).
			Find(query).
			One(&user)
		if err != nil {
			c.Error(err)
		}

		if user.Password != loginReq.Password {
			c.JSON(http.StatusOK, gin.H{
				"status": -1,
				"msg":    "Wrong Password!",
			})
		} else {
			generateToken(c, user)
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    "Json parse error!",
		})
	}
}

// 生成令牌
func generateToken(c *gin.Context, user models.User) {
	j := &myjwt.JWT{
		[]byte("FogDong"),
	}
	claims := myjwt.CustomClaims{
		user.Id,
		user.Email,
		jwtgo.StandardClaims{
			NotBefore: int64(time.Now().Unix() - 1000), // 签名生效时间
			ExpiresAt: int64(time.Now().Unix() + 3600), // 过期时间 一小时
			Issuer:    "FogDong",                       //签名的发行者
		},
	}

	token, err := j.CreateToken(claims)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": -1,
			"msg":    err.Error(),
		})
		return
	}

	log.Println(token)

	data := LoginResult{
		User:  user,
		Token: token,
	}
	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"msg":    "登录成功！",
		"data":   data,
	})
	return
}