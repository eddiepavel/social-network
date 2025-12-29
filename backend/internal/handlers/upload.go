package handlers

import (
	"errors"
	"net/http"
	"social-network/app"
	"social-network/internal/middleware"
	"social-network/internal/models"
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

		image, err := app.File.UploadHandler(multiPartFile, user)

		defer multiPartFile.Close()

		if err != nil {
			utils.Internal(w, err)
			return
		}

		url := app.File.GenerateSignImage(image.Filename, user, image.ExpiresAt)

		response := models.FileResponse{
			UUId:      image.UUId,
			Filename:  image.Filename,
			ExpiresAt: image.ExpiresAt,
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
