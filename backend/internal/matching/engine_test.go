package matching

import (
	"fmt"
	"math"
	"testing"

	"github.com/google/uuid"
)

func ptr16(v int16) *int16 { return &v }

func baseProfile(id uuid.UUID, gender int16, age int, height int16) Profile {
	return Profile{
		UserID: id, Nickname: "测试用户", Gender: gender, Age: age, HeightCM: ptr16(height),
		CityCode: "310100", EducationLevel: 3, AnnualIncomeBand: 4, MaritalStatus: 1,
		ChildrenCount: 0, WantsChildren: 2, MarriageIntent: 2, SmokingStatus: 1, DrinkingStatus: 1,
	}
}

func loosePreference(genders ...int16) Preference {
	return Preference{
		TargetGenders: genders, MinAge: 18, MaxAge: 99, AgeHard: false,
		MinHeightCM: 0, MaxHeightCM: 250, HeightHard: false,
		MinEducationLevel: 0, MinIncomeBand: 0, MinMarriageIntent: 0,
	}
}

func TestEligible_BothSatisfied(t *testing.T) {
	aID, bID := uuid.New(), uuid.New()
	eng := NewEngine(DefaultWeights())
	a := baseProfile(aID, 1, 30, 178)
	b := baseProfile(bID, 2, 28, 165)
	aPref := loosePreference(2)
	bPref := loosePreference(1)
	ok, conflicts := eng.Eligible(a, aPref, nil, b, bPref, nil)
	if !ok {
		t.Fatalf("双方满足硬条件时应 Eligible，冲突：%+v", conflicts)
	}
}

func TestEligible_OneSidedFails(t *testing.T) {
	// A 满足 B，但 B 不满足 A 的身高硬条件
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 160)

	aPref := loosePreference(2)
	aPref.MinHeightCM, aPref.HeightHard = 170, true
	bPref := loosePreference(1)

	ok, conflicts := eng.Eligible(a, aPref, nil, b, bPref, nil)
	if ok {
		t.Fatal("单向不满足时不得 Eligible")
	}
	found := false
	for _, c := range conflicts {
		if c.Dimension == "height" && c.Direction == "b_to_a" {
			found = true
		}
	}
	if !found {
		t.Fatalf("应记录身高冲突，实际：%+v", conflicts)
	}
}

func TestEligible_HardConflictCannotBeCompensated(t *testing.T) {
	// 一方坚决要孩子，另一方坚决不要：其他全部满足也不 Eligible
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	a.WantsChildren = 2
	b := baseProfile(uuid.New(), 2, 28, 165)
	b.WantsChildren = 1

	aPref := loosePreference(2)
	aPref.AllowedWantsChildren, aPref.WantsChildrenHard = []int16{2}, true
	bPref := loosePreference(1)

	ok, _ := eng.Eligible(a, aPref, nil, b, bPref, nil)
	if ok {
		t.Fatal("生育意愿硬冲突时不得 Eligible")
	}
	res := eng.Evaluate(Input{Me: a, MyPreference: aPref, Candidate: b, PeerPreference: bPref})
	if res.Eligible {
		t.Fatal("Evaluate 也必须判定为不 Eligible")
	}
	if res.OverallScore <= 0 {
		t.Log("综合分仍会计算，但不影响 Eligible 结论")
	}
}

func TestEligible_SoftConflictStillEligible(t *testing.T) {
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 160)
	aPref := loosePreference(2)
	aPref.MinHeightCM, aPref.HeightHard = 170, false // 软条件
	bPref := loosePreference(1)

	ok, _ := eng.Eligible(a, aPref, nil, b, bPref, nil)
	if !ok {
		t.Fatal("软条件不满足不应影响 Eligible")
	}
	res := eng.Evaluate(Input{Me: a, MyPreference: aPref, Candidate: b, PeerPreference: bPref})
	if res.PreferenceScore >= 1 {
		t.Fatalf("软条件不满足时条件契合度应低于 1，实际 %v", res.PreferenceScore)
	}
	if !res.Eligible {
		t.Fatal("软条件冲突时仍应 Eligible")
	}
}

