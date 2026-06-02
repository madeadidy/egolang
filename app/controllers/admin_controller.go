package controllers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/codeuiprogramming/e-commerce/app/consts"
	"github.com/codeuiprogramming/e-commerce/app/helpers"
	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"github.com/unrolled/render"
	"github.com/xuri/excelize/v2"
)

// persistError appends an error message to a local logfile for debugging.
func persistError(err error) {
    if err == nil {
        return
    }
    // use a workspace-relative logfile so the running process can write
    fpath := "create_errors.log"
    f, fErr := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if fErr != nil {
        fmt.Println("persistError: failed to open log file:", fErr)
        return
    }
    defer f.Close()
    t := time.Now().Format(time.RFC3339)
    // write both to logfile and stdout for immediate visibility
    _, _ = f.WriteString(fmt.Sprintf("%s - %v\n", t, err))
    fmt.Println("persistError: logged error:", err)
}

func newAdminRender() *render.Render {
    return render.New(render.Options{
        Layout:     "layout",
        Extensions: []string{".html", ".tmpl"},
        Funcs: []template.FuncMap{
            {"FormatPrice": helpers.FormatPrice},
        },
    })
}

// AdminIndex shows a simple dashboard summary
func (server *Server) AdminIndex(w http.ResponseWriter, r *http.Request) {
    render := newAdminRender()

    var userCount int64
    var productCount int64
    var orderCount int64
    server.DB.Model(&models.User{}).Count(&userCount)
    server.DB.Model(&models.Product{}).Where("is_temporary = ?", false).Count(&productCount)
    server.DB.Model(&models.Order{}).Count(&orderCount)

    data := server.DefaultRenderData(w, r, map[string]interface{}{
        "userCount":    userCount,
        "productCount": productCount,
        "orderCount":   orderCount,
    })

    _ = render.HTML(w, http.StatusOK, "admin/index", data)
}

// AdminProducts lists products
func (server *Server) AdminProducts(w http.ResponseWriter, r *http.Request) {
    render := newAdminRender()

    var products []models.Product
    server.DB.Preload("ProductImages").Where("is_temporary = ?", false).Order("created_at desc").Find(&products)

    data := server.DefaultRenderData(w, r, map[string]interface{}{"products": products})
    _ = render.HTML(w, http.StatusOK, "admin/products", data)
}

// AdminProductNew shows form
func (server *Server) AdminProductNew(w http.ResponseWriter, r *http.Request) {
    render := newAdminRender()
    data := server.DefaultRenderData(w, r, map[string]interface{}{})
    _ = render.HTML(w, http.StatusOK, "admin/product_form", data)
}

// AdminProductCreate handles creation (basic fields)
func (server *Server) AdminProductCreate(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "invalid form", http.StatusBadRequest)
        return
    }
    name := r.FormValue("name")
    priceStr := r.FormValue("price")
    stockStr := r.FormValue("stock")

    if name == "" {
        SetFlash(w, r, "error", "Nama produk diperlukan")
        http.Redirect(w, r, "/admin/products/new", http.StatusSeeOther)
        return
    }

    price, _ := strconv.ParseFloat(priceStr, 64)
    stock, _ := strconv.Atoi(stockStr)

    p := models.Product{
        ID:               "",
        Name:             name,
        Price:            decimal.NewFromFloat(price),
        Stock:            stock,
        ShortDescription: r.FormValue("short_description"),
        Description:      r.FormValue("description"),
    }

    // If no user is associated, assign the first available user as owner
    if p.UserID == "" {
        var firstUser models.User
        if err := server.DB.First(&firstUser).Error; err == nil {
            p.UserID = firstUser.ID
        } else {
            // log but continue; DB will report FK error if no users exist
            fmt.Println("AdminProductCreate: unable to find default user:", err)
        }
    }

    // create
    if err := server.DB.Create(&p).Error; err != nil {
        fmt.Println("AdminProductCreate create error:", err)
        // persist error for offline inspection
        persistError(err)
        SetFlash(w, r, "error", "Gagal menyimpan produk")
        http.Redirect(w, r, "/admin/products/new", http.StatusSeeOther)
        return
    }

    SetFlash(w, r, "success", "Produk berhasil dibuat")
    http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
}

