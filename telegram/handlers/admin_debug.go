package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/krau/ManyACG/adapter"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func DumpArtworkInfo(ctx context.Context, bot *telego.Bot, message telego.Message) {
	userAdmin, err := service.GetAdminByUserID(ctx, message.From.ID)
	if err != nil {
		common.Logger.Errorf("获取管理员信息失败: %s", err)
		utils.ReplyMessage(bot, message, "获取管理员信息失败")
		return
	}
	if userAdmin == nil {
		utils.ReplyMessage(bot, message, "你没有权限执行此操作")
		return
	}
	helpText := fmt.Sprintf(`
[管理员] <b>使用 /dump 命令回复一条包含作品链接的消息, 将获取作品信息并以JSON格式回复</b>
	
命令语法: %s

若不提供参数, 默认获取所有信息
			`, common.EscapeHTML("/dump [tags] [artist] [pictures]"))
	if message.ReplyToMessage == nil {
		utils.ReplyMessageWithHTML(bot, message, helpText)
		return
	}
	sourceURL := utils.FindSourceURLForMessage(message.ReplyToMessage)
	if sourceURL == "" {
		utils.ReplyMessageWithHTML(bot, message, "回复的消息中没有支持的链接, 命令帮助:\n"+helpText)
		return
	}
	_, _, args := telegoutil.ParseCommand(message.Text)
	adapterOpt := &types.AdapterOption{}
	if len(args) > 0 {
		for _, arg := range args {
			switch arg {
			case "tags":
				adapterOpt.LoadTag = true
			case "artist":
				adapterOpt.LoadArtist = true
			case "pictures":
				adapterOpt.LoadPicture = true
			}
		}
	} else {
		adapterOpt = adapter.LoadAll()
	}

	artwork, err := service.GetArtworkByURL(ctx, sourceURL, adapterOpt)
	if err != nil {
		common.Logger.Errorf("获取作品信息失败: %s", err)
		utils.ReplyMessageWithHTML(bot, message, fmt.Sprintf("获取作品信息失败\n<code>%s</code>", common.EscapeHTML(err.Error())))
		return
	}
	artworkJSON, err := json.MarshalIndent(artwork, "", "  ")
	if err != nil {
		common.Logger.Errorf("序列化作品信息失败: %s", err)
		utils.ReplyMessageWithHTML(bot, message, fmt.Sprintf("序列化作品信息失败\n<code>%s</code>", common.EscapeHTML(err.Error())))
		return
	}
	if _, err := utils.ReplyMessageWithHTML(bot, message, "<pre>"+common.EscapeHTML(string(artworkJSON))+"</pre>"); err != nil {
		utils.ReplyMessageWithHTML(bot, message, fmt.Sprintf("回复消息失败\n<code>%s</code>", common.EscapeHTML(err.Error())))
	}
}