func TestEligible_AgeBoundaryInclusive(t *testing.T) {
	eng := NewEngine(DefaultWeights())
	b := baseProfile(uuid.New(), 2, 28, 165)
	a := baseProfile(uuid.New(), 1, 30, 178)
	aPref := loosePreference(2)
	aPref.MinAge, aPref.MaxAge, aPref.AgeHard = 28, 28, true
	bPref := loosePreference(1)

	ok, conflicts := eng.Eligible(a, aPref, nil, b, bPref, nil)
	if !ok {
		t.Fatalf("年龄边界应包含等号，冲突：%+v", conflicts)
	}
	b.Age = 29
	ok, _ = eng.Eligible(a, aPref, nil, b, bPref, nil)
	if ok {
		t.Fatal("超出年龄上界应不 Eligible")
	}
}

func TestEligible_HeightBoundaryAndUnknown(t *testing.T) {
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 170)
	aPref := loosePreference(2)
	aPref.MinHeightCM, aPref.HeightHard = 170, true
	bPref := loosePreference(1)

	if ok, _ := eng.Eligible(a, aPref, nil, b, bPref, nil); !ok {
		t.Fatal("身高等于下界应通过")
	}
	b.HeightCM = nil
	if ok, _ := eng.Eligible(a, aPref, nil, b, bPref, nil); ok {
		t.Fatal("身高未知且对方有身高硬条件时不得 Eligible")
	}
}

func TestEligible_ArrayConditions(t *testing.T) {
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 165)
	aPref := loosePreference(2)
	aPref.AllowedCityCodes, aPref.CityHard = []string{"310100", "110100"}, true
	bPref := loosePreference(1)

	if ok, _ := eng.Eligible(a, aPref, nil, b, bPref, nil); !ok {
		t.Fatal("城市在允许列表时应通过")
	}
	b.CityCode = "440300"
	if ok, _ := eng.Eligible(a, aPref, nil, b, bPref, nil); ok {
		t.Fatal("城市不在允许列表时不得 Eligible")
	}
}

func TestEligible_IncomeUndisclosedFailsHardIncome(t *testing.T) {
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 165)
	b.AnnualIncomeBand = 8 // 不愿透露
	aPref := loosePreference(2)
	aPref.MinIncomeBand, aPref.IncomeHard = 5, true
	bPref := loosePreference(1)

	if ok, _ := eng.Eligible(a, aPref, nil, b, bPref, nil); ok {
		t.Fatal("收入不愿透露不得视为满足收入硬条件")
	}
}

func TestEligible_HardValueRuleConflict(t *testing.T) {
	qID := uuid.New()
	aRules := []ValueRule{{
		QuestionID: qID, QuestionCode: "finance_household", Category: "财务",
		OwnAnswer: []string{"joint_account"}, AcceptedAnswers: []string{"joint_account", "income_ratio"},
		Importance: 5, IsHard: true,
	}}
	bRules := []ValueRule{{
		QuestionID: qID, QuestionCode: "finance_household", Category: "财务",
		OwnAnswer: []string{"aa"}, AcceptedAnswers: []string{"aa"},
		Importance: 4, IsHard: false,
	}}
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 165)
	aPref, bPref := loosePreference(2), loosePreference(1)

	ok, conflicts := eng.Eligible(a, aPref, aRules, b, bPref, bRules)
	if ok {
		t.Fatalf("价值观硬条件冲突时不得 Eligible：%+v", conflicts)
	}
}

func TestEligible_MissingValueAnswerIsNotAConflict(t *testing.T) {
	qID := uuid.New()
	aRules := []ValueRule{{
		QuestionID: qID, QuestionCode: "chores", Category: "家务",
		OwnAnswer: []string{"equal"}, AcceptedAnswers: []string{"equal"},
		Importance: 5, IsHard: true,
	}}
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 165)
	aPref, bPref := loosePreference(2), loosePreference(1)

	ok, _ := eng.Eligible(a, aPref, aRules, b, bPref, nil)
	if !ok {
		t.Fatal("对方未填写价值观问题时不应视为冲突")
	}
}

func TestEvaluate_OverallWeights(t *testing.T) {
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 178)
	aPref, bPref := loosePreference(2), loosePreference(1)
	aPref.MinEducationLevel, aPref.AgeHard = 3, true

	res := eng.Evaluate(Input{Me: a, MyPreference: aPref, Candidate: b, PeerPreference: bPref})
	want := 0.4*res.PreferenceScore + 0.6*res.CompatibilityScore
	if math.Abs(res.OverallScore-want) > 1e-9 {
		t.Fatalf("综合分权重计算错误：期望 %v 实际 %v", want, res.OverallScore)
	}
}