// AdminProductEdit shows edit form
func (server *Server) AdminProductEdit(w http.ResponseWriter, r *http.Request) {
    render := newAdminRender()
    vars := mux.Vars(r)
    id := vars["id"]
    var p models.Product
    if err := server.DB.Where("id = ?", id).First(&p).Error; err != nil {
        SetFlash(w, r, "error", "Produk tidak ditemukan")
        http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
        return
    }
    data := server.DefaultRenderData(w, r, map[string]interface{}{"product": p})
    _ = render.HTML(w, http.StatusOK, "admin/product_form", data)
}

// AdminProductUpdate handles update
func (server *Server) AdminProductUpdate(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    if err := r.ParseForm(); err != nil {
        SetFlash(w, r, "error", "invalid form")
        http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
        return
    }
    name := r.FormValue("name")
    priceStr := r.FormValue("price")
    stockStr := r.FormValue("stock")
    price, _ := strconv.ParseFloat(priceStr, 64)
    stock, _ := strconv.Atoi(stockStr)

    updates := map[string]interface{}{"name": name, "stock": stock}
    updates["price"] = decimal.NewFromFloat(price)

    if err := server.DB.Model(&models.Product{}).Where("id = ?", id).Updates(updates).Error; err != nil {
        SetFlash(w, r, "error", "Gagal memperbarui produk")
    } else {
        SetFlash(w, r, "success", "Produk diperbarui")
    }
    http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
}

// AdminProductDelete deletes a product
func (server *Server) AdminProductDelete(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    if id == "" {
        SetFlash(w, r, "error", "invalid id")
        http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
        return
    }
    if err := server.DB.Where("id = ?", id).Delete(&models.Product{}).Error; err != nil {
        SetFlash(w, r, "error", "Gagal menghapus produk")
    } else {
        SetFlash(w, r, "success", "Produk dihapus")
    }
    http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
}

// AdminOrders lists orders
func (server *Server) AdminOrders(w http.ResponseWriter, r *http.Request) {
    render := newAdminRender()
    var orders []models.Order
    server.DB.Preload("OrderItems").Preload("User").Order("created_at desc").Find(&orders)
    data := server.DefaultRenderData(w, r, map[string]interface{}{"orders": orders})
    _ = render.HTML(w, http.StatusOK, "admin/orders", data)
}

// AdminOrderDetail shows order details
func (server *Server) AdminOrderDetail(w http.ResponseWriter, r *http.Request) {
    render := newAdminRender()
    vars := mux.Vars(r)
    id := vars["id"]
    var ord models.Order
    if _, err := ord.FindByID(server.DB, id); err != nil {
        SetFlash(w, r, "error", "Order tidak ditemukan")
        http.Redirect(w, r, "/admin/orders", http.StatusSeeOther)
        return
    }
    data := server.DefaultRenderData(w, r, map[string]interface{}{"order": ord})
    _ = render.HTML(w, http.StatusOK, "admin/order_detail", data)
}

// AdminCustomers lists customers
func (server *Server) AdminCustomers(w http.ResponseWriter, r *http.Request) {
    render := newAdminRender()
    var users []models.User
    server.DB.Order("created_at desc").Find(&users)
    data := server.DefaultRenderData(w, r, map[string]interface{}{"customers": users})
    _ = render.HTML(w, http.StatusOK, "admin/customers", data)
}

// AdminCustomerDetail shows a customer's orders
func (server *Server) AdminCustomerDetail(w http.ResponseWriter, r *http.Request) {
    render := newAdminRender()
    vars := mux.Vars(r)
    id := vars["id"]
    var u models.User
    if err := server.DB.Where("id = ?", id).First(&u).Error; err != nil {
        SetFlash(w, r, "error", "Pelanggan tidak ditemukan")
        http.Redirect(w, r, "/admin/customers", http.StatusSeeOther)
        return
    }
    var orders []models.Order
    server.DB.Where("user_id = ?", id).Preload("OrderItems").Find(&orders)
    data := server.DefaultRenderData(w, r, map[string]interface{}{"customer": u, "orders": orders})
    _ = render.HTML(w, http.StatusOK, "admin/customer_detail", data)
}

