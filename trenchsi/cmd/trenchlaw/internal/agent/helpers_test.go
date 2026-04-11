package agent

import "testing"

func TestAgentCmdRunsOnboardWhenSetupIsIncomplete(t *testing.T) {
	originalOnboardComplete := agentOnboardComplete
	originalRunOnboard := agentRunOnboard
	t.Cleanup(func() {
		agentOnboardComplete = originalOnboardComplete
		agentRunOnboard = originalRunOnboard
	})

	onboardCalled := false
	agentOnboardComplete = func() bool { return false }
	agentRunOnboard = func() error {
		onboardCalled = true
		return nil
	}

	if err := agentCmd("", "", "", false); err != nil {
		t.Fatalf("agentCmd() error = %v", err)
	}

	if !onboardCalled {
		t.Fatal("agentCmd() did not run onboarding when setup was incomplete")
	}
}
