package scripthelper

import (
	"os/exec"
)

var RECOMMENDER_SCRIPT_PATH = "recommendations/recommender.py"
var TRAINER_SCRIPT_PATH = "recommendations/trainer.py"

func RunRecommenderScript(userId string) ([]byte, error) {
	cmd := exec.Command("python",  RECOMMENDER_SCRIPT_PATH, userId)
	out, err := cmd.CombinedOutput()
	return out, err
}

func RunTrainerScript() ([]byte, error) {
	cmd := exec.Command("python",  TRAINER_SCRIPT_PATH)
	out, err := cmd.CombinedOutput()
	return out, err
}