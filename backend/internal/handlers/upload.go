package handlers

import (
	"errors"
	"net/http"
	"social-network/app"
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
		multiPartFile, _, _ := r.FormFile("file")

		image, err := services.NewFileService("./storage/uploads", multiPartFile).Upload()

		defer multiPartFile.Close()

		if err != nil {
			utils.Internal(w, err)
			return
		}

		utils.OK(w, image)
	}
}
