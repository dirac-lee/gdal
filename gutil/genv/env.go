// Package genv
// @Author liming.dirac
// @Date 2023/6/7
// @Description:
package genv

import "os"

func InDebugEnv() bool {
	return len(os.Getenv("DEBUG")) > 0
}
