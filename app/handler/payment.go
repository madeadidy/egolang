package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/codeuiprogramming/e-commerce/app/consts"
    "github.com/codeuiprogramming/e-commerce/app/models"
    "github.com/codeuiprogramming/e-commerce/database"
)

type NotificationPayload struct {
    TransactionStatus string `json:"transaction_status"`
    FraudStatus       string `json:"fraud_status"`
    OrderID           string `json:"order_id"`
}

func PaymentNotification(c *gin.Context) {
    var payload NotificationPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var order models.Order
    if err := database.DB.Where("id = ?", payload.OrderID).First(&order).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
        return
    }

    if payload.TransactionStatus == consts.PaymentStatusSettlement ||
        (payload.TransactionStatus == consts.PaymentStatusCapture && payload.FraudStatus == consts.FraudStatusAccept) {

        order.PaymentStatus = consts.OrderPaymentStatusPaid
        order.Status = consts.OrderStatusReceived

        if err := database.DB.Save(&order).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order"})
            return
        }
    }

    c.JSON(http.StatusOK, gin.H{"message": "Notification processed"})
}

