package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	apierrors "github.com/Roll-Play/togglelabs/pkg/api/error"
	"github.com/Roll-Play/togglelabs/pkg/models"
	apiutils "github.com/Roll-Play/togglelabs/pkg/utils/api_utils"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FeatureFlagHandler struct {
	db *mongo.Database
}

func NewFeatureFlagHandler(db *mongo.Database) *FeatureFlagHandler {
	return &FeatureFlagHandler{
		db: db,
	}
}

type PostFeatureFlagRequest struct {
	Name         string          `json:"name" validate:"required"`
	Type         models.FlagType `json:"type" validate:"required,oneof=boolean json string number"`
	DefaultValue string          `json:"default_value" validate:"required"`
	Rules        []models.Rule   `json:"rules" validate:"dive,required"`
}

type PatchFeatureFlagRequest struct {
	DefaultValue string        `json:"default_value"`
	Rules        []models.Rule `json:"rules" validate:"dive,required"`
}

type ListFeatureFlagResponse struct {
	Data     []models.FeatureFlagRecord `json:"data"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"page_size"`
	Total    int                        `json:"total"`
}

func (ffh *FeatureFlagHandler) ListFeatureFlags(c echo.Context) error {
	pageQuery := c.QueryParam("page")
	limitQuery := c.QueryParam("page_size")

	page, limit := apiutils.GetPaginationParams(pageQuery, limitQuery)

	userID, err := apiutils.GetObjectIDFromContext(c)
	if err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(
			c,
			http.StatusUnauthorized,
			apierrors.UnauthorizedError,
		)
	}

	organizationID, err := primitive.ObjectIDFromHex(c.Param("organizationID"))
	if err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(
			c,
			http.StatusBadRequest,
			apierrors.BadRequestError,
		)
	}
	organizationModel := models.NewOrganizationModel(ffh.db)
	organization, err := organizationModel.FindByID(context.Background(), organizationID)
	if err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(
			c,
			http.StatusInternalServerError,
			apierrors.InternalServerError,
		)
	}

	permission := apiutils.UserHasPermission(userID, organization, models.ReadOnly)
	if !permission {
		return apierrors.CustomError(
			c,
			http.StatusForbidden,
			apierrors.ForbiddenError,
		)
	}

	model := models.NewFeatureFlagModel(ffh.db)

	featureFlags, err := model.FindMany(context.Background(), organizationID, page, limit)
	if err != nil {
		return apierrors.CustomError(
			c,
			http.StatusInternalServerError,
			apierrors.InternalServerError,
		)
	}

	bytes, _ := json.Marshal(featureFlags)

	log.Println(apiutils.HandlerLogMessage(string(bytes), primitive.NewObjectID(), c))
	return c.JSON(http.StatusOK, ListFeatureFlagResponse{
		Data:     featureFlags,
		Page:     page,
		PageSize: limit,
		Total:    len(featureFlags),
	})
}

func (ffh *FeatureFlagHandler) PostFeatureFlag(c echo.Context) error {
	userID, err := apiutils.GetObjectIDFromContext(c)
	if err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(
			c,
			http.StatusUnauthorized,
			apierrors.UnauthorizedError,
		)
	}

	organizationID, err := primitive.ObjectIDFromHex(c.Param("organizationID"))
	if err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(
			c,
			http.StatusBadRequest,
			apierrors.BadRequestError,
		)
	}

	organizationModel := models.NewOrganizationModel(ffh.db)
	organizationRecord, err := organizationModel.FindByID(context.Background(), organizationID)
	if err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(c,
			http.StatusInternalServerError,
			apierrors.InternalServerError,
		)
	}

	permission := apiutils.UserHasPermission(userID, organizationRecord, models.Collaborator)
	if !permission {
		log.Println(apiutils.HandlerLogMessage("feature-flag", userID, c))
		return apierrors.CustomError(
			c,
			http.StatusUnauthorized,
			apierrors.UnauthorizedError,
		)
	}

	req := new(PostFeatureFlagRequest)
	if err := c.Bind(req); err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(
			c,
			http.StatusBadRequest,
			apierrors.BadRequestError,
		)
	}

	validate := validator.New()

	if err := validate.Struct(req); err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(c,
			http.StatusBadRequest,
			apierrors.BadRequestError,
		)
	}

	featureFlagModel := models.NewFeatureFlagModel(ffh.db)
	featureFlagRecord := models.NewFeatureFlagRecord(
		req.Name,
		req.DefaultValue,
		req.Type,
		req.Rules,
		organizationID,
		userID,
	)

	insertedID, err := featureFlagModel.InsertOne(context.Background(), featureFlagRecord)
	if err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(c,
			http.StatusInternalServerError,
			apierrors.InternalServerError,
		)
	}
	featureFlagRecord.ID = insertedID

	return c.JSON(http.StatusCreated, featureFlagRecord)
}

func (ffh *FeatureFlagHandler) PatchFeatureFlag(c echo.Context) error {
	userID, err := apiutils.GetObjectIDFromContext(c)
	if err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(
			c,
			http.StatusUnauthorized,
			apierrors.UnauthorizedError,
		)
	}

	organizationID, err := primitive.ObjectIDFromHex(c.Param("organizationID"))
	if err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(
			c,
			http.StatusBadRequest,
			apierrors.BadRequestError,
		)
	}

	organizationModel := models.NewOrganizationModel(ffh.db)
	organizationRecord, err := organizationModel.FindByID(context.Background(), organizationID)
	if err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(c,
			http.StatusInternalServerError,
			apierrors.InternalServerError,
		)
	}

	permission := apiutils.UserHasPermission(userID, organizationRecord, models.Collaborator)
	if !permission {
		log.Println(apiutils.HandlerLogMessage("feature-flag", userID, c))
		return apierrors.CustomError(
			c,
			http.StatusUnauthorized,
			apierrors.UnauthorizedError,
		)
	}

	featureFlagID, err := primitive.ObjectIDFromHex(c.Param("featureFlagID"))
	if err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(
			c,
			http.StatusBadRequest,
			apierrors.BadRequestError,
		)
	}

	req := new(PatchFeatureFlagRequest)
	if err := c.Bind(req); err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(
			c,
			http.StatusBadRequest,
			apierrors.BadRequestError,
		)
	}

	model := models.NewFeatureFlagModel(ffh.db)

	revision := model.NewRevisionRecord(
		req.DefaultValue,
		req.Rules,
		userID,
	)
	_, err = model.UpdateOne(
		context.Background(),
		featureFlagID,
		bson.M{
			"revisions": revision,
		},
	)
	if err != nil {
		log.Println(apiutils.HandlerErrorLogMessage(err, c))
		return apierrors.CustomError(c,
			http.StatusInternalServerError,
			apierrors.InternalServerError,
		)
	}

	return c.JSON(http.StatusOK, revision)
}