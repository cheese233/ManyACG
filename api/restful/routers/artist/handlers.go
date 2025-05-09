package artist

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/krau/ManyACG/common"

	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetArtist(ctx *gin.Context) {
	artistID := ctx.MustGet("object_id").(primitive.ObjectID)
	artist, err := service.GetArtistByID(ctx, artistID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artist not found")
			return
		}
		common.Logger.Errorf("Failed to get artist: %v", err)
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artist")
		return
	}
	ctx.JSON(http.StatusOK, common.RestfulCommonResponse[*types.Artist]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    artist,
	})
}
