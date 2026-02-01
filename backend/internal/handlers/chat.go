package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"social-network/app"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/models"
	"social-network/internal/utils"
	db_chat "social-network/pkg/db/queries/chat"
	db_followers "social-network/pkg/db/queries/followers"
	"social-network/pkg/db/sqlite"
	"time"

	"github.com/google/uuid"
)

func GetChatList(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		list, err := sqlite.NewQuery(app.DB).Chat.GetUserChatList(r.Context(),
			db_chat.GetUserChatListParams{
				SenderID: currentUserID,
				UserID:   currentUserID,
				UserID_2: currentUserID,
			})

		if err != nil {
			utils.Internal(w, err)
			return
		}

		if len(list) == 0 {
			utils.OK(w, []models.ChatList{})
			return
		}

		var listResponse []models.ChatList

		for _, chat := range list {
			roomID, _ := helpers.GenerateFromBytes(chat.RoomID)
			lastMessageID, _ := helpers.GenerateFromBytes(chat.LastMessageID)
			lastMessageSenderID, _ := helpers.GenerateFromBytes(chat.LastMessageSenderID)
			listResponse = append(listResponse, models.ChatList{
				RoomID: roomID,
				RoomName: func() *string {
					if chat.RoomName.Valid {
						return &chat.RoomName.String
					}
					return nil
				}(),
				IsGroup:             chat.IsGroup == 1,
				LastMessageID:       lastMessageID,
				LastMessageTime:     chat.LastMessageTime.Time,
				LastMessageContent:  chat.LastMessageContent,
				LastMessageSenderID: lastMessageSenderID,
				UnreadCount:         int(chat.UnreadCount),
			})
		}

		utils.OK(w, listResponse)
	}
}

func GetRoomMessages(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		roomIDHex := r.PathValue("roomId")
		if roomIDHex == "" {
			utils.BadRequest(w, errors.New("room ID required"))
			return
		}

		roomID, err := helpers.GenerateFromString(roomIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid room ID format"))
			return
		}

		cursorPagination := r.URL.Query().Get("beforeTime")
		cursorMessage := r.URL.Query().Get("beforeMessage")
		cursorTimestamp := time.Now()
		var cursorMessageID []byte
		limit := 20

		if cursorPagination != "" {
			cursorTimestamp, err = time.Parse(time.RFC3339, cursorPagination)
			if err != nil {
				utils.BadRequest(w, errors.New("invalid cursor format"))
			}
			if cursorMessage != "" {
				cursorMessageID, err = helpers.GenerateFromString(cursorMessage)
				app.Logger.Info("cursor message id", "cursorMessageID", cursorMessageID)
				if err != nil {
					utils.BadRequest(w, errors.New("invalid cursor message format"))
					return
				}
			} else {
				utils.BadRequest(w, errors.New("cursor message id required"))
				return
			}
		}

		isUserParticipant, err := sqlite.NewQuery(app.DB).Chat.CheckUserIsParticipant(r.Context(), db_chat.CheckUserIsParticipantParams{UserID: currentUserID, RoomID: roomID})
		if err != nil {
			utils.Internal(w, err)
			return
		}

		if isUserParticipant == 0 {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		chatMessages, err := sqlite.NewQuery(app.DB).Chat.GetRoomMessages(r.Context(), db_chat.GetRoomMessagesParams{
			TargetID: roomID,
			Column2: func() interface{} {
				if cursorPagination == "" {
					return nil
				}
				return cursorTimestamp
			}(),
			CreatedAt: sql.NullTime{Time: cursorTimestamp, Valid: true},
			Column4: func() interface{} {
				if cursorMessageID != nil {
					return cursorMessageID
				}
				return nil
			}(),
			MessageID: cursorMessageID,
			Limit:     int64(limit + 1),
		})

		if err != nil {
			utils.Internal(w, err)
			return
		}

		if len(chatMessages) == 0 {
			utils.OK(w, []models.ChatMessages{})
			return
		}

		hasMore := false
		if len(chatMessages) > limit {
			chatMessages = chatMessages[:limit]
			hasMore = true
		}

		var messages []models.ChatMessages
		var lastMessageTime time.Time
		var lastMessageID []byte
		for _, message := range chatMessages {
			messageID, _ := helpers.GenerateFromBytes(message.MessageID)
			senderID, _ := helpers.GenerateFromBytes(message.SenderID)
			messages = append(messages, models.ChatMessages{
				MessageID: messageID,
				Content:   message.Content,
				SenderID:  senderID,
				CreatedAt: message.CreatedAt.Time,
			})
			lastMessageTime = message.CreatedAt.Time
			lastMessageID = message.MessageID
		}

		var cursor models.CursorPagination

		if hasMore {
			lastMessageUUID, _ := helpers.GenerateFromBytes(lastMessageID)
			cursor = models.CursorPagination{
				CursorTimestamp: lastMessageTime,
				CursorID:        lastMessageUUID,
			}
		}

		response := models.ChatMessageResponse{
			Messages:   messages,
			HasMore:    hasMore,
			NextCursor: cursor,
		}

		err = sqlite.NewQuery(app.DB).Chat.MarkRoomMessagesAsRead(r.Context(), db_chat.MarkRoomMessagesAsReadParams{
			UserID: currentUserID,
			RoomID: roomID,
		})

		if err != nil {
			app.Logger.Error("could not mark room as read ", err.Error())
		}

		utils.OK(w, response)
	}
}

