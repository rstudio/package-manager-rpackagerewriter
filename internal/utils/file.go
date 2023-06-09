// Copyright (C) 2023 by Posit Software, PBC
package utils

import "os"

// osStat is exposed to make it easy to replace its implementation in tests
var osStat = os.Stat

// FileExists checks if a given file path exists or not.
func FileExists(path string) (b bool, err error) {
	if _, err = osStat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}
