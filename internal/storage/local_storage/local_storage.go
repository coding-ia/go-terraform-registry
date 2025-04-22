package local_storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"go-terraform-registry/internal/config"
	"go-terraform-registry/internal/response"
	"go-terraform-registry/internal/storage"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

var _ storage.RegistryProviderStorage = &LocalStorage{}
var _ storage.RegistryProviderStorageAssetEndpoint = &LocalStorage{}
var _ LocalStorageAssetEndpoint = &AssetEndpoint{}

type LocalStorage struct {
	AssetPath string
	Config    config.RegistryConfig
	Endpoint  string

	secretKey []byte
}

type AssetEndpoint struct {
	AssetPath string

	secretKey []byte
}

type LocalStorageAssetEndpoint interface {
	UploadFile(http.ResponseWriter, *http.Request)
	DownloadFile(http.ResponseWriter, *http.Request)
}

func NewLocalStorage(config config.RegistryConfig) storage.RegistryProviderStorage {
	return &LocalStorage{
		Config: config,
	}
}

type AssetClaims struct {
	Filename string `json:"filename"`
	jwt.RegisteredClaims
}

func (l *LocalStorage) ConfigureEndpoint(_ context.Context, cr *chi.Mux) {
	ae := &AssetEndpoint{
		secretKey: l.secretKey,
	}
	l.Endpoint = os.Getenv("LOCAL_STORAGE_ASSETS_ENDPOINT")
	if l.Endpoint == "" {
		l.Endpoint = "http://localhost:8080"
	}
	ae.AssetPath = os.Getenv("LOCAL_STORAGE_ASSETS_PATH")
	l.AssetPath = ae.AssetPath

	log.Printf("Local Storage Endpoint: %s", l.Endpoint)
	log.Printf("Local Storage Asset Path: %s", ae.AssetPath)

	cr.Route("/asset", func(r chi.Router) {
		r.Put("/upload/{token}", ae.UploadFile)
		r.Head("/download/{token}/{file}", ae.DownloadFile)
		r.Get("/download/{token}/{file}", ae.DownloadFile)
	})
}

func (l *LocalStorage) ConfigureStorage(_ context.Context) error {
	secretKey, err := generateRandomSecret(32)
	l.secretKey = []byte(secretKey)

	log.Println("Using local storage for providers & endpoints.")

	return err
}

func (l *LocalStorage) GenerateUploadURL(_ context.Context, path string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, AssetClaims{
		Filename: path,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15-minute expiry
		},
	})

	signedToken, err := token.SignedString(l.secretKey)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/asset/upload/%s", l.Endpoint, signedToken)

	return url, nil
}

func (l *LocalStorage) GenerateDownloadURL(_ context.Context, filePath string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, AssetClaims{
		Filename: filePath,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)), // 15-minute expiry
		},
	})

	signedToken, err := token.SignedString(l.secretKey)
	if err != nil {
		return "", err
	}

	filename := path.Base(filePath)
	url := fmt.Sprintf("%s/asset/download/%s/%s", l.Endpoint, signedToken, filename)

	return url, nil
}

func (a *AssetEndpoint) UploadFile(w http.ResponseWriter, r *http.Request) {
	chunkNumber := r.Header.Get("Chunk-Number")

	if chunkNumber == "" {
		uploadFile(w, r, a.AssetPath, a.secretKey)
	} else {
		uploadFileChunk(w, r, a.AssetPath, a.secretKey)
	}
}