// AdminReportMonthly shows printable monthly transactions
func (server *Server) AdminReportMonthly(w http.ResponseWriter, r *http.Request) {
    render := newAdminRender()
    // month and year params optional
    monthStr := r.URL.Query().Get("month")
    yearStr := r.URL.Query().Get("year")
    now := time.Now()
    month := int(now.Month())
    year := now.Year()
    if monthStr != "" {
        if m, err := strconv.Atoi(monthStr); err == nil {
            month = m
        }
    }
    if yearStr != "" {
        if y, err := strconv.Atoi(yearStr); err == nil {
            year = y
        }
    }
    start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
    end := start.AddDate(0, 1, 0)

    var orders []models.Order
    server.DB.Where("created_at >= ? AND created_at < ?", start, end).Preload("OrderItems").Preload("User").Find(&orders)
    data := server.DefaultRenderData(w, r, map[string]interface{}{"orders": orders, "month": month, "year": year})
    _ = render.HTML(w, http.StatusOK, "admin/report_monthly", data)
}

// APIAdminProducts handles JSON list and create for products
func (server *Server) APIAdminProducts(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    ren := newAdminRender()
    switch r.Method {
    case "GET":
        var products []models.Product
        server.DB.Preload("ProductImages").Where("is_temporary = ?", false).Order("created_at desc").Find(&products)
        _ = ren.JSON(w, http.StatusOK, products)
        return
    case "POST":
        // create product from JSON body
        var payload struct {
            Name             string  `json:"name"`
            Price            float64 `json:"price"`
            Stock            int     `json:"stock"`
            ShortDescription string  `json:"short_description"`
            Description      string  `json:"description"`
        }
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, "invalid json", http.StatusBadRequest)
            return
        }
        p := models.Product{
            Name:             payload.Name,
            Price:            decimal.NewFromFloat(payload.Price),
            Stock:            payload.Stock,
            ShortDescription: payload.ShortDescription,
            Description:      payload.Description,
        }
        // attach an existing user as owner to satisfy foreign key constraints
        // use a small struct to read the id column reliably
        type uidRow struct{
            ID string `gorm:"column:id"`
        }
        var ur uidRow
        server.DB.Raw("SELECT id FROM users LIMIT 1").Scan(&ur)
        if ur.ID != "" {
            p.UserID = ur.ID
            fmt.Println("APIAdminProducts: attaching user id (raw struct):", ur.ID)
        } else {
            fmt.Println("APIAdminProducts: no user id found via raw struct query")
        }
        if err := server.DB.Create(&p).Error; err != nil {
            fmt.Println("APIAdminProducts create error:", err)
            // persist error to logfile for debugging
            persistError(err)
            // return full error in JSON for debugging
            _ = ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
        _ = ren.JSON(w, http.StatusCreated, p)
        return
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
}

// APIAdminProduct handles GET/PUT/DELETE for a single product
func (server *Server) APIAdminProduct(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    ren := newAdminRender()
    vars := mux.Vars(r)
    id := vars["id"]
    var p models.Product
    if err := server.DB.Preload("ProductImages").Where("id = ?", id).First(&p).Error; err != nil {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }
    switch r.Method {
    case "GET":
        _ = ren.JSON(w, http.StatusOK, p)
        return
    case "PUT":
        var payload map[string]interface{}
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, "invalid json", http.StatusBadRequest)
            return
        }
        // handle price specially if present
        if priceVal, ok := payload["price"]; ok {
            if f, ok2 := priceVal.(float64); ok2 {
                payload["price"] = decimal.NewFromFloat(f)
            }
        }
        if err := server.DB.Model(&models.Product{}).Where("id = ?", id).Updates(payload).Error; err != nil {
            http.Error(w, "failed to update", http.StatusInternalServerError)
            return
        }
        server.DB.Where("id = ?", id).First(&p)
        _ = ren.JSON(w, http.StatusOK, p)
        return
    case "DELETE":
        if err := server.DB.Where("id = ?", id).Delete(&models.Product{}).Error; err != nil {
            http.Error(w, "failed to delete", http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusNoContent)
        return
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
}

