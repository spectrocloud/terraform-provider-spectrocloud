package client

import (
	"fmt"
	"github.com/spectrocloud/hapi/user/client/v1"
	"strings"
	"time"
)

func retryMethod(r v1.ClientService, maxRetries int, retryErrors []string, waitTime time.Duration, fn func() (bool, error)) error {
	if retryErrors == nil {
		retryErrors = []string{
			"Code:ResourceLocked",
		}
	}
	for {
		success, err := fn()
		if err != nil {
			// Check if the error is in the list of retryable errors
			retryable := false
			for _, retryError := range retryErrors {
				if strings.Contains(err.Error(), retryError) {
					retryable = true
					break
				}
			}
			if !retryable {
				return err
			}
		}
		if success {
			return nil
		}
		if maxRetries <= 0 {
			fmt.Errorf("maximum retries reached")
			return err
		}
		maxRetries--
		time.Sleep(waitTime)
	}
}
