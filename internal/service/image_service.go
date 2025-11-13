package service

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bytedance/ImageSearch/internal/model"
	"github.com/bytedance/ImageSearch/internal/repository"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"github.com/sirupsen/logrus"
)

// ImageService 图片服务接口
type ImageService interface {
	UploadImage(file multipart.File, fileHeader *multipart.FileHeader) (*model.Image, error)
	GetImage(id uuid.UUID) (*model.Image, error)
	ListImages(page, pageSize int) ([]model.Image, int64, error)
	DeleteImage(id uuid.UUID) error
	SearchImagesByImage(file multipart.File) ([]model.Image, []float32, error)
}

// imageService 图片服务实现
type imageService struct {
	imageRepo repository.ImageRepository
	imageDir  string
}

// NewImageService 创建图片服务
func NewImageService(imageRepo repository.ImageRepository, imageDir string) ImageService {
	// 确保图片目录存在
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		logrus.Errorf("创建图片目录失败: %v", err)
		panic(err)
	}

	return &imageService{
		imageRepo: imageRepo,
		imageDir:  imageDir,
	}
}

// UploadImage 上传图片
func (s *imageService) UploadImage(file multipart.File, fileHeader *multipart.FileHeader) (*model.Image, error) {
	// 读取文件内容
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		logrus.Errorf("读取文件内容失败: %v", err)
		return nil, err
	}

	// 重置文件指针
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logrus.Errorf("重置文件指针失败: %v", err)
		return nil, err
	}

	// 解码图片
	img, format, err := image.Decode(buffer)
	if err != nil {
		logrus.Errorf("解码图片失败: %v", err)
		return nil, err
	}

	// 检查图片格式
	if format != "jpeg" && format != "png" {
		logrus.Errorf("不支持的图片格式: %s", format)
		return nil, errors.New("只支持 JPEG 和 PNG 格式的图片")
	}

	// 生成唯一文件名
	extension := strings.ToLower(format)
	fileName := fmt.Sprintf("%s.%s", uuid.New().String(), extension)
	filePath := filepath.Join(s.imageDir, fileName)

	// 保存图片
	dst, err := os.Create(filePath)
	if err != nil {
		logrus.Errorf("创建图片文件失败: %v", err)
		return nil, err
	}
	defer dst.Close()

	// 调整图片大小（可选，根据实际需求调整）
	resizedImg := resize.Resize(800, 0, img, resize.Lanczos3)

	// 保存调整后的图片
	switch extension {
	case "jpg", "jpeg":
		if err := jpeg.Encode(dst, resizedImg, &jpeg.Options{Quality: 90}); err != nil {
			logrus.Errorf("保存 JPEG 图片失败: %v", err)
			return nil, err
		}
	case "png":
		if err := png.Encode(dst, resizedImg); err != nil {
			logrus.Errorf("保存 PNG 图片失败: %v", err)
			return nil, err
		}
	}

	// 获取图片信息
	bounds := resizedImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 获取文件大小
	fileInfo, err := dst.Stat()
	if err != nil {
		logrus.Errorf("获取文件信息失败: %v", err)
		return nil, err
	}
	size := fileInfo.Size()

	// 创建图片记录
	image := &model.Image{
		FileName:  fileHeader.Filename,
		FilePath:  filePath,
		Extension: extension,
		Width:     width,
		Height:    height,
		Size:      size,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存图片记录到数据库
	if err := s.imageRepo.CreateImage(image); err != nil {
		logrus.Errorf("保存图片记录失败: %v", err)
		// 删除已保存的图片文件
		os.Remove(filePath)
		return nil, err
	}

	// 生成图片嵌入向量（这里使用简化的实现，实际应该使用预训练模型）
	embedding := s.generateEmbedding(resizedImg)

	// 保存嵌入向量
	imageEmbedding := &model.ImageEmbedding{
		ImageID:   image.ID,
		Embedding: embedding,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.imageRepo.CreateImageEmbedding(imageEmbedding); err != nil {
		logrus.Errorf("保存图片嵌入向量失败: %v", err)
		// 删除已保存的图片文件和记录
		os.Remove(filePath)
		s.imageRepo.DeleteImage(image.ID)
		return nil, err
	}

	logrus.Infof("图片上传成功: %s", image.FileName)
	return image, nil
}

// GetImage 根据ID获取图片
func (s *imageService) GetImage(id uuid.UUID) (*model.Image, error) {
	return s.imageRepo.GetImageByID(id)
}

// ListImages 列出图片
func (s *imageService) ListImages(page, pageSize int) ([]model.Image, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 调用仓库层方法，获取指针切片
	imagePtrs, total, err := s.imageRepo.ListImages(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 转换为值切片
	images := make([]model.Image, len(imagePtrs))
	for i, imgPtr := range imagePtrs {
		images[i] = *imgPtr
	}

	return images, total, nil
}

// DeleteImage 删除图片
func (s *imageService) DeleteImage(id uuid.UUID) error {
	// 获取图片信息
	image, err := s.imageRepo.GetImageByID(id)
	if err != nil {
		return err
	}

	// 删除图片文件
	if err := os.Remove(image.FilePath); err != nil {
		logrus.Errorf("删除图片文件失败: %v", err)
		return err
	}

	// 删除图片记录和嵌入向量
	if err := s.imageRepo.DeleteImage(id); err != nil {
		logrus.Errorf("删除图片记录失败: %v", err)
		return err
	}

	logrus.Infof("图片删除成功: %s", image.FileName)
	return nil
}

// SearchImagesByImage 根据图片搜索相似图片
func (s *imageService) SearchImagesByImage(file multipart.File) ([]model.Image, []float32, error) {
	// 读取文件内容
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		logrus.Errorf("读取文件内容失败: %v", err)
		return nil, nil, err
	}

	// 重置文件指针
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logrus.Errorf("重置文件指针失败: %v", err)
		return nil, nil, err
	}

	// 解码图片
	img, _, err := image.Decode(buffer)
	if err != nil {
		logrus.Errorf("解码图片失败: %v", err)
		return nil, nil, err
	}

	// 调整图片大小
	resizedImg := resize.Resize(800, 0, img, resize.Lanczos3)

	// 生成嵌入向量
	embedding := s.generateEmbedding(resizedImg)

	// 搜索相似图片
	imagePtrs, distances, err := s.imageRepo.SearchSimilarImages(embedding, 10)
	if err != nil {
		logrus.Errorf("搜索相似图片失败: %v", err)
		return nil, nil, err
	}

	// 转换为值切片
	images := make([]model.Image, len(imagePtrs))
	for i, imgPtr := range imagePtrs {
		images[i] = *imgPtr
	}

	logrus.Infof("搜索到 %d 张相似图片", len(images))
	return images, distances, nil
}

// generateEmbedding 生成图片嵌入向量（简化实现）
// 注意：这里使用的是非常简化的实现，实际生产环境中应该使用预训练的深度学习模型
func (s *imageService) generateEmbedding(img image.Image) []float32 {
	// 这里使用一个简单的实现，实际应该使用预训练模型
	// 例如：使用 Go 绑定的 TensorFlow 或 PyTorch 模型
	
	// 简化实现：计算图片的平均颜色作为嵌入向量
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var r, g, b float32
	count := width * height

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r1, g1, b1, _ := img.At(x, y).RGBA()
			r += float32(r1) / 65535.0
			g += float32(g1) / 65535.0
			b += float32(b1) / 65535.0
		}
	}

	r /= float32(count)
	g /= float32(count)
	b /= float32(count)

	// 返回一个简单的 3 维嵌入向量
	// 实际应用中，嵌入向量的维度应该更高（例如 512 维或 1024 维）
	return []float32{r, g, b}
}