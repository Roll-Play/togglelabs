package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roll-Play/togglelabs/pkg/api/common"
	apierrors "github.com/Roll-Play/togglelabs/pkg/api/error"
	"github.com/Roll-Play/togglelabs/pkg/api/handlers"
	"github.com/Roll-Play/togglelabs/pkg/config"
	"github.com/Roll-Play/togglelabs/pkg/models"
	testutils "github.com/Roll-Play/togglelabs/pkg/utils/test_utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type SignUpHandlerTestSuite struct {
	testutils.DefaultTestSuite
	db *mongo.Database
}

func (suite *SignUpHandlerTestSuite) SetupTest() {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://test:test@localhost:27017"))
	if err != nil {
		panic(err)
	}

	suite.db = client.Database(config.TestDBName)
	suite.Server = echo.New()
}

func (suite *SignUpHandlerTestSuite) AfterTest(_, _ string) {
	if err := suite.db.Drop(context.Background()); err != nil {
		panic(err)
	}
}

func (suite *SignUpHandlerTestSuite) TearDownSuite() {
	if err := suite.db.Client().Disconnect(context.Background()); err != nil {
		panic(err)
	}

	suite.Server.Close()
}

func (suite *SignUpHandlerTestSuite) TestSignUpHandlerSuccess() {
	t := suite.T()

	model := models.NewUserModel(suite.db.Collection(models.UserCollectionName))

	requestBody := []byte(`{
		"email": "fizi@gmail.com",
		"password": "123123",
		"first_name": "fizi",
		"last_name": "valores"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	h := handlers.NewSignUpHandler(suite.db)
	c := suite.Server.NewContext(req, rec)
	var jsonRes common.AuthResponse

	assert.NoError(t, h.PostUser(c))

	ur, err := model.FindByEmail(context.Background(), "fizi@gmail.com")
	assert.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &jsonRes))
	assert.Equal(t, jsonRes.Email, ur.Email)
	assert.Equal(t, jsonRes.FirstName, ur.FirstName)
	assert.Equal(t, jsonRes.LastName, ur.LastName)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(ur.Password), []byte("123123")))
}

func (suite *SignUpHandlerTestSuite) TestSignUpHandlerUnsuccessful() {
	t := suite.T()

	model := models.NewUserModel(suite.db.Collection(models.UserCollectionName))

	r, err := models.NewUserRecord(
		"fizi@gmail.com",
		"123123",
		"fizi",
		"valores",
	)

	assert.NoError(t, err)

	_, err = model.InsertOne(context.Background(), r)

	assert.NoError(t, err)

	requestBody, err := json.Marshal(r)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(requestBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	h := handlers.NewSignUpHandler(suite.db)
	c := suite.Server.NewContext(req, rec)
	var jsonRes apierrors.Error

	assert.NoError(t, h.PostUser(c))

	_, err = model.FindByEmail(context.Background(), "fizi@gmail.com")
	assert.NoError(t, err)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &jsonRes))
	assert.Equal(t, jsonRes, apierrors.Error{
		Error:   http.StatusText(http.StatusConflict),
		Message: apierrors.EmailConflictError,
	})
}

func TestSignUpHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SignUpHandlerTestSuite))
}
