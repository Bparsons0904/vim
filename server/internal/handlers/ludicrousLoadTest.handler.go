package handlers

import (
	"server/internal/app"
	loadTestController "server/internal/controllers"
	"server/internal/logger"
	. "server/internal/models"

	"github.com/gofiber/fiber/v2"
)

type LudicrousLoadTestHandler struct {
	Handler
	controller loadTestController.LudicrousOnlyController
}

func NewLudicrousLoadTestHandler(app app.App, router fiber.Router) *LudicrousLoadTestHandler {
	log := logger.New("handlers").File("ludicrousLoadTest_handler")
	return &LudicrousLoadTestHandler{
		controller: *app.LudicrousOnlyController,
		Handler: Handler{
			log:        log,
			router:     router,
			middleware: app.Middleware,
		},
	}
}

func (h *LudicrousLoadTestHandler) Register() {
	ludicrous := h.router.Group("/ludicrous-load-tests")
	ludicrous.Post("/", h.createLudicrousLoadTest)
	ludicrous.Get("/performance-summary", h.getLudicrousPerformanceSummary)
	ludicrous.Get("/:id", h.getLudicrousLoadTest)
	ludicrous.Get("/", h.getLudicrousLoadTests)
}

func (h *LudicrousLoadTestHandler) createLudicrousLoadTest(c *fiber.Ctx) error {
	log := h.log.Function("createLudicrousLoadTest")

	var request CreateLoadTestRequest
	if err := c.BodyParser(&request); err != nil {
		log.Er("failed to parse ludicrous load test request", err)
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"message": "failed to parse ludicrous load test request"})
	}

	// Force method to ludicrous
	request.Method = "ludicrous"

	loadTest, err := h.controller.CreateAndRunTest(c.Context(), &request)
	if err != nil {
		log.Er("failed to create and run ludicrous load test", err)
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"message": "failed to create ludicrous load test", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "success", "loadTest": loadTest})
}

func (h *LudicrousLoadTestHandler) getLudicrousLoadTest(c *fiber.Ctx) error {
	log := h.log.Function("getLudicrousLoadTest")

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"message": "load test ID is required"})
	}

	loadTest, err := h.controller.GetLoadTestByID(c.Context(), id)
	if err != nil {
		log.Er("failed to get ludicrous load test", err)
		return c.Status(fiber.StatusNotFound).
			JSON(fiber.Map{"message": "ludicrous load test not found"})
	}

	return c.JSON(fiber.Map{"message": "success", "loadTest": loadTest})
}

func (h *LudicrousLoadTestHandler) getLudicrousLoadTests(c *fiber.Ctx) error {
	log := h.log.Function("getLudicrousLoadTests")

	loadTests, err := h.controller.GetAllLoadTests(c.Context())
	if err != nil {
		log.Er("failed to get ludicrous load tests", err)
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"message": "failed to get ludicrous load tests", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "success", "loadTests": loadTests})
}

func (h *LudicrousLoadTestHandler) getLudicrousPerformanceSummary(c *fiber.Ctx) error {
	log := h.log.Function("getLudicrousPerformanceSummary")

	summary, err := h.controller.GetPerformanceSummary(c.Context())
	if err != nil {
		log.Er("failed to get ludicrous performance summary", err)
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"message": "failed to get ludicrous performance summary", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "success", "performanceSummary": summary})
}