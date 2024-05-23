package controller

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit123sinha456/digitalSignage/dbmaster"
	DataModel "github.com/rohit123sinha456/digitalSignage/model"
	"go.mongodb.org/mongo-driver/bson"
)

// type EventStreamRequest struct {
// 	Message string `form:"message" json:"message" binding:"required,max=100"`
// }

func CreateScreenController(c *gin.Context) {
	var requestjsonvar DataModel.Screen
	userid := c.GetHeader("userid")
	reqerr := c.Bind(&requestjsonvar)
	log.Printf("%+v", requestjsonvar)
	if reqerr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": reqerr.Error()})
	}
	screenid, err := dbmaster.CreateScreen(c, Client, userid, requestjsonvar)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"screenid": screenid})
	}
}

func ReadScreenController(c *gin.Context) {
	var contentarray []DataModel.Screen
	userid := c.GetHeader("userid")
	contentarray, err := dbmaster.ReadScreen(c, Client, userid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"contents": contentarray})
}

func GetScreenbyIDController(c *gin.Context) {
	userid := c.GetHeader("userid")
	contentID := c.Params.ByName("id")
	user, err := dbmaster.ReadOneScreen(c, Client, userid, contentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"content": user})
	}
}

func UpdateScreenbyIDController(c *gin.Context) {
	var requestjsonvar []DataModel.ScreenBlock
	userid := c.GetHeader("userid")
	screenID := c.Params.ByName("id")
	reqerr := c.Bind(&requestjsonvar)
	log.Printf("%+v", requestjsonvar)
	if reqerr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": reqerr.Error()})
	}
	err := dbmaster.UpdateScreen(c, Client, userid, screenID, requestjsonvar)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "Updated"})
	}
}

func HandleEventStreamPost(c *gin.Context, ch chan DataModel.EventStreamRequest, screencode string) {
	var result DataModel.EventStreamRequest
	var userinfo DataModel.UserSystemIdentifeir
	coll := Client.Database("user").Collection("userSystemInfo")
	userid := c.GetHeader("userid")
	filter := bson.D{{"userid", userid}}
	err := coll.FindOne(c, filter).Decode(&userinfo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}
	result.Screencode = screencode
	result.Userinfo = userinfo
	if err := c.ShouldBind(&result); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}
	ch <- result
	c.JSON(http.StatusOK, gin.H{"status": "Updated"})
	return
}

func HandleEventStreamGet(c *gin.Context, ch chan DataModel.EventStreamRequest) {
	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-ch; ok {
			c.SSEvent("message", msg)
			return true
		}
		return false
	})

	return
}