// AdminReportCSV returns CSV for the selected month
func (server *Server) AdminReportCSV(w http.ResponseWriter, r *http.Request) {
    // support CSV export by month/year OR by start/end ISO dates (YYYY-MM-DD)
    qs := r.URL.Query()
    startStr := qs.Get("start")
    endStr := qs.Get("end")
    monthStr := qs.Get("month")
    yearStr := qs.Get("year")

    var start time.Time
    var end time.Time
    var err error

    if startStr != "" && endStr != "" {
        // parse start/end as YYYY-MM-DD
        start, err = time.Parse("2006-01-02", startStr)
        if err != nil {
            http.Error(w, "invalid start date", http.StatusBadRequest)
            return
        }
        end, err = time.Parse("2006-01-02", endStr)
        if err != nil {
            http.Error(w, "invalid end date", http.StatusBadRequest)
            return
        }
        // make end exclusive by adding one day
        end = end.Add(24 * time.Hour)
    } else {
        // fallback to month/year behavior
        now := time.Now()
        month := int(now.Month())
        year := now.Year()
        if monthStr != "" {
            if m, e := strconv.Atoi(monthStr); e == nil { month = m }
        }
        if yearStr != "" {
            if y, e := strconv.Atoi(yearStr); e == nil { year = y }
        }
        start = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
        end = start.AddDate(0, 1, 0)
    }

    var orders []models.Order
    server.DB.Where("order_date >= ? AND order_date < ?", start, end).Preload("OrderItems").Preload("User").Find(&orders)

    // If client requested XLSX, generate native Excel file
    // reuse previously read query values (qs)
    if strings.ToLower(qs.Get("export")) == "xlsx" || strings.ToLower(qs.Get("format")) == "xlsx" {
        f := excelize.NewFile()
        sheetName := "Transactions"
        // create sheet and set headers
        idx, _ := f.NewSheet(sheetName)
        f.SetActiveSheet(idx)
        headers := []string{"OrderID", "Code", "Customer", "OrderDate", "ItemsCount", "GrandTotal"}
        for i, h := range headers {
            cell, _ := excelize.CoordinatesToCellName(i+1, 1)
            f.SetCellValue(sheetName, cell, h)
        }
        // fill rows
        for rIdx, o := range orders {
            row := rIdx + 2
            f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), o.ID)
            f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), o.Code)
            f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), o.User.FirstName+" "+o.User.LastName)
            f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), o.OrderDate.Format(time.RFC3339))
            f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), len(o.OrderItems))
            f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), o.GrandTotal.String())
        }

        buf, err := f.WriteToBuffer()
        if err != nil {
            http.Error(w, "failed to build xlsx", http.StatusInternalServerError)
            return
        }
        filename := fmt.Sprintf("transactions_%s_to_%s.xlsx", start.Format("2006-01-02"), end.Add(-time.Nanosecond).Format("2006-01-02"))
        w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
        w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
        w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
        _, _ = w.Write(buf.Bytes())
        return
    }

    // default: CSV export (Excel-friendly)
    w.Header().Set("Content-Type", "text/csv; charset=utf-8")
    filename := fmt.Sprintf("transactions_%s_to_%s.csv", start.Format("2006-01-02"), end.Add(-time.Nanosecond).Format("2006-01-02"))
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
    // write UTF-8 BOM so Excel recognizes UTF-8 encoding
    _, _ = w.Write([]byte{0xEF, 0xBB, 0xBF})
    writer := csv.NewWriter(w)
    // allow client to request semicolon delimiter for locales where Excel expects ';'
    delim := qs.Get("delimiter")
    if strings.ToLower(qs.Get("export")) == "csv_excel" || strings.ToLower(delim) == ";" || strings.ToLower(delim) == "semicolon" {
        writer.Comma = ';'
    }
    defer writer.Flush()
    // write header and rows
    _ = writer.Write([]string{"OrderID", "Code", "Customer", "OrderDate", "ItemsCount", "GrandTotal"})
    for _, o := range orders {
        _ = writer.Write([]string{o.ID, o.Code, o.User.FirstName + " " + o.User.LastName, o.OrderDate.Format(time.RFC3339), strconv.Itoa(len(o.OrderItems)), o.GrandTotal.String()})
    }
}

