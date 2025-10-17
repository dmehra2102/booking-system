package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/dmehra2102/booking-system/internal/common/logger"
	"github.com/dmehra2102/booking-system/internal/user/domain"
	"github.com/dmehra2102/booking-system/pkg/response"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

type UserService interface {
	CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	UpdateUser(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error)
}

type UserHandler struct {
	service UserService
	logger  *logger.Logger
	tracer  trace.Tracer
}

func NewUserHandler(service UserService, logger *logger.Logger, tracer trace.Tracer) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
		tracer:  tracer,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	response.Created(c, user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	loginResp, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err)
		return
	}

	response.Success(c, loginResp)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.service.GetUser(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err)
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteUser(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusNotFound, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	users, total, err := h.service.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err)
		return
	}

	pagination := &response.Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}

	response.Paginated(c, users, pagination)
}
