package handlers

import (
	"server/internal/app"
	loadTestController "server/internal/controllers"
	"server/internal/logger"
	. "server/internal/models"

	"github.com/gofiber/fiber/v2"
)

type LoadTestHandler struct {
	Handler
	controller loadTestController.LoadTestController
}

func NewLoadTestHandler(app app.App, router fiber.Router) *LoadTestHandler {
	log := logger.New("handlers").File("loadTest_handler")
	return &LoadTestHandler{
		controller: *app.LoadTestController,
		Handler: Handler{
			log:        log,
			router:     router,
			middleware: app.Middleware,
		},
	}
}

func (h *LoadTestHandler) Register() {
	loadTests := h.router.Group("/load-tests")
	loadTests.Post("/", h.createLoadTest)
	loadTests.Get("/:id", h.getLoadTest)
	loadTests.Get("/", h.getLoadTests)
}

func (h *LoadTestHandler) createLoadTest(c *fiber.Ctx) error {
	log := h.log.Function("createLoadTest")

	var request CreateLoadTestRequest
	if err := c.BodyParser(&request); err != nil {
		log.Er("failed to parse load test request", err)
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"message": "failed to parse load test request"})
	}

	loadTest, err := h.controller.CreateAndRunTest(c.Context(), &request)
	if err != nil {
		log.Er("failed to create and run load test", err)
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"message": "failed to create load test", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "success", "loadTest": loadTest})
}

func (h *LoadTestHandler) getLoadTest(c *fiber.Ctx) error {
	log := h.log.Function("getLoadTest")

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"message": "load test ID is required"})
	}

	loadTest, err := h.controller.GetLoadTestByID(c.Context(), id)
	if err != nil {
		log.Er("failed to get load test", err)
		return c.Status(fiber.StatusNotFound).
			JSON(fiber.Map{"message": "load test not found"})
	}

	return c.JSON(fiber.Map{"message": "success", "loadTest": loadTest})
}

func (h *LoadTestHandler) getLoadTests(c *fiber.Ctx) error {
	log := h.log.Function("getLoadTests")

	loadTests, err := h.controller.GetAllLoadTests(c.Context())
	if err != nil {
		log.Er("failed to get load tests", err)
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"message": "failed to get load tests", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "success", "loadTests": loadTests})
}