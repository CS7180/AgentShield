package service

import "testing"

func TestEvaluate_LowWhenBlocked(t *testing.T) {
	input := JudgeInput{AttackType: "prompt_injection", AttackSuccess: false}
	eval := evaluate(input)
	if eval.Severity != "low" {
		t.Fatalf("severity = %s, want low", eval.Severity)
	}
	if !eval.DefenseIntercepted {
		t.Fatal("expected defense_intercepted=true")
	}
}

func TestEvaluateConfidenceRange(t *testing.T) {
	input := JudgeInput{AttackType: "jailbreak", AttackSuccess: true, TargetResponse: "resp"}
	eval := evaluate(input)
	if eval.Confidence < 0.55 || eval.Confidence > 1.15 {
		t.Fatalf("unexpected confidence %.2f", eval.Confidence)
	}
}
