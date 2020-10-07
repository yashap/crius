package controller

import (
	"context"
	"github.com/yashap/crius/internal/criusctx"
	"github.com/yashap/crius/internal/db"
	"go.uber.org/zap"
	"net/http"

	"github.com/yashap/crius/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/yashap/crius/internal/domain/service"
	"github.com/yashap/crius/internal/dto"
)

// Service is a controller for /service endpoints
type Service struct {
	ctx               context.Context
	database          db.Database
	serviceRepository service.Repository
	logger            *zap.SugaredLogger
}

// NewService instantiates a Service controller
func NewService(
	ctx context.Context,
	database db.Database,
	serviceRepository service.Repository,
	logger *zap.SugaredLogger,
) Service {
	return Service{ctx, database, serviceRepository, logger}
}

// Create creates a new service.Service
// POST /services { ... service DTO ... }
func (sc *Service) Create(c *gin.Context) {
	ctx, cancel := context.WithCancel(sc.ctx)
	defer cancel()
	serviceDTO, err := dto.MakeServiceFromRequest(c)
	if err != nil {
		errors.SetResponse(err, c)
		return
	}
	// TODO: can/should I move transaction handling to middleware?
	txn, err := sc.database.BeginTransaction(ctx, nil)
	if err != nil {
		errors.SetResponse(err, c)
		return
	}
	ctx = criusctx.WithTransaction(ctx, txn)
	svc := serviceDTO.ToEntity()
	err = sc.serviceRepository.Save(ctx, &svc)
	if err != nil {
		rollbackErr := txn.Rollback()
		if rollbackErr != nil {
			err = errors.DatabaseError(
				"failed to rollback transaction",
				errors.Details{"reasonForRollback": err.Error()},
				&rollbackErr,
			).Logged(sc.logger)
		}
		errors.SetResponse(err, c)
		return
	}
	err = txn.Commit()
	if err != nil {
		errors.SetResponse(errors.DatabaseError("failed to commit transaction", errors.Details{}, &err), c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": svc.ID})
}

// GetByCode gets a service.Service by the service's code
// GET /services/:code { ... service DTO ... }
func (sc *Service) GetByCode(c *gin.Context) {
	ctx, cancel := context.WithCancel(sc.ctx)
	defer cancel()
	code := c.Param("code")
	if code == "" {
		errors.SetResponse(errors.InvalidInput("required param missing", errors.Details{"param": code}, nil), c)
		return
	}
	svc, err := sc.serviceRepository.FindByCode(ctx, code)
	if err != nil {
		errors.SetResponse(err, c)
		return
	}
	if svc == nil {
		errors.SetResponse(
			errors.ServiceNotFound("service not found", errors.Details{"code": code}, nil),
			c,
		)
		return
	}
	c.JSON(http.StatusOK, dto.MakeServiceFromEntity(*svc))
}

// GetAll gets every service.Service
// GET /services
func (sc *Service) GetAll(c *gin.Context) {
	ctx, cancel := context.WithCancel(sc.ctx)
	defer cancel()
	services, err := sc.serviceRepository.FindAll(ctx)
	if err != nil {
		errors.SetResponse(err, c)
		return
	}
	c.JSON(http.StatusNotImplemented, dto.MakeServicesFromEntities(services))
}
