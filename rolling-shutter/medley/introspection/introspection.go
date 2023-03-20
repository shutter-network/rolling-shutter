package introspection

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

func GetFuncName(skip int) string {
	frm := GetFrame(skip)
	name := frm.Func.Name()
	name = name[1+strings.LastIndex(name, "."):]
	return name
}

func GetFrame(skip int) runtime.Frame {
	pc := make([]uintptr, 10)
	runtime.Callers(skip, pc)
	frms := runtime.CallersFrames(pc)
	frm, _ := frms.Next()
	return frm
}

type CallerInfo struct {
	FileLocation string
	Function     string
	Package      string
	Module       string
	Library      string
}

var mainModulePath string

func init() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	mainModulePath = bi.Main.Path
}

func GetCallerInfo(skip int) CallerInfo {
	frame := GetFrame(skip)
	fullPath := frame.Function

	idx := strings.LastIndex(fullPath, "/")
	fncWithModule := fullPath[1+idx:]

	idx = strings.LastIndex(fncWithModule, ".")
	fncName := fncWithModule[1+idx:]
	pkg := fncWithModule[:idx]

	pathSpec := fullPath[:idx]
	modulePath, _ := strings.CutPrefix(pathSpec, mainModulePath)
	// for the edgecase where we couldn't introspect the path
	if mainModulePath == "" {
		mainModulePath = "shutter"
	}

	return CallerInfo{
		FileLocation: fmt.Sprintf("%s:%d", frame.File, frame.Line),
		Function:     fncName,
		Package:      pkg,
		Module:       modulePath,
		Library:      mainModulePath,
	}
}
