package services

import (
	"path/filepath"
	"testing"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type orderFlowFixture struct {
	db                *gorm.DB
	player            models.User
	admin             models.User
	analyst           models.Analyst
	adminService      *AdminService
	analystService    *AnalystService
	statusHistoryRepo *models.OrderStatusHistoryRepository
}

func setupOrderFlowTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "order-flow.db")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Analyst{},
		&models.Report{},
		&models.ReportVersion{},
		&models.Order{},
		&models.OrderAssignment{},
		&models.OrderStatusHistory{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return db
}

func newOrderFlowFixture(t *testing.T) *orderFlowFixture {
	t.Helper()

	db := setupOrderFlowTestDB(t)
	player := createOrderFlowUser(t, db, "13910000001", models.RoleUser)
	admin := createOrderFlowUser(t, db, "13910000002", models.RoleAdmin)
	analystUser := createOrderFlowUser(t, db, "13910000003", models.RoleAnalyst)
	analyst := models.Analyst{
		UserID: analystUser.ID,
		Name:   "Analyst",
		Status: models.AnalystStatusActive,
	}
	if err := db.Create(&analyst).Error; err != nil {
		t.Fatalf("create analyst: %v", err)
	}

	userRepo := models.NewUserRepository(db)
	reportRepo := models.NewReportRepository(db)
	orderRepo := models.NewOrderRepository(db)
	analystRepo := models.NewAnalystRepository(db)
	assignmentRepo := models.NewOrderAssignmentRepository(db)
	statusHistoryRepo := models.NewOrderStatusHistoryRepository(db)

	return &orderFlowFixture{
		db:      db,
		player:  player,
		admin:   admin,
		analyst: analyst,
		adminService: NewAdminService(
			userRepo,
			reportRepo,
			orderRepo,
			analystRepo,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			assignmentRepo,
			statusHistoryRepo,
		),
		analystService:    NewAnalystService(analystRepo, orderRepo, userRepo, assignmentRepo, statusHistoryRepo),
		statusHistoryRepo: statusHistoryRepo,
	}
}

func createOrderFlowUser(t *testing.T, db *gorm.DB, phone string, role models.UserRole) models.User {
	t.Helper()

	user := models.User{
		Phone:    phone,
		Password: "test-password",
		Name:     string(role),
		Role:     role,
		Status:   models.StatusActive,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user %s: %v", phone, err)
	}
	return user
}

func createUploadedOrder(t *testing.T, db *gorm.DB, userID uint, orderNo string) models.Order {
	t.Helper()

	order := models.Order{
		UserID:     userID,
		OrderNo:    orderNo,
		Amount:     99,
		Status:     models.OrderStatusUploaded,
		OrderType:  "basic",
		PlayerName: "Demo Player",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order %s: %v", orderNo, err)
	}
	return order
}

