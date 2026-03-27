package health

import (
	"comb-dockerfile/pkg/database"
	"comb-dockerfile/internal/config"
	"comb-dockerfile/internal/pkg/response"
	healthRepository "comb-dockerfile/internal/repository/health"
)

type Health struct{
	Config *config.Config
	DB     database.DB
	HealthRepository healthRepository.Repository
}

func (h *Health) Status() response.Response {
	return response.Success("health")
}
