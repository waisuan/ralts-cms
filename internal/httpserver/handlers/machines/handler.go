package machines

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
	"ralts-cms/internal/deps"
)

type Handler struct {
	deps *deps.Dependencies
}

func NewHandler(deps *deps.Dependencies) *Handler {
	return &Handler{
		deps: deps,
	}
}

func (h *Handler) Get(c echo.Context) error {
	m, err := h.deps.MachineRepository.GetBySerialNumber(c.Param("serialnumber"))
	if err != nil {
		log.Errorf("failed to get machine by serial number: %s", err)
		return c.JSON(http.StatusInternalServerError, "")
	}

	if m == nil {
		return c.JSON(http.StatusNotFound, "machine not found")
	}

	return c.JSON(http.StatusOK, m)
}