func TestOrderAssignmentFlowAcceptCreatesAssignmentAndHistory(t *testing.T) {
	fx := newOrderFlowFixture(t)
	order := createUploadedOrder(t, fx.db, fx.player.ID, "FLOW-ACCEPT-001")

	assignedOrder, err := fx.adminService.AssignOrder(order.ID, fx.analyst.ID, fx.admin.ID)
	if err != nil {
		t.Fatalf("assign order: %v", err)
	}
	if assignedOrder.Status != models.OrderStatusAssigned {
		t.Fatalf("expected assigned order status, got %s", assignedOrder.Status)
	}
	if assignedOrder.AnalystID == nil || *assignedOrder.AnalystID != fx.analyst.ID {
		t.Fatalf("expected analyst %d on assigned order, got %#v", fx.analyst.ID, assignedOrder.AnalystID)
	}
	if assignedOrder.AssignedAt == nil || assignedOrder.Deadline == nil {
		t.Fatalf("expected assigned_at and deadline to be set")
	}

	var assignment models.OrderAssignment
	if err := fx.db.Where("order_id = ? AND analyst_id = ?", order.ID, fx.analyst.ID).First(&assignment).Error; err != nil {
		t.Fatalf("find assignment: %v", err)
	}
	if assignment.Status != models.OrderAssignmentStatusPending {
		t.Fatalf("expected pending assignment, got %s", assignment.Status)
	}
	if assignment.AssignedBy == nil || *assignment.AssignedBy != fx.admin.ID {
		t.Fatalf("expected assigned_by %d, got %#v", fx.admin.ID, assignment.AssignedBy)
	}

	histories, err := fx.statusHistoryRepo.FindByOrderID(order.ID)
	if err != nil {
		t.Fatalf("find histories after assign: %v", err)
	}
	if len(histories) != 1 {
		t.Fatalf("expected 1 status history after assign, got %d", len(histories))
	}
	if histories[0].FromStatus != models.OrderStatusUploaded || histories[0].ToStatus != models.OrderStatusAssigned {
		t.Fatalf("expected uploaded -> assigned history, got %s -> %s", histories[0].FromStatus, histories[0].ToStatus)
	}
	if histories[0].ActorID == nil || *histories[0].ActorID != fx.admin.ID || histories[0].ActorRole != "admin" {
		t.Fatalf("expected admin actor on assign history, got %#v/%s", histories[0].ActorID, histories[0].ActorRole)
	}

	if err := fx.analystService.AcceptOrder(fx.analyst.ID, order.ID); err != nil {
		t.Fatalf("accept order: %v", err)
	}

	var updated models.Order
	if err := fx.db.First(&updated, order.ID).Error; err != nil {
		t.Fatalf("find updated order: %v", err)
	}
	if updated.Status != models.OrderStatusProcessing {
		t.Fatalf("expected processing order after accept, got %s", updated.Status)
	}
	if updated.AcceptedAt == nil {
		t.Fatalf("expected accepted_at to be set")
	}

	if err := fx.db.First(&assignment, assignment.ID).Error; err != nil {
		t.Fatalf("reload assignment: %v", err)
	}
	if assignment.Status != models.OrderAssignmentStatusAccepted {
		t.Fatalf("expected accepted assignment, got %s", assignment.Status)
	}
	if assignment.RespondedAt == nil {
		t.Fatalf("expected responded_at to be set")
	}
	if assignment.RejectedReason != "" {
		t.Fatalf("expected empty rejected reason after accept, got %q", assignment.RejectedReason)
	}

	histories, err = fx.statusHistoryRepo.FindByOrderID(order.ID)
	if err != nil {
		t.Fatalf("find histories after accept: %v", err)
	}
	if len(histories) != 2 {
		t.Fatalf("expected 2 status histories after accept, got %d", len(histories))
	}
	if histories[1].FromStatus != models.OrderStatusAssigned || histories[1].ToStatus != models.OrderStatusProcessing {
		t.Fatalf("expected assigned -> processing history, got %s -> %s", histories[1].FromStatus, histories[1].ToStatus)
	}
	if histories[1].ActorID == nil || *histories[1].ActorID != fx.analyst.ID || histories[1].ActorRole != "analyst" {
		t.Fatalf("expected analyst actor on accept history, got %#v/%s", histories[1].ActorID, histories[1].ActorRole)
	}
	if histories[1].Reason != "分析师接单" {
		t.Fatalf("expected accept reason, got %q", histories[1].Reason)
	}
}

