package handlers

import (
	"server/internal/app"
	loadTestController "server/internal/controllers"
	"server/internal/logger"
	. "server/internal/models"

	"github.com/gofiber/fiber/v2"
)

type OptimizedLoadTestHandler struct {
	Handler
	controller loadTestController.OptimizedOnlyController
}

func NewOptimizedLoadTestHandler(app app.App, router fiber.Router) *OptimizedLoadTestHandler {
	log := logger.New("handlers").File("optimizedLoadTest_handler")
	return &OptimizedLoadTestHandler{
		controller: *app.OptimizedOnlyController,
		Handler: Handler{
			log:        log,
			router:     router,
			middleware: app.Middleware,
		},
	}
}

func (h *OptimizedLoadTestHandler) Register() {
	optimized := h.router.Group("/optimized-load-tests")
	optimized.Post("/", h.createOptimizedLoadTest)
	optimized.Get("/performance-summary", h.getOptimizedPerformanceSummary)
	optimized.Get("/:id", h.getOptimizedLoadTest)
	optimized.Get("/", h.getOptimizedLoadTests)
}

func (h *OptimizedLoadTestHandler) createOptimizedLoadTest(c *fiber.Ctx) error {
	log := h.log.Function("createOptimizedLoadTest")

	var request CreateLoadTestRequest
	if err := c.BodyParser(&request); err != nil {
		log.Er("failed to parse optimized load test request", err)
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"message": "failed to parse optimized load test request"})
	}

	// Force method to optimized
	request.Method = "optimized"

	loadTest, err := h.controller.CreateAndRunTest(c.Context(), &request)
	if err != nil {
		log.Er("failed to create and run optimized load test", err)
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"message": "failed to create optimized load test", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "success", "loadTest": loadTest})
}

func (h *OptimizedLoadTestHandler) getOptimizedLoadTest(c *fiber.Ctx) error {
	log := h.log.Function("getOptimizedLoadTest")

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"message": "load test ID is required"})
	}

	loadTest, err := h.controller.GetLoadTestByID(c.Context(), id)
	if err != nil {
		log.Er("failed to get optimized load test", err)
		return c.Status(fiber.StatusNotFound).
			JSON(fiber.Map{"message": "optimized load test not found"})
	}

	return c.JSON(fiber.Map{"message": "success", "loadTest": loadTest})
}

func (h *OptimizedLoadTestHandler) getOptimizedLoadTests(c *fiber.Ctx) error {
	log := h.log.Function("getOptimizedLoadTests")

	loadTests, err := h.controller.GetAllLoadTests(c.Context())
	if err != nil {
		log.Er("failed to get optimized load tests", err)
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"message": "failed to get optimized load tests", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "success", "loadTests": loadTests})
}

func (h *OptimizedLoadTestHandler) getOptimizedPerformanceSummary(c *fiber.Ctx) error {
	log := h.log.Function("getOptimizedPerformanceSummary")

	summary, err := h.controller.GetPerformanceSummary(c.Context())
	if err != nil {
		log.Er("failed to get optimized performance summary", err)
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"message": "failed to get optimized performance summary", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "success", "performanceSummary": summary})
}