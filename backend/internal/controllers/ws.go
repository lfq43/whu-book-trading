package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"book-trading/backend/internal/models"
	"book-trading/backend/internal/utils"
	"book-trading/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsPacket struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type wsSendMessagePayload struct {
	ToUserID uint   `json:"to_user_id"`
	Content  string `json:"content"`
}

func Websocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "未提供认证令牌",
			Data:    nil,
		})
		return
	}

	claims, err := utils.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "无效的认证令牌: " + err.Error(),
			Data:    nil,
		})
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		UserID: claims.UserID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}
	ws.DefaultHub.Register(client)
	defer ws.DefaultHub.Unregister(claims.UserID)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var packet wsPacket
		if err := json.Unmarshal(msg, &packet); err != nil {
			continue
		}
		switch packet.Type {
		case "send_message":
			var payload wsSendMessagePayload
			if err := json.Unmarshal(packet.Data, &payload); err != nil {
				continue
			}
			message, err := createMessage(claims.UserID, payload.ToUserID, payload.Content)
			if err != nil {
				client.Send <- marshal(gin.H{"type": "error", "message": err.Error()})
				continue
			}

			client.Send <- marshal(gin.H{"type": "message_sent", "data": message})
			ws.DefaultHub.SendToUser(payload.ToUserID, gin.H{"type": "message_received", "data": message})
		case "ping":
			client.Send <- marshal(gin.H{"type": "pong"})
		}
	}
}

func marshal(payload interface{}) []byte {
	data, _ := json.Marshal(payload)
	return data
}
