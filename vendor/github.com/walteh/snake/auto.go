package snake

import (
	"context"
	"os/exec"
	"strings"

	"github.com/walteh/terrors"
)

func AutoENVVar(ctx context.Context, str string) (string, error) {

	if strings.HasPrefix(str, "snake-auto(") {

		rep := strings.Replace(str, "snake-auto(", "", -1)
		rep = strings.TrimSuffix(rep, ")")
		strs := strings.Split(rep, " ")

		for i, s := range strs {
			strs[i] = strings.TrimPrefix(s, "'")
			strs[i] = strings.TrimSuffix(strs[i], "'")
		}

		extra := make([]string, 0)
		if len(strs) > 1 {
			extra = strs[1:]
		}

		cmd := exec.Command(strs[0], extra...)
		res, err := cmd.CombinedOutput()
		if err != nil {
			return "", terrors.Wrap(err, string(res))
		}

		str = strings.TrimSpace(string(res))
	}

	return str, nil
}
