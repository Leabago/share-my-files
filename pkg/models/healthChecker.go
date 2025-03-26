package models

import "github.com/go-redis/redis"

type HealthChecker interface {
	Check() error
}

type HealthCheck struct {
	checkers map[string]HealthChecker
}

func NewHealthCheck() *HealthCheck {
	return &HealthCheck{
		checkers: make(map[string]HealthChecker),
	}
}

func (h *HealthCheck) AddChecker(name string, checker HealthChecker) {
	h.checkers[name] = checker
}

func (h *HealthCheck) Check() (map[string]error, bool) {
	errors := make(map[string]error)
	allHealthy := true

	for name, checker := range h.checkers {
		if err := checker.Check(); err != nil {
			errors[name] = err
			allHealthy = false
		}
	}

	return errors, allHealthy
}

// Example DB checker
type RedisChecker struct {
	RedisClient *redis.Client
}

func (d *RedisChecker) Check() error {
	_, err := d.RedisClient.Ping().Result()
	return err
}
