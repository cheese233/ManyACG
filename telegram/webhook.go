package telegram

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mymmrac/telego"
)

func WebhookGin(router gin.IRouter, path string, secretToken string) func(handler telego.WebhookHandler) error {
	return func(handler telego.WebhookHandler) error {
		router.POST(path, func(c *gin.Context) {
			if c.GetHeader(telego.WebhookSecretTokenHeader) != secretToken {
				c.Status(http.StatusUnauthorized)
				return
			}

			body, err := c.GetRawData()
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}

			if err := handler(c, body); err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}

			c.Status(http.StatusOK)
		})

		return nil
	}
}
