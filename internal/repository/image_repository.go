package repository

import (
	"encoding/json"
	"math"
	"sort"
	"time"

	"github.com/bytedance/ImageSearch/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ImageRepository 图片仓库接口
type ImageRepository interface {
	CreateImage(image *model.Image) error
	GetImageByID(id uuid.UUID) (*model.Image, error)
	ListImages(page, pageSize int) ([]*model.Image, int64, error)
	DeleteImage(id uuid.UUID) error
	CreateImageEmbedding(embedding *model.ImageEmbedding) error
	GetImageEmbeddingByImageID(imageID uuid.UUID) (*model.ImageEmbedding, error)
	SearchSimilarImages(targetEmbedding []float32, limit int) ([]*model.Image, []float32, error)
}

// imageRepository 图片仓库实现
type imageRepository struct {
	DB *gorm.DB
}

// NewImageRepository 创建图片仓库
func NewImageRepository(db *Database) ImageRepository {
	return &imageRepository{
		DB: db.DB,
	}
}

// CreateImage 创建图片记录
func (r *imageRepository) CreateImage(image *model.Image) error {
	return r.DB.Create(image).Error
}

// GetImageByID 根据ID获取图片
func (r *imageRepository) GetImageByID(id uuid.UUID) (*model.Image, error) {
	var image model.Image
	result := r.DB.First(&image, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &image, nil
}

// ListImages 列出图片
func (r *imageRepository) ListImages(page, pageSize int) ([]*model.Image, int64, error) {
	var images []*model.Image
	var total int64

	// 计算总数
	r.DB.Model(&model.Image{}).Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	result := r.DB.Offset(offset).Limit(pageSize).Find(&images)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return images, total, nil
}

// DeleteImage 删除图片
func (r *imageRepository) DeleteImage(id uuid.UUID) error {
	// 开启事务
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// 删除图片嵌入向量
		if err := tx.Where("image_id = ?", id).Delete(&model.ImageEmbedding{}).Error; err != nil {
			return err
		}

		// 删除图片记录
		if err := tx.Delete(&model.Image{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

// CreateImageEmbedding 创建图片嵌入向量
func (r *imageRepository) CreateImageEmbedding(embedding *model.ImageEmbedding) error {
	// 将浮点数数组转换为JSON字符串
	embeddingJSON, err := json.Marshal(embedding.Embedding)
	if err != nil {
		return err
	}
	
	// 创建一个临时结构体用于数据库操作
	type TempEmbedding struct {
		ID        uuid.UUID
		ImageID   uuid.UUID
		Embedding []byte
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	
	temp := TempEmbedding{
		ID:        embedding.ID,
		ImageID:   embedding.ImageID,
		Embedding: embeddingJSON,
		CreatedAt: embedding.CreatedAt,
		UpdatedAt: embedding.UpdatedAt,
	}
	
	return r.DB.Table("image_embeddings").Create(&temp).Error
}

// GetImageEmbeddingByImageID 根据图片ID获取嵌入向量
func (r *imageRepository) GetImageEmbeddingByImageID(imageID uuid.UUID) (*model.ImageEmbedding, error) {
	// 使用临时结构体查询
	type TempEmbedding struct {
		ID        uuid.UUID
		ImageID   uuid.UUID
		Embedding []byte
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	
	var temp TempEmbedding
	result := r.DB.Table("image_embeddings").First(&temp, "image_id = ?", imageID)
	if result.Error != nil {
		return nil, result.Error
	}
	
	// 解析JSON数据
	var embeddingData []float32
	if err := json.Unmarshal(temp.Embedding, &embeddingData); err != nil {
		return nil, err
	}
	
	// 构造返回值
	embedding := &model.ImageEmbedding{
		ID:        temp.ID,
		ImageID:   temp.ImageID,
		Embedding: embeddingData,
		CreatedAt: temp.CreatedAt,
		UpdatedAt: temp.UpdatedAt,
	}
	
	return embedding, nil
}

// SearchSimilarImages 搜索相似图片
func (r *imageRepository) SearchSimilarImages(targetEmbedding []float32, limit int) ([]*model.Image, []float32, error) {
	// 获取所有图片嵌入向量
	type TempEmbedding struct {
		ID        uuid.UUID
		ImageID   uuid.UUID
		Embedding []byte
	}
	
	var tempEmbeddings []*TempEmbedding
	if err := r.DB.Table("image_embeddings").Find(&tempEmbeddings).Error; err != nil {
		return nil, nil, err
	}
	
	// 转换为model.ImageEmbedding格式
	embeddings := make([]*model.ImageEmbedding, len(tempEmbeddings))
	for i, temp := range tempEmbeddings {
		var embeddingData []float32
		if err := json.Unmarshal(temp.Embedding, &embeddingData); err != nil {
			continue // 跳过解析失败的嵌入向量
		}
		embeddings[i] = &model.ImageEmbedding{
			ID:        temp.ID,
			ImageID:   temp.ImageID,
			Embedding: embeddingData,
		}
	}

	// 计算距离并排序
	type imageDistance struct {
		imageID  uuid.UUID
		distance float32
	}

	var distances []imageDistance
	for _, emb := range embeddings {
		dist := calculateEuclideanDistance(targetEmbedding, emb.Embedding)
		distances = append(distances, imageDistance{
			imageID:  emb.ImageID,
			distance: dist,
		})
	}

	// 按距离排序（升序）
	sort.Slice(distances, func(i, j int) bool {
		return distances[i].distance < distances[j].distance
	})

	// 限制结果数量
	if len(distances) > limit {
		distances = distances[:limit]
	}

	// 获取图片信息
	var images []*model.Image
	var resultDistances []float32

	for _, d := range distances {
		var image model.Image
		if err := r.DB.First(&image, "id = ?", d.imageID).Error; err != nil {
			continue
		}
		images = append(images, &image)
		resultDistances = append(resultDistances, d.distance)
	}

	return images, resultDistances, nil
}

// calculateEuclideanDistance 计算欧几里得距离
func calculateEuclideanDistance(v1, v2 []float32) float32 {
	if len(v1) != len(v2) {
		return float32(math.MaxFloat32)
	}

	var sum float32
	for i := range v1 {
		diff := v1[i] - v2[i]
		sum += diff * diff
	}

	return float32(math.Sqrt(float64(sum)))
}