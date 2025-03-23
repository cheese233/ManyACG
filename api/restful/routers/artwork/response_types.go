package artwork

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/types"
)

type ArtworkResponseData struct {
	ID          string             `json:"id"`
	CreatedAt   string             `json:"created_at"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	SourceURL   string             `json:"source_url"`
	R18         bool               `json:"r18"`
	LikeCount   uint               `json:"like_count"`
	Tags        []string           `json:"tags"`
	Artist      *types.Artist      `json:"artist"`
	SourceType  types.SourceType   `json:"source_type"`
	Pictures    []*PictureResponse `json:"pictures"`
}

type PictureResponse struct {
	ID        string `json:"id"`
	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	Index     uint   `json:"index"`
	Hash      string `json:"hash"`
	FileName  string `json:"file_name"`
	MessageID int    `json:"message_id"`
	Thumbnail string `json:"thumbnail"`
	Regular   string `json:"regular"`
	Original  string `json:"original"`
}

func ResponseFromArtwork(ctx *gin.Context, artwork *types.Artwork, isAuthorized bool) *common.RestfulCommonResponse[any] {
	// if isAuthorized {
	// 	return &common.RestfulCommonResponse[any]{
	// 		Status:  http.StatusOK,
	// 		Message: "Success",
	// 		Data:    artwork,
	// 	}
	// } // why?
	return &common.RestfulCommonResponse[any]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    ResponseDataFromArtwork(ctx, artwork),
	}
}

func getPictureUrl(ctx *gin.Context, picture *types.Picture, quality string) string {
	var storageInfo *types.StorageDetail
	var host string
	switch quality {
	case "thumbnail":
		storageInfo = picture.StorageInfo.Thumb
	case "regular":
		storageInfo = picture.StorageInfo.Regular
	case "original":
		storageInfo = picture.StorageInfo.Original
	}

	if picture.StorageInfo == nil || storageInfo == nil {
		return picture.Thumbnail
	}

	if storageInfo.Type == types.StorageTypeAlist {
		return common.ApplyApiPathRule(storageInfo.Path)
	}

	if config.Cfg.API.Host != "" {
		host = config.Cfg.API.Host
	} else {
		host = "//" + ctx.Request.Host
	}
	return host + "/api/v1/picture/file/" + picture.ID + "?quality=" + quality
}

func ResponseDataFromArtwork(ctx *gin.Context, artwork *types.Artwork) *ArtworkResponseData {
	pictures := make([]*PictureResponse, len(artwork.Pictures))
	for i, picture := range artwork.Pictures {
		pictures[i] = &PictureResponse{
			ID:        picture.ID,
			Width:     picture.Width,
			Height:    picture.Height,
			Index:     picture.Index,
			Hash:      picture.Hash,
			FileName:  picture.GetFileName(),
			MessageID: picture.TelegramInfo.MessageID,
			Thumbnail: getPictureUrl(ctx, picture, "thumbnail"),
			Regular:   getPictureUrl(ctx, picture, "regular"),
			Original:  getPictureUrl(ctx, picture, "original"),
		}
	}
	return &ArtworkResponseData{
		ID:          artwork.ID,
		CreatedAt:   artwork.CreatedAt.Format("2006-01-02 15:04:05"),
		Title:       artwork.Title,
		Description: artwork.Description,
		SourceURL:   artwork.SourceURL,
		R18:         artwork.R18,
		LikeCount:   artwork.LikeCount,
		Tags:        artwork.Tags,
		Artist:      artwork.Artist,
		SourceType:  artwork.SourceType,
		Pictures:    pictures,
	}
}

func ResponseFromArtworks(ctx *gin.Context, artworks []*types.Artwork, isAuthorized bool) *common.RestfulCommonResponse[any] {
	// if isAuthorized {
	// 	return &common.RestfulCommonResponse[any]{
	// 		Status:  http.StatusOK,
	// 		Message: "Success",
	// 		Data:    artworks,
	// 	}
	// }
	responses := make([]*ArtworkResponseData, 0, len(artworks))
	for _, artwork := range artworks {
		responses = append(responses, ResponseDataFromArtwork(ctx, artwork))
	}
	return &common.RestfulCommonResponse[any]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    responses,
	}
}

type FetchedArtworkResponseData struct {
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	SourceURL   string                    `json:"source_url"`
	R18         bool                      `json:"r18"`
	Tags        []string                  `json:"tags"`
	Artist      *FetchedArtistResponse    `json:"artist"`
	SourceType  types.SourceType          `json:"source_type"`
	Pictures    []*FetchedPictureResponse `json:"pictures"`
}

type FetchedArtistResponse struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	UID      string `json:"uid"`
}

type FetchedPictureResponse struct {
	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	Index     uint   `json:"index"`
	Thumbnail string `json:"thumbnail"`
	Original  string `json:"original"`
	FileName  string `json:"file_name"`
}

func ResponseFromFetchedArtwork(artwork *types.Artwork) *common.RestfulCommonResponse[FetchedArtworkResponseData] {
	return &common.RestfulCommonResponse[FetchedArtworkResponseData]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    ResponseDataFromFetchedArtwork(artwork),
	}
}

func ResponseDataFromFetchedArtwork(artwork *types.Artwork) FetchedArtworkResponseData {
	pictures := make([]*FetchedPictureResponse, 0, len(artwork.Pictures))
	for _, picture := range artwork.Pictures {
		pictures = append(pictures, &FetchedPictureResponse{
			Width:     picture.Width,
			Height:    picture.Height,
			Index:     picture.Index,
			Thumbnail: picture.Thumbnail,
			Original:  picture.Original,
			FileName: func() string {
				fileName, err := sources.GetFileName(artwork, picture)
				if err != nil {
					return picture.GetFileName()
				}
				return fileName
			}(),
		})
	}
	return FetchedArtworkResponseData{
		Title:       artwork.Title,
		Description: artwork.Description,
		SourceURL:   artwork.SourceURL,
		R18:         artwork.R18,
		Tags:        artwork.Tags,
		Artist: &FetchedArtistResponse{
			Name:     artwork.Artist.Name,
			Username: artwork.Artist.Username,
			UID:      artwork.Artist.UID,
		},
		SourceType: artwork.SourceType,
		Pictures:   pictures,
	}
}