// APIAdminReportTransactions returns JSON list of orders filtered by date range + pagination
func (server *Server) APIAdminReportTransactions(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    ren := newAdminRender()

    qs := r.URL.Query()
    startStr := qs.Get("start")
    endStr := qs.Get("end")
    perPage := 20
    page := 1
    if pp := qs.Get("per_page"); pp != "" {
        if v, e := strconv.Atoi(pp); e == nil && v > 0 { perPage = v }
    }
    if pg := qs.Get("page"); pg != "" {
        if v, e := strconv.Atoi(pg); e == nil && v > 0 { page = v }
    }

    // allow CSV export via export=csv
    exportFmt := qs.Get("export")

    // Build base filter (for count and sum) and fetch query with preloads
    baseQ := server.DB.Model(&models.Order{})
    fetchQ := server.DB.Preload("OrderItems").Preload("User").Order("order_date desc")
    var startT time.Time
    var endT time.Time
    if startStr != "" {
        if t, e := time.Parse("2006-01-02", startStr); e == nil {
            startT = t
            baseQ = baseQ.Where("order_date >= ?", startT)
            fetchQ = fetchQ.Where("order_date >= ?", startT)
        }
    }
    if endStr != "" {
        if t, e := time.Parse("2006-01-02", endStr); e == nil {
            // make end exclusive
            t = t.Add(24 * time.Hour)
            endT = t
            baseQ = baseQ.Where("order_date < ?", endT)
            fetchQ = fetchQ.Where("order_date < ?", endT)
        }
    }

    // apply customer filter (partial match) and payment/status filters
    customerQ := qs.Get("customer")
    statusQ := qs.Get("status")
    paymentQ := qs.Get("payment_status")
    if customerQ != "" {
        like := "%%%s%%"
        like = fmt.Sprintf(like, customerQ)
        baseQ = baseQ.Joins("LEFT JOIN users ON users.id = orders.user_id").Where("users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ?", like, like, like)
        fetchQ = fetchQ.Joins("LEFT JOIN users ON users.id = orders.user_id").Where("users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ?", like, like, like)
    }
    if paymentQ != "" {
        switch strings.ToLower(paymentQ) {
        case "paid":
            baseQ = baseQ.Where("payment_status = ?", consts.OrderPaymentStatusPaid)
            fetchQ = fetchQ.Where("payment_status = ?", consts.OrderPaymentStatusPaid)
        case "unpaid":
            baseQ = baseQ.Where("payment_status != ?", consts.OrderPaymentStatusPaid)
            fetchQ = fetchQ.Where("payment_status != ?", consts.OrderPaymentStatusPaid)
        }
    }
    if statusQ != "" {
        // accept label names or numeric
        switch strings.ToLower(statusQ) {
        case "pending":
            baseQ = baseQ.Where("status = ?", consts.OrderStatusPending)
            fetchQ = fetchQ.Where("status = ?", consts.OrderStatusPending)
        case "received":
            baseQ = baseQ.Where("status = ?", consts.OrderStatusReceived)
            fetchQ = fetchQ.Where("status = ?", consts.OrderStatusReceived)
        case "delivered", "shipped":
            baseQ = baseQ.Where("status = ?", consts.OrderStatusDelivered)
            fetchQ = fetchQ.Where("status = ?", consts.OrderStatusDelivered)
        case "cancelled", "canceled":
            baseQ = baseQ.Where("status = ?", consts.OrderStatusCancelled)
            fetchQ = fetchQ.Where("status = ?", consts.OrderStatusCancelled)
        default:
            // try numeric
            if n, e := strconv.Atoi(statusQ); e == nil {
                baseQ = baseQ.Where("status = ?", n)
                fetchQ = fetchQ.Where("status = ?", n)
            }
        }
    }

    // total count for pagination
    var totalCount int64
    _ = baseQ.Count(&totalCount).Error

    // total revenue (sum of grand_total)
    var totalRevenueStr string
    _ = baseQ.Select("COALESCE(SUM(grand_total)::text, '0')").Row().Scan(&totalRevenueStr)

    // revenue by status (use the same date/customer/payment filters but partitioned by status)
    revenueByStatus := map[string]string{}
    statuses := map[string]int{"pending": consts.OrderStatusPending, "received": consts.OrderStatusReceived, "delivered": consts.OrderStatusDelivered, "cancelled": consts.OrderStatusCancelled}
    for k, v := range statuses {
        var s string
        // use a copy of baseQ without status filter to compute per-status sums
        q := server.DB.Model(&models.Order{})
        if !startT.IsZero() {
            q = q.Where("order_date >= ?", startT)
        }
        if !endT.IsZero() {
            q = q.Where("order_date < ?", endT)
        }
        if customerQ := qs.Get("customer"); customerQ != "" {
            like := fmt.Sprintf("%%%s%%", customerQ)
            q = q.Joins("LEFT JOIN users ON users.id = orders.user_id").Where("users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ?", like, like, like)
        }
        if paymentQ := qs.Get("payment_status"); paymentQ != "" {
            switch strings.ToLower(paymentQ) {
            case "paid":
                q = q.Where("payment_status = ?", consts.OrderPaymentStatusPaid)
            case "unpaid":
                q = q.Where("payment_status != ?", consts.OrderPaymentStatusPaid)
            }
        }
        q = q.Where("status = ?", v)
        _ = q.Select("COALESCE(SUM(grand_total)::text, '0')").Row().Scan(&s)
        revenueByStatus[k] = s
    }

    // handle CSV export directly from API
    if exportFmt == "csv" {
        var orders []models.Order
        fetchQ.Find(&orders)
        w.Header().Set("Content-Type", "text/csv")
        fname := "transactions.csv"
        if !startT.IsZero() {
            fname = fmt.Sprintf("transactions_%s_to_%s.csv", startT.Format("2006-01-02"), endT.Add(-time.Nanosecond).Format("2006-01-02"))
        }
        w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fname))
        writer := csv.NewWriter(w)
        defer writer.Flush()
        writer.Write([]string{"OrderID", "Code", "Customer", "OrderDate", "ItemsCount", "GrandTotal"})
        for _, o := range orders {
            writer.Write([]string{o.ID, o.Code, o.User.FirstName + " " + o.User.LastName, o.OrderDate.Format(time.RFC3339), strconv.Itoa(len(o.OrderItems)), o.GrandTotal.String()})
        }
        return
    }

    // fetch page
    offset := (page - 1) * perPage
    var orders []models.Order
    fetchQ.Limit(perPage).Offset(offset).Find(&orders)

    type itemView struct {
        ID        string `json:"id"`
        ProductID string `json:"product_id"`
        Name      string `json:"name"`
        Qty       int    `json:"qty"`
        SubTotal  string `json:"sub_total"`
    }
    type userView struct {
        ID        string `json:"id"`
        FirstName string `json:"first_name"`
        LastName  string `json:"last_name"`
        Email     string `json:"email"`
    }
    type orderView struct {
        ID         string      `json:"id"`
        Code       string      `json:"code"`
        User       *userView   `json:"user"`
        UserEmail  string      `json:"user_email"`
        OrderDate  time.Time   `json:"order_date"`
        Items      []itemView  `json:"order_items"`
        ItemsCount int         `json:"items_count"`
        GrandTotal string      `json:"grand_total"`
    }

    var out []orderView
    for _, o := range orders {
        ov := orderView{ID: o.ID, Code: o.Code, OrderDate: o.OrderDate, GrandTotal: o.GrandTotal.String()}
        if o.User.ID != "" {
            ov.User = &userView{ID: o.User.ID, FirstName: o.User.FirstName, LastName: o.User.LastName, Email: o.User.Email}
            ov.UserEmail = o.User.Email
        }
        for _, it := range o.OrderItems {
            iv := itemView{ID: it.ID, ProductID: it.ProductID, Name: it.Name, Qty: it.Qty, SubTotal: it.SubTotal.String()}
            ov.Items = append(ov.Items, iv)
        }
        ov.ItemsCount = len(ov.Items)
        out = append(out, ov)
    }

    // respond with metadata
    _ = ren.JSON(w, http.StatusOK, map[string]interface{}{"orders": out, "meta": map[string]interface{}{"total_count": totalCount, "total_revenue": totalRevenueStr, "revenue_by_status": revenueByStatus, "page": page, "per_page": perPage}})
}

