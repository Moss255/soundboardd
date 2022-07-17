package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type File struct {
	ID        int
	Filename  string
	Fileurl   string
	Filegroup int
}

type ResponseMessage struct {
	Message string `json:"message"`
}

var db *gorm.DB

func setupDB() {

	dbUser := os.Getenv("SQL_USER")
	dbPass := os.Getenv("SQL_PASS")
	dbServer := os.Getenv("SQL_SERVER")
	dbDatabase := os.Getenv("SQL_DATABASE")

	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True", dbUser, dbPass, dbServer, dbDatabase)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}
}

func WriteResponse(w http.ResponseWriter, status int, data any) {
	switch data.(type) {
	case string:
		var message ResponseMessage
		message.Message = fmt.Sprintf("%v", data)
		data = message
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func fetchMinioFile(fileName string) ([]byte, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	fileErr := minioClient.FGetObject(ctx, os.Getenv("MINIO_BUCKET"), fileName, fileName, minio.GetObjectOptions{})

	if fileErr != nil {
		return nil, fileErr
	}

	body, err := os.ReadFile(fileName)

	if err != nil {
		return nil, err
	}

	return body, nil

}

func handlePlayback(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:

		Id := r.URL.Query().Get("ID")

		var file File
		db.First(&file, "ID = ?", Id)
		if file.ID == 0 {
			WriteResponse(w, http.StatusBadRequest, "Cannot find file with Id")
			return
		}

		contents, err := fetchMinioFile(file.Fileurl)

		if err != nil {
			WriteResponse(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Write(contents)
	}
}

func handleFiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:

		Id := r.URL.Query().Get("ID")

		if Id != "" {
			var file File
			db.First(&file, "ID = ?", Id)
			if file.ID == 0 {
				WriteResponse(w, http.StatusBadRequest, "Cannot find file with Id")
				return
			}
			WriteResponse(w, http.StatusOK, file)
			return
		}

		var files []File
		db.Find(&files)
		if len(files) > 0 {
			WriteResponse(w, http.StatusOK, files)
			return
		}

		WriteResponse(w, http.StatusBadRequest, "Cannot find file with Id")
	case http.MethodPost:
		defer r.Body.Close()

		var file File
		json.NewDecoder(r.Body).Decode(&file)
		db.Create(&file)

		WriteResponse(w, http.StatusOK, "Added file")
	case http.MethodPut:
		defer r.Body.Close()

		var file File
		json.NewDecoder(r.Body).Decode(&file)
		db.Save(&file)

		WriteResponse(w, http.StatusOK, "Updated file")
	}
}

func main() {

	setupDB()

	http.HandleFunc("/files", handleFiles)

	http.HandleFunc("/play", handlePlayback)

	fmt.Println("Listening on Port 8080")

	http.ListenAndServe(":8080", nil)
}