func TestOrderAssignmentFlowRejectResetsOrderAndRecordsReason(t *testing.T) {
	fx := newOrderFlowFixture(t)
	order := createUploadedOrder(t, fx.db, fx.player.ID, "FLOW-REJECT-001")

	if _, err := fx.adminService.AssignOrder(order.ID, fx.analyst.ID, fx.admin.ID); err != nil {
		t.Fatalf("assign order: %v", err)
	}

	reason := "schedule conflict"
	if err := fx.analystService.RejectOrder(fx.analyst.ID, order.ID, reason); err != nil {
		t.Fatalf("reject order: %v", err)
	}

	var updated models.Order
	if err := fx.db.First(&updated, order.ID).Error; err != nil {
		t.Fatalf("find updated order: %v", err)
	}
	if updated.Status != models.OrderStatusUploaded {
		t.Fatalf("expected uploaded order after reject, got %s", updated.Status)
	}
	if updated.AnalystID != nil || updated.AssignedAt != nil || updated.AcceptedAt != nil || updated.Deadline != nil {
		t.Fatalf("expected analyst assignment fields to be cleared after reject")
	}
	if updated.CancelReason != reason {
		t.Fatalf("expected cancel reason %q, got %q", reason, updated.CancelReason)
	}

	var assignment models.OrderAssignment
	if err := fx.db.Where("order_id = ? AND analyst_id = ?", order.ID, fx.analyst.ID).First(&assignment).Error; err != nil {
		t.Fatalf("find assignment: %v", err)
	}
	if assignment.Status != models.OrderAssignmentStatusRejected {
		t.Fatalf("expected rejected assignment, got %s", assignment.Status)
	}
	if assignment.RejectedReason != reason {
		t.Fatalf("expected rejected reason %q, got %q", reason, assignment.RejectedReason)
	}
	if assignment.RespondedAt == nil {
		t.Fatalf("expected responded_at to be set")
	}

	histories, err := fx.statusHistoryRepo.FindByOrderID(order.ID)
	if err != nil {
		t.Fatalf("find histories after reject: %v", err)
	}
	if len(histories) != 2 {
		t.Fatalf("expected 2 status histories after reject, got %d", len(histories))
	}
	if histories[1].FromStatus != models.OrderStatusAssigned || histories[1].ToStatus != models.OrderStatusUploaded {
		t.Fatalf("expected assigned -> uploaded history, got %s -> %s", histories[1].FromStatus, histories[1].ToStatus)
	}
	if histories[1].ActorID == nil || *histories[1].ActorID != fx.analyst.ID || histories[1].ActorRole != "analyst" {
		t.Fatalf("expected analyst actor on reject history, got %#v/%s", histories[1].ActorID, histories[1].ActorRole)
	}
	if histories[1].Reason != reason {
		t.Fatalf("expected reject reason %q, got %q", reason, histories[1].Reason)
	}
}

func TestReportReviewCompletesOrderAfterAdminApproval(t *testing.T) {
	fx := newOrderFlowFixture(t)
	order := createUploadedOrder(t, fx.db, fx.player.ID, "FLOW-REPORT-APPROVE-001")

	if _, err := fx.adminService.AssignOrder(order.ID, fx.analyst.ID, fx.admin.ID); err != nil {
		t.Fatalf("assign order: %v", err)
	}
	if err := fx.analystService.AcceptOrder(fx.analyst.ID, order.ID); err != nil {
		t.Fatalf("accept order: %v", err)
	}

	report := models.Report{
		OrderID:        order.ID,
		UserID:         fx.player.ID,
		AnalystID:      fx.analyst.ID,
		PlayerName:     "Demo Player",
		PlayerPosition: "winger",
		Content:        "submitted report",
		Status:         models.ReportStatusProcessing,
	}
	if err := fx.db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}

	if err := fx.adminService.ReviewReport(report.ID, models.ReportStatusCompleted, "", fx.admin.ID); err != nil {
		t.Fatalf("review report: %v", err)
	}

	var updatedOrder models.Order
	if err := fx.db.First(&updatedOrder, order.ID).Error; err != nil {
		t.Fatalf("find order: %v", err)
	}
	if updatedOrder.Status != models.OrderStatusCompleted {
		t.Fatalf("expected completed order after report approval, got %s", updatedOrder.Status)
	}
	if updatedOrder.ReportID == nil || *updatedOrder.ReportID != report.ID {
		t.Fatalf("expected order report_id %d, got %#v", report.ID, updatedOrder.ReportID)
	}
	if updatedOrder.CompletedAt == nil {
		t.Fatalf("expected completed_at after report approval")
	}

	var updatedReport models.Report
	if err := fx.db.First(&updatedReport, report.ID).Error; err != nil {
		t.Fatalf("find report: %v", err)
	}
	if updatedReport.Status != models.ReportStatusCompleted {
		t.Fatalf("expected completed report, got %s", updatedReport.Status)
	}

	histories, err := fx.statusHistoryRepo.FindByOrderID(order.ID)
	if err != nil {
		t.Fatalf("find histories: %v", err)
	}
	if len(histories) != 3 {
		t.Fatalf("expected 3 histories after approval, got %d", len(histories))
	}
	last := histories[2]
	if last.FromStatus != models.OrderStatusProcessing || last.ToStatus != models.OrderStatusCompleted {
		t.Fatalf("expected processing -> completed history, got %s -> %s", last.FromStatus, last.ToStatus)
	}
	if last.ActorID == nil || *last.ActorID != fx.admin.ID || last.ActorRole != "admin" {
		t.Fatalf("expected admin actor on approval history, got %#v/%s", last.ActorID, last.ActorRole)
	}
}

