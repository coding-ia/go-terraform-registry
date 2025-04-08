package local_storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go-terraform-registry/internal/config"
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
	Config   config.RegistryConfig
	Endpoint string

	secretKey []byte
}

type AssetEndpoint struct {
	AssetPath string

	secretKey []byte
}

type LocalStorageAssetEndpoint interface {
	UploadFile(c *gin.Context)
	DownloadFile(c *gin.Context)
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

func (l *LocalStorage) ConfigureEndpoint(_ context.Context, routerGroup *gin.RouterGroup) {
	ae := &AssetEndpoint{
		secretKey: l.secretKey,
	}
	l.Endpoint = os.Getenv("LOCAL_STORAGE_ASSETS_ENDPOINT")
	if l.Endpoint == "" {
		l.Endpoint = "http://localhost:8080"
	}
	ae.AssetPath = os.Getenv("LOCAL_STORAGE_ASSETS_PATH")

	log.Printf("Local Storage Endpoint: %s", l.Endpoint)
	log.Printf("Local Storage Asset Path: %s", ae.AssetPath)

	routerGroup.PUT("/upload/:token", ae.UploadFile)
	routerGroup.HEAD("/download/:token/:file", ae.DownloadFile)
	routerGroup.GET("/download/:token/:file", ae.DownloadFile)
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

func (a *AssetEndpoint) UploadFile(c *gin.Context) {
	chunkNumber := c.GetHeader("Chunk-Number")

	if chunkNumber == "" {
		uploadFile(c, a.AssetPath, a.secretKey)
	} else {
		uploadFileChunk(c, a.AssetPath, a.secretKey)
	}
}

func (a *AssetEndpoint) DownloadFile(c *gin.Context) {
	tokenString := c.Param("token")

	token, err := jwt.ParseWithClaims(tokenString, &AssetClaims{}, func(token *jwt.Token) (interface{}, error) {
		return a.secretKey, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	claims, ok := token.Claims.(*AssetClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
		return
	}

	convertedPath := filepath.FromSlash(claims.Filename)
	joinedPath := filepath.Join(a.AssetPath, convertedPath)
	fileName := path.Base(joinedPath)

	if c.Request.Method == http.MethodHead {
		log.Printf("HEAD request for file: %s", joinedPath)
		fileInfo, err := os.Stat(joinedPath)
		if err != nil {
			log.Printf("File not found: %s", joinedPath)
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}

		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		c.Status(http.StatusOK)

		return
	}

	_, err = os.Stat(joinedPath)
	if err != nil {
		log.Printf("File not found: %s", joinedPath)
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.FileAttachment(joinedPath, fileName)
}

func uploadFile(c *gin.Context, assetPath string, secretKey []byte) {
	tokenString := c.Param("token")

	token, err := jwt.ParseWithClaims(tokenString, &AssetClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	claims, ok := token.Claims.(*AssetClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
		return
	}

	fileData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	convertedPath := filepath.FromSlash(claims.Filename)
	joinedPath := filepath.Join(assetPath, convertedPath)
	directoryPath := filepath.Dir(joinedPath)
	if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create provider directory"})
		return
	}
	if err := os.WriteFile(joinedPath, fileData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}
}

func uploadFileChunk(c *gin.Context, assetPath string, secretKey []byte) {
	tokenString := c.Param("token")

	token, err := jwt.ParseWithClaims(tokenString, &AssetClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	claims, ok := token.Claims.(*AssetClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
		return
	}

	chunkNumberStr := c.GetHeader("Chunk-Number")
	totalChunksStr := c.GetHeader("Total-Chunks")
	chunkFileName := c.GetHeader("File-Name")

	chunkNumber, err := strconv.Atoi(chunkNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Chunk-Number"})
		return
	}

	totalChunks, err := strconv.Atoi(totalChunksStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Total-Chunks"})
		return
	}

	chunkData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read chunk data"})
		return
	}

	convertedPath := filepath.FromSlash(claims.Filename)
	joinedPath := filepath.Join(assetPath, convertedPath)
	directoryPath := filepath.Dir(joinedPath)

	chunkedFileName := fmt.Sprintf("%s.part%d", chunkFileName, chunkNumber)
	chunkedFilePath := filepath.Join(directoryPath, chunkedFileName)

	if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create provider directory"})
		return
	}
	if err := os.WriteFile(chunkedFilePath, chunkData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	if chunkNumber == totalChunks {
		err = assembleFile(directoryPath, chunkFileName, totalChunks)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assemble file"})
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