// APIAdminUsers returns JSON list of users (admin-only)
func (server *Server) APIAdminUsers(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    ren := newAdminRender()
    var users []models.User
    server.DB.Order("created_at desc").Find(&users)
    // respond with a lightweight projection to avoid leaking sensitive fields
    type userView struct {
        ID        string    `json:"id"`
        FirstName string    `json:"first_name"`
        LastName  string    `json:"last_name"`
        Email     string    `json:"email"`
        Role      string    `json:"role"`
        CreatedAt time.Time `json:"created_at"`
    }
    var out []userView
    for _, u := range users {
        out = append(out, userView{ID: u.ID, FirstName: u.FirstName, LastName: u.LastName, Email: u.Email, Role: u.Role, CreatedAt: u.CreatedAt})
    }
    _ = ren.JSON(w, http.StatusOK, out)
}

// APIAdminOrders returns JSON list of orders (admin-only)
func (server *Server) APIAdminOrders(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    ren := newAdminRender()
    var orders []models.Order
    server.DB.Preload("OrderItems").Preload("User").Order("created_at desc").Find(&orders)

    type itemView struct {
        ID        string `json:"id"`
        ProductID string `json:"product_id"`
        Name      string `json:"name"`
        Qty       int    `json:"qty"`
        SubTotal  string `json:"sub_total"`
    }
    type userView struct {
        ID        string `json:"id"`
        FirstName string `json:"first_name"`
        LastName  string `json:"last_name"`
        Email     string `json:"email"`
    }
    type orderView struct {
        ID         string      `json:"id"`
        Code       string      `json:"code"`
        User       *userView   `json:"user"`
        UserEmail  string      `json:"user_email"`
        OrderDate  time.Time   `json:"order_date"`
        Items      []itemView  `json:"order_items"`
        ItemsCount int         `json:"items_count"`
        GrandTotal string      `json:"grand_total"`
    }

    var out []orderView
    for _, o := range orders {
        ov := orderView{ID: o.ID, Code: o.Code, OrderDate: o.OrderDate, GrandTotal: o.GrandTotal.String()}
        if o.User.ID != "" {
            ov.User = &userView{ID: o.User.ID, FirstName: o.User.FirstName, LastName: o.User.LastName, Email: o.User.Email}
            ov.UserEmail = o.User.Email
        } else {
            ov.User = nil
            ov.UserEmail = ""
        }
        for _, it := range o.OrderItems {
            iv := itemView{ID: it.ID, ProductID: it.ProductID, Name: it.Name, Qty: it.Qty, SubTotal: it.SubTotal.String()}
            ov.Items = append(ov.Items, iv)
        }
        ov.ItemsCount = len(ov.Items)
        out = append(out, ov)
    }

    _ = ren.JSON(w, http.StatusOK, out)
}

