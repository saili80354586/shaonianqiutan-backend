package repositories

import (
	"fmt"
	"testing"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	// Auto migrate required tables
	if err := db.AutoMigrate(
		&models.User{},
		&models.Club{},
		&models.Coach{},
		&models.Analyst{},
		&models.Scout{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func seedTestData(t *testing.T, db *gorm.DB) {
	// Users (players)
	users := []models.User{
		{Phone: "13800000001", Name: "Player1", Role: models.RoleUser, Status: models.StatusActive, Province: "上海", City: "上海市", Position: "前锋", Age: 12},
		{Phone: "13800000002", Name: "Player2", Role: models.RoleUser, Status: models.StatusActive, Province: "上海", City: "上海市", Position: "中场", Age: 13},
		{Phone: "13800000003", Name: "Player3", Role: models.RoleUser, Status: models.StatusActive, Province: "北京", City: "北京市", Position: "后卫", Age: 14},
		{Phone: "13800000004", Name: "Player4", Role: models.RoleUser, Status: models.StatusActive, Province: "广东", City: "广州市", Position: "门将", Age: 11},
	}
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	// Clubs
	clubs := []models.Club{
		{UserID: users[0].ID, Name: "Club1", Province: "上海", City: "上海市", ClubSize: "large"},
		{UserID: users[1].ID, Name: "Club2", Province: "上海", City: "上海市", ClubSize: "medium"},
		{UserID: users[2].ID, Name: "Club3", Province: "北京", City: "北京市", ClubSize: "small"},
	}
	for i := range clubs {
		if err := db.Create(&clubs[i]).Error; err != nil {
			t.Fatalf("create club: %v", err)
		}
	}

	// Coaches (linked to users)
	coaches := []models.Coach{
		{UserID: users[0].ID, LicenseType: "A级", CoachingYears: 10, CurrentClub: "Club1", City: "上海市"},
		{UserID: users[1].ID, LicenseType: "B级", CoachingYears: 5, CurrentClub: "Club2", City: "上海市"},
		{UserID: users[2].ID, LicenseType: "A级", CoachingYears: 8, CurrentClub: "Club3", City: "北京市"},
	}
	for i := range coaches {
		if err := db.Create(&coaches[i]).Error; err != nil {
			t.Fatalf("create coach: %v", err)
		}
	}

	// Analysts (linked to users)
	analysts := []models.Analyst{
		{UserID: users[0].ID, Name: "Analyst1", Specialty: "技术", Experience: 5, Rating: 4.5},
		{UserID: users[1].ID, Name: "Analyst2", Specialty: "战术", Experience: 3, Rating: 4.0},
		{UserID: users[2].ID, Name: "Analyst3", Specialty: "技术", Experience: 7, Rating: 4.8},
	}
	for i := range analysts {
		if err := db.Create(&analysts[i]).Error; err != nil {
			t.Fatalf("create analyst: %v", err)
		}
	}

	// Scouts (linked to users)
	scouts := []models.Scout{
		{UserID: users[0].ID, ScoutingExperience: "5-10", CurrentOrganization: "Org1", TotalDiscovered: 20, TotalReports: 10, TotalAdopted: 8},
		{UserID: users[1].ID, ScoutingExperience: "3-5", CurrentOrganization: "Org2", TotalDiscovered: 15, TotalReports: 5, TotalAdopted: 3},
		{UserID: users[2].ID, ScoutingExperience: "10+", CurrentOrganization: "Org3", TotalDiscovered: 50, TotalReports: 20, TotalAdopted: 15},
	}
	for i := range scouts {
		if err := db.Create(&scouts[i]).Error; err != nil {
			t.Fatalf("create scout: %v", err)
		}
	}
}

// ===== National Tests =====

func TestGetNationalAggregates_Players(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, err := repo.GetNationalAggregates(LayerPlayers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agg) != 3 {
		t.Errorf("expected 3 provinces, got %d", len(agg))
	}

	m := make(map[string]int64)
	for _, a := range agg {
		m[a.ProvinceName] = a.Count
	}
	if m["上海"] != 2 {
		t.Errorf("expected 上海=2, got %d", m["上海"])
	}
	if m["北京"] != 1 {
		t.Errorf("expected 北京=1, got %d", m["北京"])
	}
	if m["广东"] != 1 {
		t.Errorf("expected 广东=1, got %d", m["广东"])
	}
}

func TestGetNationalAggregates_Clubs(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, err := repo.GetNationalAggregates(LayerClubs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agg) != 2 {
		t.Errorf("expected 2 provinces, got %d", len(agg))
	}

	m := make(map[string]int64)
	for _, a := range agg {
		m[a.ProvinceName] = a.Count
	}
	if m["上海"] != 2 {
		t.Errorf("expected 上海=2 clubs, got %d", m["上海"])
	}
	if m["北京"] != 1 {
		t.Errorf("expected 北京=1 club, got %d", m["北京"])
	}
}

func TestGetNationalAggregates_Coaches(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, err := repo.GetNationalAggregates(LayerCoaches)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := make(map[string]int64)
	for _, a := range agg {
		m[a.ProvinceName] = a.Count
	}
	if m["上海"] != 2 {
		t.Errorf("expected 上海=2 coaches, got %d", m["上海"])
	}
}

func TestGetNationalAggregates_Analysts(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, err := repo.GetNationalAggregates(LayerAnalysts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := make(map[string]int64)
	for _, a := range agg {
		m[a.ProvinceName] = a.Count
	}
	if m["上海"] != 2 {
		t.Errorf("expected 上海=2 analysts, got %d", m["上海"])
	}
}

func TestGetNationalAggregates_Scouts(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, err := repo.GetNationalAggregates(LayerScouts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := make(map[string]int64)
	for _, a := range agg {
		m[a.ProvinceName] = a.Count
	}
	if m["上海"] != 2 {
		t.Errorf("expected 上海=2 scouts, got %d", m["上海"])
	}
}

func TestGetNationalAggregates_All(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, err := repo.GetNationalAggregates(LayerAll)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agg) != 3 {
		t.Errorf("expected 3 provinces, got %d", len(agg))
	}

	m := make(map[string]int64)
	for _, a := range agg {
		m[a.ProvinceName] = a.Count
	}
	// 上海: 2 players + 2 clubs + 2 coaches + 2 analysts + 2 scouts = 10
	if m["上海"] != 10 {
		t.Errorf("expected 上海=10 total, got %d", m["上海"])
	}
	// 北京: 1+1+1+1+1 = 5
	if m["北京"] != 5 {
		t.Errorf("expected 北京=5 total, got %d", m["北京"])
	}
	// 广东: 1+0+0+0+0 = 1
	if m["广东"] != 1 {
		t.Errorf("expected 广东=1 total, got %d", m["广东"])
	}
}

// ===== Provincial Tests =====

func TestGetProvincialAggregates_Players(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, err := repo.GetProvincialAggregates("上海", LayerPlayers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agg) != 1 {
		t.Errorf("expected 1 city, got %d", len(agg))
	}
	if agg[0].CityName != "上海市" || agg[0].Count != 2 {
		t.Errorf("expected 上海市=2, got %s=%d", agg[0].CityName, agg[0].Count)
	}
}

func TestGetProvincialAggregates_All(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, err := repo.GetProvincialAggregates("上海", LayerAll)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agg) != 1 {
		t.Errorf("expected 1 city, got %d", len(agg))
	}
	if agg[0].PlayerCount != 2 || agg[0].ClubCount != 2 || agg[0].CoachCount != 2 {
		t.Errorf("unexpected counts: %+v", agg[0])
	}
}

// ===== City Tests =====

func TestGetCityEntities_Players(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	items, err := repo.GetCityEntities("上海", "上海市", LayerPlayers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 players, got %d", len(items))
	}
	for _, item := range items {
		if item.Type != "player" {
			t.Errorf("expected type=player, got %s", item.Type)
		}
	}
}

func TestGetCityEntities_Clubs(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	items, err := repo.GetCityEntities("上海", "上海市", LayerClubs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 clubs, got %d", len(items))
	}
	for _, item := range items {
		if item.Type != "club" {
			t.Errorf("expected type=club, got %s", item.Type)
		}
	}
}

func TestGetCityEntities_Coaches(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	items, err := repo.GetCityEntities("上海", "上海市", LayerCoaches)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 coaches, got %d", len(items))
	}
}

func TestGetCityEntities_Analysts(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	items, err := repo.GetCityEntities("上海", "上海市", LayerAnalysts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 analysts, got %d", len(items))
	}
}

func TestGetCityEntities_Scouts(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	items, err := repo.GetCityEntities("上海", "上海市", LayerScouts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 scouts, got %d", len(items))
	}
}

func TestGetCityEntities_All(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	items, err := repo.GetCityEntities("上海", "上海市", LayerAll)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2 players + 2 clubs + 2 coaches + 2 analysts + 2 scouts = 10
	if len(items) != 10 {
		t.Errorf("expected 10 entities, got %d", len(items))
	}
}

// ===== P2-14~P2-17 Fill Method Tests =====

func TestFillSizeDistributionForNational(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, _ := repo.GetNationalAggregates(LayerClubs)
	if len(agg) == 0 {
		t.Fatal("no aggregates")
	}
	for _, a := range agg {
		if a.ProvinceName == "上海" {
			if a.SizeDistribution == nil {
				t.Fatal("expected SizeDistribution for 上海")
			}
			if a.SizeDistribution["large"] != 1 || a.SizeDistribution["medium"] != 1 {
				t.Errorf("unexpected size distribution for 上海: %+v", a.SizeDistribution)
			}
		}
		if a.ProvinceName == "北京" {
			if a.SizeDistribution == nil || a.SizeDistribution["small"] != 1 {
				t.Errorf("expected small=1 for 北京, got %+v", a.SizeDistribution)
			}
		}
	}
}

func TestFillLicenseDistributionForNational(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, _ := repo.GetNationalAggregates(LayerCoaches)
	for _, a := range agg {
		if a.ProvinceName == "上海" {
			if a.LicenseDistribution == nil {
				t.Fatal("expected LicenseDistribution for 上海")
			}
			if a.LicenseDistribution["A级"] != 1 || a.LicenseDistribution["B级"] != 1 {
				t.Errorf("unexpected license distribution for 上海: %+v", a.LicenseDistribution)
			}
		}
	}
}

func TestFillSpecialtyDistributionForNational(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, _ := repo.GetNationalAggregates(LayerAnalysts)
	for _, a := range agg {
		if a.ProvinceName == "上海" {
			if a.SpecialtyDistribution == nil {
				t.Fatal("expected SpecialtyDistribution for 上海")
			}
			if a.SpecialtyDistribution["技术"] != 1 || a.SpecialtyDistribution["战术"] != 1 {
				t.Errorf("unexpected specialty distribution for 上海: %+v", a.SpecialtyDistribution)
			}
		}
	}
}

func TestFillAdoptionRateForNational(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, _ := repo.GetNationalAggregates(LayerScouts)
	for _, a := range agg {
		if a.ProvinceName == "上海" {
			// Scout1: 8/10 = 0.8, Scout2: 3/5 = 0.6, avg = 0.7
			expected := 0.7
			if a.AdoptionRate != expected {
				t.Errorf("expected adoptionRate=%.2f for 上海, got %.2f", expected, a.AdoptionRate)
			}
		}
		if a.ProvinceName == "北京" {
			// Scout3: 15/20 = 0.75
			expected := 0.75
			if a.AdoptionRate != expected {
				t.Errorf("expected adoptionRate=%.2f for 北京, got %.2f", expected, a.AdoptionRate)
			}
		}
	}
}

// ===== Provincial Fill Tests =====

func TestFillSizeDistributionForProvincial(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, _ := repo.GetProvincialAggregates("上海", LayerClubs)
	if len(agg) != 1 {
		t.Fatalf("expected 1 city, got %d", len(agg))
	}
	if agg[0].SizeDistribution == nil {
		t.Fatal("expected SizeDistribution")
	}
	if agg[0].SizeDistribution["large"] != 1 || agg[0].SizeDistribution["medium"] != 1 {
		t.Errorf("unexpected size distribution: %+v", agg[0].SizeDistribution)
	}
}

func TestFillLicenseDistributionForProvincial(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, _ := repo.GetProvincialAggregates("上海", LayerCoaches)
	if len(agg) != 1 {
		t.Fatalf("expected 1 city, got %d", len(agg))
	}
	if agg[0].LicenseDistribution == nil {
		t.Fatal("expected LicenseDistribution")
	}
	if agg[0].LicenseDistribution["A级"] != 1 || agg[0].LicenseDistribution["B级"] != 1 {
		t.Errorf("unexpected license distribution: %+v", agg[0].LicenseDistribution)
	}
}

func TestFillSpecialtyDistributionForProvincial(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, _ := repo.GetProvincialAggregates("上海", LayerAnalysts)
	if len(agg) != 1 {
		t.Fatalf("expected 1 city, got %d", len(agg))
	}
	if agg[0].SpecialtyDistribution == nil {
		t.Fatal("expected SpecialtyDistribution")
	}
	if agg[0].SpecialtyDistribution["技术"] != 1 || agg[0].SpecialtyDistribution["战术"] != 1 {
		t.Errorf("unexpected specialty distribution: %+v", agg[0].SpecialtyDistribution)
	}
}

func TestFillAdoptionRateForProvincial(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	agg, _ := repo.GetProvincialAggregates("上海", LayerScouts)
	if len(agg) != 1 {
		t.Fatalf("expected 1 city, got %d", len(agg))
	}
	// avg(8/10, 3/5) = avg(0.8, 0.6) = 0.7
	if agg[0].AdoptionRate != 0.7 {
		t.Errorf("expected adoptionRate=0.7, got %.2f", agg[0].AdoptionRate)
	}
}

// ===== Edge Cases =====

func TestEmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMultiLayerMapRepository(db)

	agg, err := repo.GetNationalAggregates(LayerPlayers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agg) != 0 {
		t.Errorf("expected 0 aggregates for empty db, got %d", len(agg))
	}
}

// ===== Performance Tests (P2-19) =====

func seedLargeDataset(t *testing.T, db *gorm.DB) {
	provinces := []string{"北京", "上海", "广东", "浙江", "江苏", "山东", "河南", "四川", "湖北", "湖南"}
	cities := []string{"北京市", "上海市", "广州市", "杭州市", "南京市", "济南市", "郑州市", "成都市", "武汉市", "长沙市"}
	sizes := []string{"small", "medium", "large"}
	licenses := []string{"A级", "B级", "C级"}
	specialties := []string{"技术", "战术", "体能", "心理"}

	// Create 2000 users (players)
	for i := 0; i < 2000; i++ {
		u := models.User{
			Phone:    fmt.Sprintf("139%08d", i),
			Name:     fmt.Sprintf("Player%d", i),
			Role:     models.RoleUser,
			Status:   models.StatusActive,
			Province: provinces[i%len(provinces)],
			City:     cities[i%len(cities)],
			Position: "前锋",
			Age:      12 + i%6,
		}
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	// Create 500 clubs
	for i := 0; i < 500; i++ {
		c := models.Club{
			UserID:   uint(i + 1),
			Name:     fmt.Sprintf("Club%d", i),
			Province: provinces[i%len(provinces)],
			City:     cities[i%len(cities)],
			ClubSize: sizes[i%len(sizes)],
		}
		if err := db.Create(&c).Error; err != nil {
			t.Fatalf("create club: %v", err)
		}
	}

	// Create 500 coaches
	for i := 0; i < 500; i++ {
		c := models.Coach{
			UserID:        uint(i + 1),
			LicenseType:   licenses[i%len(licenses)],
			CoachingYears: 5 + i%15,
			CurrentClub:   fmt.Sprintf("Club%d", i),
			City:          cities[i%len(cities)],
		}
		if err := db.Create(&c).Error; err != nil {
			t.Fatalf("create coach: %v", err)
		}
	}

	// Create 500 analysts
	for i := 0; i < 500; i++ {
		a := models.Analyst{
			UserID:    uint(i + 1),
			Name:      fmt.Sprintf("Analyst%d", i),
			Specialty: specialties[i%len(specialties)],
			Experience: 1 + i%10,
			Rating:    3.0 + float64(i%20)/10.0,
		}
		if err := db.Create(&a).Error; err != nil {
			t.Fatalf("create analyst: %v", err)
		}
	}

	// Create 500 scouts
	for i := 0; i < 500; i++ {
		s := models.Scout{
			UserID:             uint(i + 1),
			ScoutingExperience: "3-5",
			CurrentOrganization: fmt.Sprintf("Org%d", i),
			TotalDiscovered:    10 + i%50,
			TotalReports:       5 + i%20,
			TotalAdopted:       1 + i%10,
		}
		if err := db.Create(&s).Error; err != nil {
			t.Fatalf("create scout: %v", err)
		}
	}
}

func TestPerformance_NationalAggregation(t *testing.T) {
	db := setupTestDB(t)
	seedLargeDataset(t, db)
	repo := NewMultiLayerMapRepository(db)

	start := time.Now()
	agg, err := repo.GetNationalAggregates(LayerAll)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agg) == 0 {
		t.Fatal("expected non-empty aggregates")
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("national aggregation too slow: %v (limit: 500ms)", elapsed)
	}
	t.Logf("National aggregation (all layers, ~4000 entities): %v", elapsed)
}

func TestPerformance_NationalSingleLayer(t *testing.T) {
	db := setupTestDB(t)
	seedLargeDataset(t, db)
	repo := NewMultiLayerMapRepository(db)

	layers := []EntityLayer{LayerPlayers, LayerClubs, LayerCoaches, LayerAnalysts, LayerScouts}
	for _, layer := range layers {
		start := time.Now()
		_, err := repo.GetNationalAggregates(layer)
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("unexpected error for %s: %v", layer, err)
		}
		if elapsed > 500*time.Millisecond {
			t.Errorf("national %s aggregation too slow: %v (limit: 500ms)", layer, elapsed)
		}
		t.Logf("National %s aggregation: %v", layer, elapsed)
	}
}

