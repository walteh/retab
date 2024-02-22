package diff

import (
	"strings"

	"github.com/k0kubun/pp/v3"
	"github.com/kylelemons/godebug/diff"
)

func DiffExportedOnly[T any](want T, got T) string {
	printer := pp.New()
	printer.SetExportedOnly(true)
	printer.SetColoringEnabled(false)

	var abc string

	switch any(want).(type) {
	case string:
		gotarr := strings.Split(any(got).(string), "\n")
		wantarr := strings.Split(any(want).(string), "\n")

		abc = diff.Diff(printer.Sprint(gotarr), printer.Sprint(wantarr))
	default:

		abc = diff.Diff(printer.Sprint(got), printer.Sprint(want))
	}
	if abc == "" {
		return ""
	}

	str := "\n\n"
	str += "to convert ACTUAL ⏩️ EXPECTED:\n\n"
	str += "add:    ➕\n"
	str += "remove: ➖\n"
	str += "\n"
	str += strings.ReplaceAll(strings.ReplaceAll(abc, "\n-", "\n➖"), "\n+", "\n➕")

	return str
}

// func Diff[T any](got T, want T) string {
// 	abc := diff.Diff(fmt.Sprintf("%v", want), fmt.Sprintf("%v", got))
// 	return strings.ReplaceAll(strings.ReplaceAll(abc, "\n-", "\n❌"), "\n+", "\n✅")
// }

// func ExpectEquals[X any](t *testing.T, want X, got X) {
// 	t.Helper()

// 	if reflect.DeepEqual(got, want) {

// 		return
// 	}

// 	diffa := diff.Diff(fmt.Sprintf("%v", got), fmt.Sprintf("%v", want))

// 	if diffa == "" {
// 		return
// 	}

// 	diffb := Diff(want, got)

// 	if diffb == "" {

// 		t.Errorf("%s - %s", t.Name(), diffa)
// 		return
// 	}

// 	t.Errorf("%s - %s", t.Name(), diffb)

// }
