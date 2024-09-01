package machines

import (
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/testutils/factory"
	"testing"
)

type MachineHandlerTestSuite struct {
	suite.Suite
	repo    *machines.MockRepository
	handler *Handler
}

func (suite *MachineHandlerTestSuite) SetupTest() {
	d := deps.Initialise()

	mockCtrl := gomock.NewController(suite.T())
	defer mockCtrl.Finish()
	repo := machines.NewMockRepository(mockCtrl)
	suite.repo = repo
	d.MachineRepository = repo

	suite.handler = NewHandler(d)
}

func (suite *MachineHandlerTestSuite) TearDownTest() {}

func (suite *MachineHandlerTestSuite) TestGet() {
	var (
		t = suite.T()
		m = factory.BuildMachine()
	)

	suite.repo.EXPECT().GetBySerialNumber(gomock.Any()).Return(m, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/machines/:serialnumber")
	c.SetParamNames("serialnumber")
	c.SetParamValues("123")

	err := suite.handler.Get(c)
	require.Nil(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var machine machines.Machine
	err = json.Unmarshal(rec.Body.Bytes(), &machine)
	require.Nil(t, err)
	assert.Equal(t, m, &machine)
}

func (suite *MachineHandlerTestSuite) TestGetInternalServerError() {
	var (
		t = suite.T()
	)

	suite.repo.EXPECT().GetBySerialNumber(gomock.Any()).Return(nil, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/machines/:serialnumber")
	c.SetParamNames("serialnumber")
	c.SetParamValues("123")

	err := suite.handler.Get(c)
	require.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func (suite *MachineHandlerTestSuite) TestGetNotFound() {
	var (
		t = suite.T()
	)

	suite.repo.EXPECT().GetBySerialNumber(gomock.Any()).Return(nil, errors.New("something went wrong"))

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/machines/:serialnumber")
	c.SetParamNames("serialnumber")
	c.SetParamValues("123")

	err := suite.handler.Get(c)
	require.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestMachineHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MachineHandlerTestSuite))
}