func CreateMessage(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		roomIDHex := r.PathValue("roomId")
		if roomIDHex == "" {
			utils.BadRequest(w, errors.New("room ID required"))
			return
		}

		roomID, err := helpers.GenerateFromString(roomIDHex)
		if err != nil {
			utils.BadRequest(w, errors.New("invalid room ID format"))
			return
		}

		isUserParticipant, err := sqlite.NewQuery(app.DB).Chat.CheckUserIsParticipant(r.Context(), db_chat.CheckUserIsParticipantParams{UserID: currentUserID, RoomID: roomID})
		if err != nil {
			utils.Internal(w, err)
			return
		}

		if isUserParticipant == 0 {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		var req models.CreateMessageRequest

		inputs := helpers.ValidateMessage.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)
		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		messageID, err := uuid.New().MarshalBinary()
		if err != nil {
			app.Logger.Error("failed uuid", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		err = sqlite.NewQuery(app.DB).Chat.CreateMessage(r.Context(), db_chat.CreateMessageParams{
			MessageID: messageID,
			Content:   req.Content,
			SenderID:  currentUserID,
			TargetID:  roomID,
		})

		if err != nil {
			utils.Internal(w, err)
			return
		}

		utils.OK(w, "message sent successfully")
	}
}

func CreateRoomAndMessage(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		var req models.FirstCreateMessageRequest

		inputs := helpers.ValidateFirstMessage.Build(r, app)

		ok, errValidation := utils.Validate(r, inputs, &req)
		if !ok {
			utils.Error(w, http.StatusUnprocessableEntity, "422", "validation error", errValidation)
			return
		}

		targetID, _ := helpers.GenerateFromString(req.TargetID)

		targetUser := helpers.FetchUser(app, targetID, r.Context(), w)

		if targetUser.UserID == nil {
			return
		}

		if !targetUser.IsPublic {
			follower := true
			followee := true

			_, err := sqlite.NewQuery(app.DB).Followers.CheckIfUserFollows(r.Context(), db_followers.CheckIfUserFollowsParams{
				FollowerID: currentUserID,
				FolloweeID: targetID,
			})

			if errors.Is(err, sql.ErrNoRows) {
				follower = false
			} else if err != nil {
				utils.Internal(w, err)
				return
			}

			_, err = sqlite.NewQuery(app.DB).Followers.CheckIfUserFollows(r.Context(), db_followers.CheckIfUserFollowsParams{
				FollowerID: targetID,
				FolloweeID: currentUserID,
			})

			if errors.Is(err, sql.ErrNoRows) {
				followee = false
			} else if err != nil {
				utils.Internal(w, err)
				return
			}

			if !follower && !followee {
				utils.BadRequest(w, errors.New("cannot send message to this user unless they follow you or you are following them"))
				return
			}
		}

		roomExists, err := sqlite.NewQuery(app.DB).Chat.FindRoomBetweenUsers(r.Context(), db_chat.FindRoomBetweenUsersParams{
			UserID:   currentUserID,
			UserID_2: targetID,
		})

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			app.Logger.Error("failed to find room between users", "error", err.Error())
			utils.Internal(w, err)
			return
		}

		if roomExists != nil {
			roomID, _ := helpers.GenerateFromBytes(roomExists)
			utils.Internal(w, errors.New(fmt.Sprintf("room already exists, ID: %v", roomID)))
			return
		}

		roomID, err := uuid.New().MarshalBinary()
		if err != nil {
			app.Logger.Error("failed uuid", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
		}

		messageID, err := uuid.New().MarshalBinary()
		if err != nil {
			app.Logger.Error("failed uuid", "error", err.Error())
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		err = sqlite.NewQuery(app.DB).Chat.CreateRoom(r.Context(), db_chat.CreateRoomParams{
			RoomID:  roomID,
			Name:    sql.NullString{Valid: false, String: ""},
			IsGroup: 0,
		})

		err = sqlite.NewQuery(app.DB).Chat.AddRoomParticipant(r.Context(), db_chat.AddRoomParticipantParams{
			RoomID: roomID,
			UserID: currentUserID,
		})

		err = sqlite.NewQuery(app.DB).Chat.AddRoomParticipant(r.Context(), db_chat.AddRoomParticipantParams{
			RoomID: roomID,
			UserID: targetID,
		})

		if err != nil {
			utils.Internal(w, err)
			return
		}

		err = sqlite.NewQuery(app.DB).Chat.CreateMessage(r.Context(), db_chat.CreateMessageParams{
			MessageID: messageID,
			Content:   req.Content,
			SenderID:  currentUserID,
			TargetID:  roomID,
		})

		if err != nil {
			utils.Internal(w, err)
			return
		}

		roomUUID, _ := helpers.GenerateFromBytes(roomID)
		utils.OK(w, map[string]string{"room_id": roomUUID})
	}
}
