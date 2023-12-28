package handlers

import (
	"context"
	"net/http"

	"github.com/Roll-Play/togglelabs/pkg/api/common"
	"github.com/Roll-Play/togglelabs/pkg/config"
	apierror "github.com/Roll-Play/togglelabs/pkg/error"
	"github.com/Roll-Play/togglelabs/pkg/models"
	apiutils "github.com/Roll-Play/togglelabs/pkg/utils/api_utils"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

const UserCollectionName = "user"

type SignUpHandler struct {
	db *mongo.Database
}

func NewSignUpHandler(db *mongo.Database) *SignUpHandler {
	return &SignUpHandler{
		db: db,
	}
}

type SignUpRequest struct {
	Email     string `json:"email" bson:"email"`
	Password  string `json:"password" bson:"password"`
	FirstName string `json:"first_name" bson:"first_name"`
	LastName  string `json:"last_name" bson:"last_name"`
}

func (sh *SignUpHandler) PostUser(c echo.Context) error {
	req := new(SignUpRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	model := models.NewUserModel(sh.db.Collection(models.UserCollectionName))
	_, err := model.FindByEmail(context.Background(), req.Email)
	// err := collection.FindOne(context.Background(), bson.D{{Key: "email", Value: req.Email}}).Decode(&foundRecord)
	if err == nil {
		return apierror.CustomError(c, http.StatusConflict, apierror.EmailConflictError)
	}

	ur, err := models.NewUserRecord(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		return apierror.CustomError(c, http.StatusInternalServerError, apierror.InternalServerError)
	}

	objectID, err := model.InsertOne(context.Background(), ur)
	if err != nil {
		return apierror.CustomError(c, http.StatusInternalServerError, apierror.InternalServerError)
	}

	token, err := apiutils.CreateJWT(objectID, config.JWTExpireTime)
	if err != nil {
		return apierror.CustomError(c, http.StatusInternalServerError, apierror.InternalServerError)
	}

	return c.JSON(http.StatusCreated, common.AuthResponse{
		ID:        objectID,
		Email:     ur.Email,
		FirstName: ur.FirstName,
		LastName:  ur.LastName,
		Token:     token,
	})
}
