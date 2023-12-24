package terrors

import "runtime"

func ZeroLogCallerMarshalFunc(pc uintptr, file string, line int) string {
	pkg, _ := GetPackageAndFuncFromFuncName(runtime.FuncForPC(pc).Name())
	return FormatCaller(pkg, file, line)
}