func makeRules(n int, same bool) ([]ValueRule, []ValueRule) {
	mine := make([]ValueRule, 0, n)
	peer := make([]ValueRule, 0, n)
	for i := 0; i < n; i++ {
		q := uuid.New()
		answerA, answerB := "a", "a"
		if !same {
			answerB = "b"
		}
		mine = append(mine, ValueRule{
			QuestionID: q, QuestionCode: fmt.Sprintf("q%d", i), Category: fmt.Sprintf("分类%d", i%3),
			OwnAnswer: []string{answerA}, AcceptedAnswers: []string{answerA}, Importance: 3,
		})
		peer = append(peer, ValueRule{
			QuestionID: q, QuestionCode: fmt.Sprintf("q%d", i), Category: fmt.Sprintf("分类%d", i%3),
			OwnAnswer: []string{answerB}, AcceptedAnswers: []string{answerB}, Importance: 3,
		})
	}
	return mine, peer
}

func TestEvaluate_CompatibilityInsufficient(t *testing.T) {
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 165)
	mine, peer := makeRules(3, true)
	res := eng.Evaluate(Input{Me: a, MyPreference: loosePreference(2), MyRules: mine,
		Candidate: b, PeerPreference: loosePreference(1), PeerRules: peer})
	if !res.Insufficient {
		t.Fatalf("有效价值观题 %d < %d 时应标记数据不足", res.ValidQuestions, MinValidQuestions)
	}
	if res.ValidQuestions != 3 {
		t.Fatalf("有效题数应为 3，实际 %d", res.ValidQuestions)
	}
}

func TestEvaluate_CompatibilitySufficient(t *testing.T) {
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 165)
	mine, peer := makeRules(8, true)
	res := eng.Evaluate(Input{Me: a, MyPreference: loosePreference(2), MyRules: mine,
		Candidate: b, PeerPreference: loosePreference(1), PeerRules: peer})
	if res.Insufficient {
		t.Fatal("有效题数达到阈值时不应标记数据不足")
	}
	if math.Abs(res.CompatibilityScore-1) > 1e-9 {
		t.Fatalf("答案完全互相接受时应为 1，实际 %v", res.CompatibilityScore)
	}
	if len(res.CategoryScores) == 0 {
		t.Fatal("应输出分类维度得分")
	}
}

func TestEvaluate_MissingAnswersExcludedFromDenominator(t *testing.T) {
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 165)
	mine, peer := makeRules(8, true)
	// 对方只填了 2 道
	res := eng.Evaluate(Input{Me: a, MyPreference: loosePreference(2), MyRules: mine,
		Candidate: b, PeerPreference: loosePreference(1), PeerRules: peer[:2]})
	if res.ValidQuestions != 2 {
		t.Fatalf("未填写的问题不应计入分母，期望 2 实际 %d", res.ValidQuestions)
	}
	if res.Insufficient != true {
		t.Fatal("仅 2 道有效题应标记数据不足")
	}
}

func TestEvaluate_MismatchedAnswersLowerScore(t *testing.T) {
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 165)
	mine, peer := makeRules(8, false)
	res := eng.Evaluate(Input{Me: a, MyPreference: loosePreference(2), MyRules: mine,
		Candidate: b, PeerPreference: loosePreference(1), PeerRules: peer})
	if res.CompatibilityScore >= 0.5 {
		t.Fatalf("答案互不接受时兼容度应较低，实际 %v", res.CompatibilityScore)
	}
}

func TestEvaluate_ReasonsAndDisclaimer(t *testing.T) {
	eng := NewEngine(DefaultWeights())
	a := baseProfile(uuid.New(), 1, 30, 178)
	b := baseProfile(uuid.New(), 2, 28, 165)
	aPref := loosePreference(2)
	bPref := loosePreference(1)
	res := eng.Evaluate(Input{Me: a, MyPreference: aPref, Candidate: b, PeerPreference: bPref})
	if len(res.Reasons) == 0 {
		t.Fatal("应输出匹配解释")
	}
	if Disclaimer == "" || InsufficientText == "" {
		t.Fatal("免责声明与数据不足文案不能为空")
	}
}

func TestWeightsNormalize(t *testing.T) {
	w := Weights{Preference: 3, Compatibility: 3}.Normalize()
	if math.Abs(w.Preference-0.5) > 1e-9 || math.Abs(w.Compatibility-0.5) > 1e-9 {
		t.Fatalf("权重归一化失败：%+v", w)
	}
	empty := Weights{}.Normalize()
	if empty.Preference != 0.4 || empty.Compatibility != 0.6 {
		t.Fatalf("零权重应回退默认值：%+v", empty)
	}
}
