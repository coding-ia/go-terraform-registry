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
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func (l *LocalStorage) ConfigureEndpoint(ctx context.Context, routerGroup *gin.RouterGroup) {
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
	routerGroup.GET("/download/:token", ae.DownloadFile)
}

func (l *LocalStorage) ConfigureStorage(ctx context.Context) error {
	secretKey, err := generateRandomSecret(32)
	l.secretKey = []byte(secretKey)

	log.Println("Using local storage for providers & endpoints.")

	return err
}

func (l *LocalStorage) GenerateUploadURL(ctx context.Context, path string) (string, error) {
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

func (l *LocalStorage) GenerateDownloadURL(ctx context.Context, path string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, AssetClaims{
		Filename: path,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)), // 15-minute expiry
		},
	})

	signedToken, err := token.SignedString(l.secretKey)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/asset/download/%s", l.Endpoint, signedToken)

	return url, nil
}

func (a *AssetEndpoint) UploadFile(c *gin.Context) {
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

	fileData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	convertedPath := filepath.FromSlash(claims.Filename)
	joinedPath := filepath.Join(a.AssetPath, convertedPath)
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

	c.File(joinedPath)
}

func generateRandomSecret(n int) (string, error) {
	bytes := make([]byte, n/2) // Each byte is 2 hex chars
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
