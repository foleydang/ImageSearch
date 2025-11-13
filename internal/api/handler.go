package api

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/bytedance/ImageSearch/internal/model"
	"github.com/bytedance/ImageSearch/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Handler API处理器
type Handler struct {
	imageService service.ImageService
	imageDir     string
}

// NewHandler 创建API处理器
func NewHandler(imageService service.ImageService, imageDir string) *Handler {
	return &Handler{
		imageService: imageService,
		imageDir:     imageDir,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// 根路径处理程序
	router.GET("/", h.RootHandler)

	// 静态文件服务
	router.Static("/images", h.imageDir)

	// API路由组
	api := router.Group("/api")
	{
		// 图片相关路由
		images := api.Group("/images")
		{
			images.POST("", h.UploadImage)
			images.GET("", h.ListImages)
			images.GET("/:id", h.GetImage)
			images.DELETE("/:id", h.DeleteImage)
			images.POST("/search", h.SearchImages)
		}
	}

	// 健康检查
	router.GET("/health", h.HealthCheck)
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "ImageSearch API is running",
	})
}

// RootHandler 根路径处理程序
func (h *Handler) RootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":        "ImageSearch API",
		"version":     "1.0.0",
		"description": "基于Go语言的图片搜索系统API",
		"status":      "running",
		"message":     "请使用 /api 路径访问API接口，或使用 /health 检查服务状态",
		"api": map[string]interface{}{
			"base_url": "/api",
			"endpoints": map[string]interface{}{
				"images": map[string]string{
					"upload": "POST /api/images",
					"list":   "GET /api/images",
					"get":    "GET /api/images/:id",
					"delete": "DELETE /api/images/:id",
					"search": "POST /api/images/search",
				},
				"health": "GET /health",
			},
		},
	})
}

// UploadImage 上传图片
// @Summary 上传图片
// @Description 上传一张图片并生成嵌入向量
// @Tags 图片
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "图片文件"
// @Success 200 {object} model.Image
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/images [post]
func (h *Handler) UploadImage(c *gin.Context) {
	// 获取上传的文件
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		logrus.Errorf("获取上传文件失败: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "请选择要上传的图片文件",
		})
		return
	}
	defer file.Close()

	// 上传图片
	image, err := h.imageService.UploadImage(file, fileHeader)
	if err != nil {
		logrus.Errorf("上传图片失败: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, image)
}

// GetImage 获取图片
// @Summary 获取图片信息
// @Description 根据ID获取图片信息
// @Tags 图片
// @Produce json
// @Param id path string true "图片ID"
// @Success 200 {object} model.Image
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/images/{id} [get]
func (h *Handler) GetImage(c *gin.Context) {
	// 解析图片ID
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logrus.Errorf("解析图片ID失败: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "无效的图片ID",
		})
		return
	}

	// 获取图片信息
	image, err := h.imageService.GetImage(id)
	if err != nil {
		logrus.Errorf("获取图片信息失败: %v", err)
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "图片不存在",
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, image)
}

// ListImages 列出图片
// @Summary 列出图片
// @Description 分页列出所有图片
// @Tags 图片
// @Produce json
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页大小，默认10"
// @Success 200 {object} ListImagesResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/images [get]
func (h *Handler) ListImages(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 获取图片列表
	images, total, err := h.imageService.ListImages(page, pageSize)
	if err != nil {
		logrus.Errorf("获取图片列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "获取图片列表失败",
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, ListImagesResponse{
		Images: images,
		Total:  total,
		Page:   page,
		Size:   pageSize,
	})
}

// DeleteImage 删除图片
// @Summary 删除图片
// @Description 根据ID删除图片
// @Tags 图片
// @Produce json
// @Param id path string true "图片ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/images/{id} [delete]
func (h *Handler) DeleteImage(c *gin.Context) {
	// 解析图片ID
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logrus.Errorf("解析图片ID失败: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "无效的图片ID",
		})
		return
	}

	// 删除图片
	if err := h.imageService.DeleteImage(id); err != nil {
		logrus.Errorf("删除图片失败: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "删除图片失败",
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, SuccessResponse{
		Message: "图片删除成功",
	})
}

// SearchImages 搜索图片
// @Summary 搜索相似图片
// @Description 上传一张图片，搜索相似的图片
// @Tags 图片
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "搜索用的图片文件"
// @Success 200 {object} SearchImagesResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/images/search [post]
func (h *Handler) SearchImages(c *gin.Context) {
	// 获取上传的文件
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		logrus.Errorf("获取上传文件失败: %v", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "请选择要搜索的图片文件",
		})
		return
	}
	defer file.Close()

	// 搜索相似图片
	images, distances, err := h.imageService.SearchImagesByImage(file)
	if err != nil {
		logrus.Errorf("搜索相似图片失败: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	// 构建响应数据
	results := make([]SearchResult, len(images))
	for i, img := range images {
		results[i] = SearchResult{
			Image:    img,
			Distance: distances[i],
			// 生成图片URL
			ImageURL: "/images/" + filepath.Base(img.FilePath),
		}
	}

	// 返回结果
	c.JSON(http.StatusOK, SearchImagesResponse{
		Results: results,
		Total:   len(results),
	})
}

// 响应结构

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string `json:"message"`
}

// ListImagesResponse 图片列表响应
type ListImagesResponse struct {
	Images []model.Image `json:"images"`
	Total  int64         `json:"total"`
	Page   int           `json:"page"`
	Size   int           `json:"size"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Image    interface{} `json:"image"`
	Distance float32     `json:"distance"`
	ImageURL string      `json:"image_url"`
}

// SearchImagesResponse 图片搜索响应
type SearchImagesResponse struct {
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
}