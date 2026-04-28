package models

import (
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupOrderAssignmentTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "order-assignment.db")), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := db.AutoMigrate(
		&User{},
		&Analyst{},
		&Report{},
		&Order{},
		&OrderAssignment{},
		&OrderStatusHistory{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return db
}

func TestBackfillOrderAssignmentsFromOrders(t *testing.T) {
	db := setupOrderAssignmentTestDB(t)
	user := User{
		Phone:    "13920000001",
		Password: "test-password",
		Name:     "Player",
		Role:     RoleUser,
		Status:   StatusActive,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	analystUser := User{
		Phone:    "13920000002",
		Password: "test-password",
		Name:     "Analyst User",
		Role:     RoleAnalyst,
		Status:   StatusActive,
	}
	if err := db.Create(&analystUser).Error; err != nil {
		t.Fatalf("create analyst user: %v", err)
	}
	analyst := Analyst{
		UserID: analystUser.ID,
		Name:   "Analyst",
		Status: AnalystStatusActive,
	}
	if err := db.Create(&analyst).Error; err != nil {
		t.Fatalf("create analyst: %v", err)
	}

	analystID := analyst.ID
	assignedAt := time.Date(2026, 4, 25, 9, 0, 0, 0, time.UTC)
	acceptedAt := assignedAt.Add(2 * time.Hour)
	updatedAt := assignedAt.Add(3 * time.Hour)
	orders := []Order{
		{
			UserID:     user.ID,
			AnalystID:  &analystID,
			OrderNo:    "BACKFILL-ASSIGNED",
			Amount:     99,
			Status:     OrderStatusAssigned,
			AssignedAt: &assignedAt,
			UpdatedAt:  updatedAt,
		},
		{
			UserID:     user.ID,
			AnalystID:  &analystID,
			OrderNo:    "BACKFILL-PROCESSING",
			Amount:     199,
			Status:     OrderStatusProcessing,
			AssignedAt: &assignedAt,
			AcceptedAt: &acceptedAt,
			UpdatedAt:  updatedAt,
		},
		{
			UserID:       user.ID,
			AnalystID:    &analystID,
			OrderNo:      "BACKFILL-CANCELLED",
			Amount:       299,
			Status:       OrderStatusCancelled,
			AssignedAt:   &assignedAt,
			CancelReason: "no capacity",
			UpdatedAt:    updatedAt,
		},
		{
			UserID:  user.ID,
			OrderNo: "BACKFILL-UNASSIGNED",
			Amount:  399,
			Status:  OrderStatusUploaded,
		},
	}
	for i := range orders {
		if err := db.Create(&orders[i]).Error; err != nil {
			t.Fatalf("create order %s: %v", orders[i].OrderNo, err)
		}
	}

	if err := BackfillOrderAssignmentsFromOrders(db); err != nil {
		t.Fatalf("backfill assignments: %v", err)
	}

	var assignments []OrderAssignment
	if err := db.Order("order_id ASC").Find(&assignments).Error; err != nil {
		t.Fatalf("find assignments: %v", err)
	}
	if len(assignments) != 3 {
		t.Fatalf("expected 3 backfilled assignments, got %d", len(assignments))
	}

	byOrderID := map[uint]OrderAssignment{}
	for _, assignment := range assignments {
		byOrderID[assignment.OrderID] = assignment
	}
	if byOrderID[orders[0].ID].Status != OrderAssignmentStatusPending {
		t.Fatalf("expected assigned order to backfill as pending, got %s", byOrderID[orders[0].ID].Status)
	}
	processingAssignment := byOrderID[orders[1].ID]
	if processingAssignment.Status != OrderAssignmentStatusAccepted {
		t.Fatalf("expected processing order to backfill as accepted, got %s", processingAssignment.Status)
	}
	if processingAssignment.RespondedAt == nil || !processingAssignment.RespondedAt.Equal(acceptedAt) {
		t.Fatalf("expected processing assignment responded_at %s, got %#v", acceptedAt, processingAssignment.RespondedAt)
	}
	cancelledAssignment := byOrderID[orders[2].ID]
	if cancelledAssignment.Status != OrderAssignmentStatusRejected {
		t.Fatalf("expected cancelled order with reason to backfill as rejected, got %s", cancelledAssignment.Status)
	}
	if cancelledAssignment.RejectedReason != "no capacity" {
		t.Fatalf("expected rejected reason to be copied, got %q", cancelledAssignment.RejectedReason)
	}
	if _, exists := byOrderID[orders[3].ID]; exists {
		t.Fatalf("expected unassigned order to be skipped")
	}

	if err := BackfillOrderAssignmentsFromOrders(db); err != nil {
		t.Fatalf("backfill assignments second time: %v", err)
	}

	var count int64
	if err := db.Model(&OrderAssignment{}).Count(&count).Error; err != nil {
		t.Fatalf("count assignments: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected backfill to be idempotent, got %d assignments", count)
	}
}

func TestOrderAssignmentRepositoryMarkLatestPendingWithTx(t *testing.T) {
	db := setupOrderAssignmentTestDB(t)
	repo := NewOrderAssignmentRepository(db)
	now := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	assignments := []OrderAssignment{
		{
			OrderID:    1,
			AnalystID:  10,
			AssignedAt: now.Add(-2 * time.Hour),
			Status:     OrderAssignmentStatusPending,
		},
		{
			OrderID:    1,
			AnalystID:  10,
			AssignedAt: now.Add(-1 * time.Hour),
			Status:     OrderAssignmentStatusPending,
		},
		{
			OrderID:    1,
			AnalystID:  10,
			AssignedAt: now.Add(-3 * time.Hour),
			Status:     OrderAssignmentStatusAccepted,
		},
	}
	for i := range assignments {
		if err := db.Create(&assignments[i]).Error; err != nil {
			t.Fatalf("create assignment %d: %v", i, err)
		}
	}

	reason := "schedule conflict"
	if err := repo.MarkLatestPendingWithTx(nil, 1, 10, OrderAssignmentStatusRejected, reason, now); err != nil {
		t.Fatalf("mark latest pending: %v", err)
	}

	var oldPending OrderAssignment
	if err := db.First(&oldPending, assignments[0].ID).Error; err != nil {
		t.Fatalf("find old pending assignment: %v", err)
	}
	if oldPending.Status != OrderAssignmentStatusPending {
		t.Fatalf("expected old pending assignment to stay pending, got %s", oldPending.Status)
	}

	var latestPending OrderAssignment
	if err := db.First(&latestPending, assignments[1].ID).Error; err != nil {
		t.Fatalf("find latest pending assignment: %v", err)
	}
	if latestPending.Status != OrderAssignmentStatusRejected {
		t.Fatalf("expected latest pending assignment to be rejected, got %s", latestPending.Status)
	}
	if latestPending.RejectedReason != reason {
		t.Fatalf("expected rejected reason %q, got %q", reason, latestPending.RejectedReason)
	}
	if latestPending.RespondedAt == nil || !latestPending.RespondedAt.Equal(now) {
		t.Fatalf("expected responded_at %s, got %#v", now, latestPending.RespondedAt)
	}

	var accepted OrderAssignment
	if err := db.First(&accepted, assignments[2].ID).Error; err != nil {
		t.Fatalf("find accepted assignment: %v", err)
	}
	if accepted.Status != OrderAssignmentStatusAccepted {
		t.Fatalf("expected accepted assignment to stay accepted, got %s", accepted.Status)
	}
}
