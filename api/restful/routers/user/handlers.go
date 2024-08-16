package user

import (
	. "ManyACG/logger"
	"ManyACG/model"
	"ManyACG/service"
	"errors"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUnauthUser(ctx *gin.Context) {
	objectID, ok := ctx.Get("object_id")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid object id"})
		return
	}
	user, err := service.GetUnauthUserByID(ctx, objectID.(primitive.ObjectID))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get user"})
		return
	}
	ctx.JSON(http.StatusOK, &UnauthUserResponse{
		ID:         user.ID.Hex(),
		Username:   user.Username,
		TelegramID: user.TelegramID,
	})
}

func GetProfile(ctx *gin.Context) {
	claims := jwt.ExtractClaims(ctx)
	username := claims["id"].(string)
	user, err := service.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}
		Logger.Errorf("failed to get user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get user"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "success",
		"data": &UserResponse{
			Username:   user.Username,
			Email:      user.Email,
			TelegramID: user.TelegramID,
			Settings:   user.Settings,
		},
	})

}

func UpdateSettings(ctx *gin.Context) {
	var settings UserSettingsRequest
	if err := ctx.ShouldBind(&settings); err != nil {
		Logger.Errorf("failed to bind json: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}

	claims := jwt.ExtractClaims(ctx)
	username := claims["id"].(string)

	user, err := service.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}
		Logger.Errorf("failed to get user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get user"})
		return
	}
	res, err := service.UpdateUserSettings(ctx, user.ID, (*model.UserSettings)(&settings))
	if err != nil {
		Logger.Errorf("failed to update user settings: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update user settings"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "success",
		"data":    res,
	})
}
