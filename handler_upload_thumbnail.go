package main

import (
	"mime"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"crypto/rand"
	"encoding/base64"
	
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here

	
    // validate the request

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	// "thumbnail" should match the HTML form input name
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()
	
	// `file` is an `io.Reader` that we can read from to get the image data
	fileType,  _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type", err)
		return
	}
	if fileType != "image/jpeg" && fileType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Invalid file type", nil)
		return
	}


	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "vidieo not in database", err)
		return
	}

	if video.CreateVideoParams.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized user",err)
		return
	}

	fileExt := strings.Split(fileType, "/")
	if len(fileExt) != 2 {
		respondWithError(w, http.StatusBadRequest,"file type wrong", err)
		return
	}
	fileName := make([]byte, 32)
	rand.Read(fileName)
	vidFile := base64.RawURLEncoding.EncodeToString(fileName)


	tnfilepath := filepath.Join(cfg.assetsRoot,fmt.Sprintf("%s.%s",vidFile,fileExt[1]))
	
	tnFile, err := os.Create(tnfilepath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create file", err)
	}

	_, err = io.Copy(tnFile,file)
	tnURL := fmt.Sprintf("http://localhost:8091/assets/%s.%s",vidFile,fileExt[1])
	video.ThumbnailURL = &tnURL
	cfg.db.UpdateVideo(video)

	respondWithJSON(w, http.StatusOK, video)
}