func TestPerformance_ProvincialAggregation(t *testing.T) {
	db := setupTestDB(t)
	seedLargeDataset(t, db)
	repo := NewMultiLayerMapRepository(db)

	start := time.Now()
	_, err := repo.GetProvincialAggregates("上海", LayerAll)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("provincial aggregation too slow: %v (limit: 500ms)", elapsed)
	}
	t.Logf("Provincial aggregation (all layers, 上海): %v", elapsed)
}

func TestPerformance_CityEntities(t *testing.T) {
	db := setupTestDB(t)
	seedLargeDataset(t, db)
	repo := NewMultiLayerMapRepository(db)

	start := time.Now()
	items, err := repo.GetCityEntities("上海", "上海市", LayerPlayers)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("city entities query too slow: %v (limit: 500ms)", elapsed)
	}
	t.Logf("City entities (上海/上海市/players): %d items in %v", len(items), elapsed)
}

func TestInvalidLayerDefaultsToPlayers(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)
	repo := NewMultiLayerMapRepository(db)

	layer := ParseEntityLayer("invalid")
	if layer != LayerPlayers {
		t.Errorf("expected default LayerPlayers, got %s", layer)
	}
	agg, err := repo.GetNationalAggregates(layer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(agg) != 3 {
		t.Errorf("expected 3 provinces, got %d", len(agg))
	}
}
