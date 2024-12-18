package machines

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
	"ralts-cms/internal/deps"
	"ralts-cms/internal/machines"
	"time"
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	m, err := h.deps.MachineRepository.GetBySerialNumber(ctx, c.Param("serialnumber"))
	if err != nil {
		log.Errorf("failed to get machine by serial number: %s", err)
		return c.JSON(http.StatusInternalServerError, "")
	}

	if m == nil {
		return c.JSON(http.StatusNotFound, machines.ErrMachineNotFound.Error())
	}

	return c.JSON(http.StatusOK, m)
}