func TestReportReviewRejectsReportWithoutCompletingOrder(t *testing.T) {
	fx := newOrderFlowFixture(t)
	order := createUploadedOrder(t, fx.db, fx.player.ID, "FLOW-REPORT-REJECT-001")

	if _, err := fx.adminService.AssignOrder(order.ID, fx.analyst.ID, fx.admin.ID); err != nil {
		t.Fatalf("assign order: %v", err)
	}
	if err := fx.analystService.AcceptOrder(fx.analyst.ID, order.ID); err != nil {
		t.Fatalf("accept order: %v", err)
	}

	report := models.Report{
		OrderID:        order.ID,
		UserID:         fx.player.ID,
		AnalystID:      fx.analyst.ID,
		PlayerName:     "Demo Player",
		PlayerPosition: "winger",
		Content:        "submitted report",
		Status:         models.ReportStatusProcessing,
	}
	if err := fx.db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}

	remark := "needs more detail"
	if err := fx.adminService.ReviewReport(report.ID, models.ReportStatusFailed, remark, fx.admin.ID); err != nil {
		t.Fatalf("review report: %v", err)
	}

	var updatedOrder models.Order
	if err := fx.db.First(&updatedOrder, order.ID).Error; err != nil {
		t.Fatalf("find order: %v", err)
	}
	if updatedOrder.Status != models.OrderStatusProcessing {
		t.Fatalf("expected processing order after report rejection, got %s", updatedOrder.Status)
	}
	if updatedOrder.CompletedAt != nil {
		t.Fatalf("expected completed_at to remain empty after rejection")
	}

	var updatedReport models.Report
	if err := fx.db.First(&updatedReport, report.ID).Error; err != nil {
		t.Fatalf("find report: %v", err)
	}
	if updatedReport.Status != models.ReportStatusFailed {
		t.Fatalf("expected failed report, got %s", updatedReport.Status)
	}
	if updatedReport.ReviewRemark != remark {
		t.Fatalf("expected review remark %q, got %q", remark, updatedReport.ReviewRemark)
	}
}

func TestAdminGetPendingReportsReturnsProcessingReports(t *testing.T) {
	fx := newOrderFlowFixture(t)
	order := createUploadedOrder(t, fx.db, fx.player.ID, "FLOW-PENDING-REPORT-001")
	report := models.Report{
		OrderID:        order.ID,
		UserID:         fx.player.ID,
		AnalystID:      fx.analyst.ID,
		PlayerName:     "Demo Player",
		PlayerPosition: "winger",
		Content:        "submitted report",
		Status:         models.ReportStatusProcessing,
	}
	if err := fx.db.Create(&report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}

	reports, total, err := fx.adminService.GetPendingReports(1, 10)
	if err != nil {
		t.Fatalf("get pending reports: %v", err)
	}
	if total != 1 || len(reports) != 1 || reports[0].ID != report.ID {
		t.Fatalf("expected one pending report %d, got total=%d list=%v", report.ID, total, reports)
	}
}
