package models

type ScoreDimensionDetail struct {
	GroupKey   string  `json:"group_key"`
	GroupLabel string  `json:"group_label"`
	FieldKey   string  `json:"field_key"`
	FieldLabel string  `json:"field_label"`
	Score      float64 `json:"score"`
	Comment    string  `json:"comment"`
}

type scoreDimensionMeta struct {
	groupKey   string
	groupLabel string
	fieldKey   string
	fieldLabel string
}

var scoreDimensionMetas = []scoreDimensionMeta{
	{groupKey: "overall", groupLabel: "整体表现", fieldKey: "ball_control", fieldLabel: "控球能力"},
	{groupKey: "overall", groupLabel: "整体表现", fieldKey: "off_ball_movement", fieldLabel: "无球跑动"},
	{groupKey: "overall", groupLabel: "整体表现", fieldKey: "pressing_awareness", fieldLabel: "逼抢意识"},
	{groupKey: "overall", groupLabel: "整体表现", fieldKey: "positioning", fieldLabel: "站位/选位"},
	{groupKey: "offense", groupLabel: "进攻能力", fieldKey: "width_participation", fieldLabel: "拉开宽度参与"},
	{groupKey: "offense", groupLabel: "进攻能力", fieldKey: "off_ball_support", fieldLabel: "无球支援"},
	{groupKey: "offense", groupLabel: "进攻能力", fieldKey: "one_v_one", fieldLabel: "1v1过人能力"},
	{groupKey: "offense", groupLabel: "进攻能力", fieldKey: "crossing_assist", fieldLabel: "传中/助攻"},
	{groupKey: "offense", groupLabel: "进攻能力", fieldKey: "combat_ability", fieldLabel: "对抗能力"},
	{groupKey: "offense", groupLabel: "进攻能力", fieldKey: "pace_rhythm", fieldLabel: "节奏把控"},
	{groupKey: "offense", groupLabel: "进攻能力", fieldKey: "pass_vision", fieldLabel: "传球视野"},
	{groupKey: "offense", groupLabel: "进攻能力", fieldKey: "body_posture", fieldLabel: "身体姿态"},
	{groupKey: "defense", groupLabel: "防守能力", fieldKey: "defensive_commitment", fieldLabel: "防守投入度"},
	{groupKey: "defense", groupLabel: "防守能力", fieldKey: "loss_recovery", fieldLabel: "丢球回追"},
	{groupKey: "defense", groupLabel: "防守能力", fieldKey: "teammate_coordination", fieldLabel: "队友协防配合"},
	{groupKey: "defense", groupLabel: "防守能力", fieldKey: "second_ball", fieldLabel: "二点球争抢"},
	{groupKey: "defense", groupLabel: "防守能力", fieldKey: "aerial_duel", fieldLabel: "空中争顶"},
	{groupKey: "defense", groupLabel: "防守能力", fieldKey: "defensive_shape", fieldLabel: "防守阵型保持"},
	{groupKey: "defense", groupLabel: "防守能力", fieldKey: "role_adjustment", fieldLabel: "角色调整能力"},
	{groupKey: "defense", groupLabel: "防守能力", fieldKey: "defensive_rhythm", fieldLabel: "防守节奏"},
}

func ScoreDimensionMetas() []ScoreDimensionDetail {
	result := make([]ScoreDimensionDetail, 0, len(scoreDimensionMetas))
	for _, meta := range scoreDimensionMetas {
		result = append(result, ScoreDimensionDetail{
			GroupKey:   meta.groupKey,
			GroupLabel: meta.groupLabel,
			FieldKey:   meta.fieldKey,
			FieldLabel: meta.fieldLabel,
		})
	}
	return result
}

func ScoreDimensionDetails(scores *VideoAnalysisScores) []ScoreDimensionDetail {
	if scores == nil {
		scores = NewDefaultScores()
	}
	lookup := map[string]RatingDimension{
		"ball_control":          scores.BallControl,
		"off_ball_movement":     scores.OffBallMovement,
		"pressing_awareness":    scores.PressingAwareness,
		"positioning":           scores.Positioning,
		"width_participation":   scores.WidthParticipation,
		"off_ball_support":      scores.OffBallSupport,
		"one_v_one":             scores.OneVOne,
		"crossing_assist":       scores.CrossingAssist,
		"combat_ability":        scores.CombatAbility,
		"pace_rhythm":           scores.PaceRhythm,
		"pass_vision":           scores.PassVision,
		"body_posture":          scores.BodyPosture,
		"defensive_commitment":  scores.DefensiveCommitment,
		"loss_recovery":         scores.LossRecovery,
		"teammate_coordination": scores.TeammateCoordination,
		"second_ball":           scores.SecondBall,
		"aerial_duel":           scores.AerialDuel,
		"defensive_shape":       scores.DefensiveShape,
		"role_adjustment":       scores.RoleAdjustment,
		"defensive_rhythm":      scores.DefensiveRhythm,
	}
	result := make([]ScoreDimensionDetail, 0, len(scoreDimensionMetas))
	for _, meta := range scoreDimensionMetas {
		dimension := lookup[meta.fieldKey]
		result = append(result, ScoreDimensionDetail{
			GroupKey:   meta.groupKey,
			GroupLabel: meta.groupLabel,
			FieldKey:   meta.fieldKey,
			FieldLabel: meta.fieldLabel,
			Score:      dimension.Score,
			Comment:    dimension.Comment,
		})
	}
	return result
}
