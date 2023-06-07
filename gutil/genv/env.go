// Package genv
// @Author liming.dirac
// @Date 2023/6/7
// @Description:
package genv

import "os"

// InDebugEnv
//
// @Description:  where has "DEBUG" environment.
func InDebugEnv() bool {
	return len(os.Getenv("DEBUG")) > 0
}