// APIAdminOrder returns a single order by ID (admin-only)
func (server *Server) APIAdminOrder(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    ren := newAdminRender()
    vars := mux.Vars(r)
    id := vars["id"]
    var ord models.Order
    if err := server.DB.Preload("OrderItems").Preload("User").Where("id = ?", id).First(&ord).Error; err != nil {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }

    type itemView struct {
        ID        string `json:"id"`
        ProductID string `json:"product_id"`
        Name      string `json:"name"`
        Qty       int    `json:"qty"`
        SubTotal  string `json:"sub_total"`
    }
    type userView struct {
        ID        string `json:"id"`
        FirstName string `json:"first_name"`
        LastName  string `json:"last_name"`
        Email     string `json:"email"`
    }
    type orderView struct {
        ID         string      `json:"id"`
        Code       string      `json:"code"`
        User       *userView   `json:"user"`
        UserEmail  string      `json:"user_email"`
        OrderDate  time.Time   `json:"order_date"`
        Items      []itemView  `json:"order_items"`
        ItemsCount int         `json:"items_count"`
        GrandTotal string      `json:"grand_total"`
    }

    ov := orderView{ID: ord.ID, Code: ord.Code, OrderDate: ord.OrderDate, GrandTotal: ord.GrandTotal.String()}
    if ord.User.ID != "" {
        ov.User = &userView{ID: ord.User.ID, FirstName: ord.User.FirstName, LastName: ord.User.LastName, Email: ord.User.Email}
        ov.UserEmail = ord.User.Email
    }
    for _, it := range ord.OrderItems {
        iv := itemView{ID: it.ID, ProductID: it.ProductID, Name: it.Name, Qty: it.Qty, SubTotal: it.SubTotal.String()}
        ov.Items = append(ov.Items, iv)
    }
    ov.ItemsCount = len(ov.Items)

    _ = ren.JSON(w, http.StatusOK, ov)
}

