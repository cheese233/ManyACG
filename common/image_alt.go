package common

import (
	"fmt"
	"math"
	"os"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/krau/ManyACG/types"
)

func CompressImageByVips(input *vips.ImageRef, maxEdgeLength int) {

	if input.Width() > int(maxEdgeLength) || input.Height() > int(maxEdgeLength) {
		if input.Width() > input.Height() {
			input.Resize(float64(maxEdgeLength)/float64(input.Width()), vips.KernelLanczos3)
		} else {
			input.Resize(float64(maxEdgeLength)/float64(input.Height()), vips.KernelLanczos3)
		}
	}
}

func CompressImageFile(input, output string, maxEdgeLength int, NearLossless bool) error {
	inputImage, err := vips.LoadImageFromFile(input, nil)
	if err != nil {
		return err
	}
	CompressImageByVips(inputImage, maxEdgeLength)
	param := vips.NewWebpExportParams()
	param.NearLossless = NearLossless
	outputBytes, _, err := inputImage.ExportWebp(param)
	if err != nil {
		return err
	}
	os.WriteFile(output, outputBytes, 0666)
	return nil
}

func CompressImageForTelegramFromBytes(input []byte) ([]byte, error) {
	inputImage, err := vips.LoadImageFromBuffer(input, nil)
	if err != nil {
		return nil, err
	}
	currentTotalSideLength := inputImage.Width() + inputImage.Height()
	inputLen := len(input)
	if currentTotalSideLength <= types.TelegramMaxPhotoTotalSideLength && inputLen <= types.TelegramMaxPhotoFileSize {
		return input, nil
	}
	scaleFactor := float64(types.TelegramMaxPhotoTotalSideLength) / float64(currentTotalSideLength)
	inputImage.Resize(scaleFactor, vips.KernelLanczos3)
	compressFactor := 0.9
	for {
		if compressFactor < 0.4 {
			return nil, fmt.Errorf("failed to compress image")
		}
		compressParam := vips.NewJpegExportParams()
		compressParam.Quality = int(math.Floor(100 * compressFactor))
		buf, _, err := inputImage.ExportJpeg(compressParam)
		if err != nil {
			return nil, err
		}
		if len(buf) > types.TelegramMaxPhotoFileSize {
			Logger.Debugf("recompressing...;current: compressed image size: %.2f MB, input size: %.2f MB, factor: %.2f;", float64(len(buf))/1024/1024, float64(inputLen)/1024/1024, compressFactor)
			compressFactor *= 0.8
			continue
		}
		return buf, nil
	}
}

func InitImage() {
	vips.Startup(nil)
	vips.LoggingSettings(VipsLogger, vips.LogLevelDebug)
}

func ShutdownImage() {
	vips.Shutdown()
}
