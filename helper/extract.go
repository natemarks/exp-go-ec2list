package helper

import "strings"

// ExtractName returns the last part of an ARN
func ExtractName(arn string) string {
	parts := strings.Split(arn, "/")
	return parts[len(parts)-1]
}

// ExtractImageName returns the last part of a container image name
func ExtractImageName(image string) string {
	parts := strings.Split(image, "/")
	return parts[len(parts)-1]
}
