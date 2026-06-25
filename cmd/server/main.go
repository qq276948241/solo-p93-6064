package main

import (
	"appliance-recycle/internal/config"
	"appliance-recycle/internal/handler"
	"appliance-recycle/internal/pkg/database"
	"appliance-recycle/internal/pkg/middleware"
	"appliance-recycle/internal/pkg/response"
	"appliance-recycle/internal/pkg/upload_helper"
	"appliance-recycle/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("load config failed: " + err.Error())
	}

	if err := database.Init(&cfg.MySQL); err != nil {
		panic("init database failed: " + err.Error())
	}

	if err := service.EnsureDefaultAdmin(); err != nil {
		panic("ensure default admin failed: " + err.Error())
	}

	storage := upload_helper.NewLocalStorage(cfg.Upload)
	appointmentSvc := service.NewAppointmentService(storage, cfg.Upload)

	r := gin.Default()
	r.MaxMultipartMemory = 32 << 20
	r.Static("/uploads", cfg.Upload.LocalPath)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success(gin.H{"status": "ok"}))
	})

	setupRoutes(r, appointmentSvc)

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		panic("start server failed: " + err.Error())
	}
}

func setupRoutes(r *gin.Engine, appointmentSvc *service.AppointmentService) {
	residentHandler := handler.NewResidentHandler(appointmentSvc)
	adminHandler := handler.NewAdminHandler()

	r.GET("/api/appliance-types", residentHandler.GetApplianceTypes)
	r.GET("/api/public/slots", residentHandler.GetSlots)

	api := r.Group("/api")
	{
		resident := api.Group("/resident")
		{
			resident.POST("/register", residentHandler.Register)
			resident.POST("/login", residentHandler.Login)
		}

		admin := api.Group("/admin")
		{
			admin.POST("/login", adminHandler.Login)
		}
	}

	residentAuth := r.Group("/api/resident")
	residentAuth.Use(middleware.ResidentAuth())
	{
		residentAuth.GET("/slots", residentHandler.GetSlots)
		residentAuth.POST("/appointments", residentHandler.CreateAppointment)
		residentAuth.GET("/appointments", residentHandler.MyAppointments)
		residentAuth.POST("/appointments/:id/cancel", residentHandler.CancelAppointment)
	}

	adminAuth := r.Group("/api/admin")
	adminAuth.Use(middleware.AdminAuth())
	{
		adminAuth.GET("/appliance-types", adminHandler.GetApplianceTypes)
		adminAuth.GET("/slots", adminHandler.GetSlots)
		adminAuth.GET("/appointments", adminHandler.ListAppointments)
		adminAuth.GET("/appointments/:id", adminHandler.GetAppointmentDetail)
		adminAuth.PUT("/appointments/:id/status", adminHandler.UpdateAppointmentStatus)
		adminAuth.GET("/statistics", adminHandler.Statistics)
	}
}
