package controllers

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/codeuiprogramming/e-commerce/app/consts"
	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/codeuiprogramming/e-commerce/database"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (server *Server) PaymentNotification(w http.ResponseWriter, r *http.Request) {
	var payload models.MidtransNotification

	// Decode JSON dari Midtrans
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fmt.Printf("[MIDTRANS] Notifikasi diterima: OrderID=%s, Status=%s, PaymentType=%s\n",
		payload.OrderID, payload.TransactionStatus, payload.PaymentType)

	// Validasi Signature Key
	serverKey := os.Getenv("API_MIDTRANS_SERVER_KEY")
	if err := validateSignatureKey(&payload, serverKey); err != nil {
		http.Error(w, "Invalid signature key", http.StatusForbidden)
		return
	}

	// Ambil order dari database
	order := models.Order{}
	found, err := order.FindByID(server.DB, payload.OrderID)
	if err != nil {
		http.Error(w, "Order tidak ditemukan", http.StatusNotFound)
		return
	}

	// Hindari double payment
	if found.IsPaid() {
		http.Error(w, "Order sudah dibayar sebelumnya", http.StatusForbidden)
		return
	}

	// Simpan log pembayaran
	payment := models.Payment{}
	amount, _ := decimal.NewFromString(payload.GrossAmount)
	jsonPayload, _ := json.Marshal(payload)

	if _, err := payment.CreatePayment(server.DB, &models.Payment{
		OrderID:           found.ID,
		Amount:            amount,
		TransactionID:     payload.TransactionID,
		TransactionStatus: payload.TransactionStatus,
		Payload:           datatypes.JSON(jsonPayload),
		PaymentType:       payload.PaymentType,
	}); err != nil {
		http.Error(w, "Gagal menyimpan pembayaran: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update status order sesuai status Midtrans
	if isPaymentSuccess(&payload) {
		if err := found.MarkAsPaid(server.DB); err != nil {
			http.Error(w, "Gagal update status order: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf(" Order %s berhasil ditandai sebagai PAID\n", payload.OrderID)
	} else {
		if err := updateOrderStatus(server.DB, payload.OrderID, payload.TransactionStatus); err != nil {
			http.Error(w, "Gagal update status order: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Response ke Midtrans
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Payment notification processed",
		"order_id": payload.OrderID,
		"status":   payload.TransactionStatus,
	})
}

func PaymentNotificationHandler(w http.ResponseWriter, r *http.Request) {
	var notif map[string]interface{}
	json.NewDecoder(r.Body).Decode(&notif)

	orderID := notif["order_id"].(string)
	statusCode := notif["status_code"].(string)
	grossAmount := notif["gross_amount"].(string)
	signatureKey := notif["signature_key"].(string)
	transactionStatus := notif["transaction_status"].(string)
	fraudStatus := notif["fraud_status"].(string)

	// Buat hash untuk verifikasi
	serverKey := "YOUR_SERVER_KEY"
	signature := sha512.Sum512([]byte(orderID + statusCode + grossAmount + serverKey))
	expectedSignature := hex.EncodeToString(signature[:])

	if signatureKey != expectedSignature {
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	// Update order di DB
	var order models.Order
	if err := database.DB.Where("code = ?", orderID).First(&order).Error; err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	if transactionStatus == consts.PaymentStatusCapture && fraudStatus == consts.FraudStatusAccept ||
		transactionStatus == consts.PaymentStatusSettlement {
		order.PaymentStatus = consts.OrderPaymentStatusPaid
		order.Status = consts.OrderStatusReceived
		database.DB.Save(&order)
	}

	w.WriteHeader(http.StatusOK)
}

func isPaymentSuccess(payload *models.MidtransNotification) bool {
	paymentStatus := false
	if payload.PaymentType == string(snap.PaymentTypeCreditCard) {
		paymentStatus = (payload.TransactionStatus == consts.PaymentStatusCapture) && (payload.FraudStatus == consts.FraudStatusAccept)
	} else {
		paymentStatus = (payload.TransactionStatus == consts.PaymentStatusSettlement) && (payload.FraudStatus == consts.FraudStatusAccept)
	}

	return paymentStatus
}

func validateSignatureKey(payload *models.MidtransNotification, serverKey string) error {
	signaturePayload := payload.OrderID + payload.StatusCode + payload.GrossAmount + serverKey
	hash := sha512.Sum512([]byte(signaturePayload))
	expected := fmt.Sprintf("%x", hash)
	if expected != payload.SignatureKey {
		return errors.New("invalid signature key")
	}
	return nil
}

// Fungsi bantu update order
func updateOrderStatus(db *gorm.DB, orderID, status string) error {
	// Map status Midtrans ke status internal aplikasi
	var mappedStatus string
	switch status {
	case "capture":
		mappedStatus = "paid"
	case "settlement":
		mappedStatus = "paid"
	case "pending":
		mappedStatus = "pending"
	case "deny":
		mappedStatus = "failed"
	case "expire":
		mappedStatus = "expired"
	case "cancel":
		mappedStatus = "canceled"
	default:
		mappedStatus = "unknown"
	}

	// Update ke tabel orders
	// result := db.Exec(`UPDATE orders SET status = ? WHERE id = ?`, mappedStatus, orderID)
	result := db.Exec(`UPDATE orders SET payment_status = ? WHERE id = ?`, mappedStatus, orderID)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("order %s not found", orderID)
	}

	fmt.Printf("Order %s diupdate ke status %s\n", orderID, mappedStatus)
	return nil
}
