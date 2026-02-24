package handlers

import (
	"errors"
	"net/http"
	"social-network/app"
	contextkeys "social-network/internal/contextKeys"
	"social-network/internal/helpers"
	"social-network/internal/middleware"
	"social-network/internal/models"
	"social-network/internal/services"
	"social-network/internal/utils"
)

const maxUpload = 5 << 20

func Upload(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

		err := r.ParseMultipartForm(maxUpload)
		if err != nil {
			if err.Error() == "http: request body too large" {
				utils.Error(w, http.StatusRequestEntityTooLarge, "413", "entity too large", "max upload size is 5mb")
				return
			}
			utils.BadRequest(w, errors.New("badss request"))
			return
		}

		user, ok := middleware.GetUserIDFromContext(r.Context())

		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		multiPartFile, _, _ := r.FormFile("file")

		image, err := app.File.UploadHandler(multiPartFile, user, app.File.BasePath)

		defer multiPartFile.Close()

		if err != nil {
			utils.Internal(w, err)
			return
		}

		privateFile, ok := image.(*services.File)

		if !ok {
			utils.Internal(w, errors.New("internal server error"))
			return
		}

		url := app.File.GenerateSignImage(privateFile.Filename, user, privateFile.ExpiresAt)

		response := models.FileResponse{
			UUId:      privateFile.UUId,
			Filename:  privateFile.Filename,
			ExpiresAt: privateFile.ExpiresAt,
			Url:       url,
		}

		utils.OK(w, response)
	}
}

func GetImage(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		value := r.PathValue("image")

		if value == "" {
			utils.BadRequest(w, errors.New("bad request: missing image"))
			return
		}

		user, ok := middleware.GetUserIDFromContext(r.Context())

		if !ok {
			utils.Unauthorized(w, "missing user")
			return
		}

		validate, err := app.File.ValidateImageSign(r, user)

		if err != nil {
			utils.BadRequest(w, err)
			return
		}

		if validate {
			path := app.File.BasePath + "/" + value
			http.ServeFile(w, r, path)
			return
		}

		utils.BadRequest(w, errors.New("wrong sign"))
	}
}

func GetPublicImage(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		value := r.PathValue("image")

		if value == "" {
			utils.BadRequest(w, errors.New("bad request: missing image"))
			return
		}

		guestUser, ok := r.Context().Value(contextkeys.GuestSession).(string)

		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		if !ok {
			utils.Unauthorized(w, "missing user")
			return
		}

		gGuest, err := helpers.GenerateFromString(guestUser)

		validate, err := app.File.ValidateImageSign(r, gGuest)

		if err != nil {
			utils.BadRequest(w, err)
			return
		}

		if validate {
			path := app.File.PublicPath + "/" + value
			http.ServeFile(w, r, path)
			return
		}

		utils.BadRequest(w, errors.New("wrong sign"))
	}
}

func UploadPublic(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

		err := r.ParseMultipartForm(maxUpload)
		if err != nil {
			if err.Error() == "http: request body too large" {
				utils.Error(w, http.StatusRequestEntityTooLarge, "413", "entity too large", "max upload size is 5mb")
				return
			}
			utils.BadRequest(w, errors.New("badss request"))
			return
		}

		guestUser, ok := r.Context().Value(contextkeys.GuestSession).(string)

		if !ok {
			utils.Unauthorized(w, "Unauthorized")
			return
		}

		multiPartFile, _, _ := r.FormFile("file")

		userUuid, err := helpers.GenerateFromString(guestUser)

		if err != nil {
			utils.BadRequest(w, errors.New("guest session not found"))
			return
		}

		image, err := app.File.UploadHandler(multiPartFile, userUuid, app.File.PublicPath)

		defer multiPartFile.Close()

		if err != nil {
			utils.Internal(w, err)
			return
		}

		publicFile, ok := image.(*services.PublicFile)

		if !ok {
			utils.Internal(w, errors.New("internal"))
			return
		}

		url := app.File.GeneratePublicSignImage(publicFile.Filename, publicFile.GuestSession, publicFile.ExpiresAt)

		response := models.FileResponse{
			UUId:      publicFile.UUId,
			Filename:  publicFile.Filename,
			ExpiresAt: publicFile.ExpiresAt,
			Url:       url,
		}

		utils.OK(w, response)
	}
}