// APIAdminProductImageUpload handles image uploads for a product (multipart/form-data)
func (server *Server) APIAdminProductImageUpload(w http.ResponseWriter, r *http.Request) {
    // require POST
    if r.Method != "POST" {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    vars := mux.Vars(r)
    id := vars["id"]
    // ensure product exists
    var p models.Product
    if err := server.DB.Where("id = ?", id).First(&p).Error; err != nil {
        http.Error(w, "product not found", http.StatusNotFound)
        return
    }

    // parse multipart
    if err := r.ParseMultipartForm(10 << 20); err != nil {
        http.Error(w, "failed to parse form", http.StatusBadRequest)
        return
    }
    file, header, err := r.FormFile("image")
    if err != nil {
        http.Error(w, "image file required", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // ensure upload dir
    uploadDir := "public/uploads/products/" + id
    if err := ensureDir(uploadDir); err != nil {
        http.Error(w, "failed to create upload dir", http.StatusInternalServerError)
        return
    }

    // build filename
    ext := ""
    for i := len(header.Filename) - 1; i >= 0; i-- {
        if header.Filename[i] == '.' {
            ext = header.Filename[i:]
            break
        }
    }
    fname := fmt.Sprintf("img_%d%s", time.Now().UnixNano(), ext)
    outPath := uploadDir + "/" + fname

    out, err := createFile(outPath)
    if err != nil {
        http.Error(w, "failed to save file", http.StatusInternalServerError)
        return
    }
    if _, err := io.Copy(out, file); err != nil {
        out.Close()
        http.Error(w, "failed to write file", http.StatusInternalServerError)
        return
    }
    // close file before image processing to avoid file locking on Windows
    _ = out.Close()

    relPath := "/uploads/products/" + id + "/" + fname
    img := models.ProductImage{ID: uuid.New().String(), ProductID: id, Path: relPath}

    // attempt to create resized variants (extra_large, large, medium, small)
    // use the saved file on disk as source
    srcPath := outPath
    // derive name without ext
    nameNoExt := fname
    if ext != "" {
        nameNoExt = strings.TrimSuffix(fname, ext)
    }
    if srcImg, err := imaging.Open(srcPath); err == nil {
        sizes := map[string]int{"ExtraLarge": 1600, "Large": 1024, "Medium": 512, "Small": 128}
        for k, s := range sizes {
            newName := fmt.Sprintf("%s_%s%s", nameNoExt, strings.ToLower(k), ext)
            newPath := filepath.Join(uploadDir, newName)
            dst := imaging.Resize(srcImg, s, 0, imaging.Lanczos)
            if err := imaging.Save(dst, newPath); err == nil {
                rel := "/uploads/products/" + id + "/" + newName
                switch k {
                case "ExtraLarge":
                    img.ExtraLarge = rel
                case "Large":
                    img.Large = rel
                case "Medium":
                    img.Medium = rel
                case "Small":
                    img.Small = rel
                }
            }
            if err := imaging.Save(dst, newPath); err != nil {
                fmt.Println("imaging save error:", err, "path:", newPath)
            }
        }
    } else {
        fmt.Println("imaging open error:", err, "path:", srcPath)
    }

    if err := server.DB.Create(&img).Error; err != nil {
        http.Error(w, "failed to create image record", http.StatusInternalServerError)
        return
    }

    ren := newAdminRender()
    _ = ren.JSON(w, http.StatusCreated, img)
}

// helper small wrappers to avoid importing many packages at top of file
func ensureDir(path string) error {
    return os.MkdirAll(path, 0755)
}

func createFile(path string) (*os.File, error) {
    return os.Create(path)
}


// APIAdminUser handles GET/PUT/DELETE for a single user (admin-only)
func (server *Server) APIAdminUser(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    ren := newAdminRender()
    vars := mux.Vars(r)
    id := vars["id"]
    var u models.User
    if err := server.DB.Where("id = ?", id).First(&u).Error; err != nil {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }
    switch r.Method {
    case "GET":
        // lightweight view
        _ = ren.JSON(w, http.StatusOK, map[string]interface{}{"id": u.ID, "first_name": u.FirstName, "last_name": u.LastName, "email": u.Email, "role": u.Role, "created_at": u.CreatedAt})
        return
    case "PUT":
        var payload map[string]interface{}
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, "invalid json", http.StatusBadRequest)
            return
        }
        updates := map[string]interface{}{}
        if roleVal, ok := payload["role"]; ok {
            if s, ok2 := roleVal.(string); ok2 {
                updates["role"] = s
            }
        }
        if len(updates) == 0 {
            http.Error(w, "nothing to update", http.StatusBadRequest)
            return
        }
        if err := server.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
            http.Error(w, "failed to update", http.StatusInternalServerError)
            return
        }
        server.DB.Where("id = ?", id).First(&u)
        _ = ren.JSON(w, http.StatusOK, map[string]interface{}{"id": u.ID, "first_name": u.FirstName, "last_name": u.LastName, "email": u.Email, "role": u.Role, "created_at": u.CreatedAt})
        return
    case "DELETE":
        if err := server.DB.Where("id = ?", id).Delete(&models.User{}).Error; err != nil {
            http.Error(w, "failed to delete", http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusNoContent)
        return
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
}
