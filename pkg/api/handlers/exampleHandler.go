package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Roll-Play/togglelabs/pkg/api_utils"
	"github.com/Roll-Play/togglelabs/pkg/config"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ExampleHandler struct {
	db *mongo.Database
}

type ExampleResponse struct {
	ID   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
}

type ExampleListResponse struct {
	Data     []ExampleResponse `json:"data"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int               `json:"total"`
}

func NewExampleRouter(db *mongo.Database) *ExampleHandler {
	return &ExampleHandler{
		db: db,
	}
}

func (eh *ExampleHandler) GetExamples(c echo.Context) error {
	pageQuery := c.QueryParam("page")
	limitQuery := c.QueryParam("page_size")

	page, limit := api_utils.GetPaginationParams(pageQuery, limitQuery)

	collection := eh.db.Collection("example")

	ctx, cancel := context.WithTimeout(context.Background(), config.DBFetchTimeout*time.Second)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSkip(int64((page - 1) * limit))
	findOptions.SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	defer cursor.Close(context.Background())

	records := []ExampleResponse{}

	if cursor.Next(context.Background()) {
		if err = cursor.All(context.Background(), &records); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, ExampleListResponse{
		Data:     records,
		Page:     page,
		PageSize: limit,
		Total:    len(records),
	})
}