func (a *AssetEndpoint) DownloadFile(w http.ResponseWriter, r *http.Request) {
	tokenString := chi.URLParam(r, "token")
	_ = chi.URLParam(r, "file")

	token, err := jwt.ParseWithClaims(tokenString, &AssetClaims{}, func(token *jwt.Token) (interface{}, error) {
		return a.secretKey, nil
	})
	if err != nil || !token.Valid {
		response.JsonResponse(w, http.StatusUnauthorized, response.ErrorResponse{
			Error: "invalid or expired token",
		})
		return
	}

	claims, ok := token.Claims.(*AssetClaims)
	if !ok {
		response.JsonResponse(w, http.StatusUnauthorized, response.ErrorResponse{
			Error: "invalid claims",
		})
		return
	}

	convertedPath := filepath.FromSlash(claims.Filename)
	joinedPath := filepath.Join(a.AssetPath, convertedPath)
	fileName := path.Base(joinedPath)

	if r.Method == http.MethodHead {
		log.Printf("HEAD request for file: %s", joinedPath)
		fileInfo, err := os.Stat(joinedPath)
		if err != nil {
			log.Printf("File not found: %s", joinedPath)
			response.JsonResponse(w, http.StatusNotFound, response.ErrorResponse{
				Error: "file not found",
			})
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		w.WriteHeader(http.StatusOK)

		return
	}

	_, err = os.Stat(joinedPath)
	if err != nil {
		log.Printf("File not found: %s", joinedPath)
		response.JsonResponse(w, http.StatusNotFound, response.ErrorResponse{
			Error: "file not found",
		})
		return
	}

	response.FileResponse(w, r, joinedPath, fileName)
}

func (l *LocalStorage) RemoveFile(_ context.Context, path string) error {
	log.Printf("Removing file: %s", path)

	convertedPath := filepath.FromSlash(path)
	joinedPath := filepath.Join(l.AssetPath, convertedPath)

	_, err := os.Stat(joinedPath)
	if err == nil {
		return os.Remove(joinedPath)
	}

	return err
}

func uploadFile(w http.ResponseWriter, r *http.Request, assetPath string, secretKey []byte) {
	tokenString := chi.URLParam(r, "token")

	token, err := jwt.ParseWithClaims(tokenString, &AssetClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		response.JsonResponse(w, http.StatusUnauthorized, response.ErrorResponse{
			Error: "invalid or expired token",
		})
		return
	}

	claims, ok := token.Claims.(*AssetClaims)
	if !ok {
		response.JsonResponse(w, http.StatusUnauthorized, response.ErrorResponse{
			Error: "invalid claims",
		})
		return
	}

	fileData, err := io.ReadAll(r.Body)
	if err != nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "failed to read file",
		})
		return
	}

	convertedPath := filepath.FromSlash(claims.Filename)
	joinedPath := filepath.Join(assetPath, convertedPath)
	directoryPath := filepath.Dir(joinedPath)
	if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "unable to create provider directory",
		})
		return
	}
	if err := os.WriteFile(joinedPath, fileData, 0644); err != nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "failed to save file",
		})
		return
	}
}

func uploadFileChunk(w http.ResponseWriter, r *http.Request, assetPath string, secretKey []byte) {
	tokenString := chi.URLParam(r, "token")

	token, err := jwt.ParseWithClaims(tokenString, &AssetClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		response.JsonResponse(w, http.StatusUnauthorized, response.ErrorResponse{
			Error: "invalid or expired token",
		})
		return
	}

	claims, ok := token.Claims.(*AssetClaims)
	if !ok {
		response.JsonResponse(w, http.StatusUnauthorized, response.ErrorResponse{
			Error: "invalid claims",
		})
		return
	}

	filePath := claims.Filename
	fileName := path.Base(filePath)

	chunkNumberStr := r.Header.Get("Chunk-Number")
	totalChunksStr := r.Header.Get("Total-Chunks")

	chunkNumber, err := strconv.Atoi(chunkNumberStr)
	if err != nil {
		response.JsonResponse(w, http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid Chunk-Number",
		})
		return
	}

	totalChunks, err := strconv.Atoi(totalChunksStr)
	if err != nil {
		response.JsonResponse(w, http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid Total-Chunks",
		})
		return
	}

	chunkData, err := io.ReadAll(r.Body)
	if err != nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "failed to read chunk data",
		})
		return
	}

	convertedPath := filepath.FromSlash(filePath)
	joinedPath := filepath.Join(assetPath, convertedPath)
	directoryPath := filepath.Dir(joinedPath)

	chunkedFileName := fmt.Sprintf("%s.part%d", fileName, chunkNumber)
	chunkedFilePath := filepath.Join(directoryPath, chunkedFileName)

	if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "unable to create provider directory",
		})
		return
	}
	if err := os.WriteFile(chunkedFilePath, chunkData, 0644); err != nil {
		response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "failed to save file",
		})
		return
	}

	if chunkNumber == totalChunks {
		err = assembleFile(directoryPath, fileName, totalChunks)
		if err != nil {
			response.JsonResponse(w, http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to assemble file",
			})
			return
		}
	}
}

func assembleFile(directoryPath string, fileName string, totalChunks int) error {
	fullPath := filepath.Join(directoryPath, fileName)
	assembledFile, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer func(assembledFile *os.File) {
		err := assembledFile.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
		}
	}(assembledFile)

	for i := 1; i <= totalChunks; i++ {
		err := processChunk(assembledFile, fileName, directoryPath, i)
		if err != nil {
			return err
		}
	}

	return nil
}

func processChunk(assembledFile *os.File, fileName string, directoryPath string, chunkNumber int) error {
	chunkedFileName := fmt.Sprintf("%s.part%d", fileName, chunkNumber)
	chunkedPath := filepath.Join(directoryPath, chunkedFileName)
	chunkFile, err := os.Open(chunkedPath)
	if err != nil {
		return err
	}
	defer func(chunkFile *os.File, path string) {
		err := chunkFile.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
		}
		err = os.Remove(path)
		if err != nil {
			fmt.Println("Error removing file:", err)
		}
	}(chunkFile, chunkedPath)

	_, err = io.Copy(assembledFile, chunkFile)
	if err != nil {
		return err
	}

	return nil
}

func generateRandomSecret(n int) (string, error) {
	bytes := make([]byte, n/2) // Each byte is 2 hex chars
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
