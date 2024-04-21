package webdav

import (
	"ManyACG-Bot/common"
	"ManyACG-Bot/config"
	. "ManyACG-Bot/logger"
	"ManyACG-Bot/types"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Webdav struct{}

func (w *Webdav) SavePicture(artwork *types.Artwork, picture *types.Picture) (*types.StorageInfo, error) {
	Logger.Debugf("saving picture %d of artwork %s", picture.Index, artwork.Title)
	fileName := artwork.Title + "_" + filepath.Base(picture.Original)
	fileDir := strings.TrimSuffix(config.Cfg.Storage.Webdav.Path, "/") + "/" + string(artwork.SourceType) + "/" + artwork.Artist.Name + "/"
	if err := Client.MkdirAll(fileDir, os.ModePerm); err != nil {
		if strings.Contains(err.Error(), "409") {
			fileDir = strings.TrimSuffix(config.Cfg.Storage.Webdav.Path, "/") + "/" + string(artwork.SourceType) + "/" + artwork.Artist.Username + "/"
			if err := Client.MkdirAll(fileDir, os.ModePerm); err != nil {
				return nil, err
			}
			fileName = filepath.Base(picture.Original)
		} else {
			return nil, err
		}
	}
	fileBytes, err := common.DownloadFromURL(picture.Original)
	if err != nil {
		return nil, err
	}

	filePath := fileDir + fileName
	if err := Client.Write(filePath, fileBytes, os.ModePerm); err != nil {
		return nil, err
	}
	Logger.Infof("picture %d of artwork %s saved to %s", picture.Index, artwork.Title, filePath)
	return &types.StorageInfo{
		Type: types.StorageTypeWebdav,
		Path: filePath,
	}, nil
}

func (w *Webdav) GetFile(info *types.StorageInfo) ([]byte, error) {
	Logger.Debugf("Getting file %s", info.Path)
	if config.Cfg.Storage.Webdav.CacheDir != "" {
		return w.GetFileWithCache(info)
	}
	return Client.Read(info.Path)
}

func (w *Webdav) GetFileWithCache(info *types.StorageInfo) ([]byte, error) {
	cacheDir := config.Cfg.Storage.Webdav.CacheDir
	cachePath := strings.TrimSuffix(cacheDir, "/") + "/" + filepath.Base(info.Path)
	data, err := os.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}
	data, err = Client.Read(info.Path)
	if err != nil {
		return nil, err
	}
	if err := common.MkFile(cachePath, data); err != nil {
		Logger.Errorf("failed to save cache file: %s", err)
	} else {
		go common.PurgeFileAfter(cachePath, time.Duration(config.Cfg.Storage.Webdav.CacheTTL)*time.Second)
	}
	return data, nil
}

func (w *Webdav) DeletePicture(info *types.StorageInfo) error {
	Logger.Debugf("deleting file %s", info.Path)
	return Client.Remove(info.Path)
}
