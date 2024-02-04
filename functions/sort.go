package functions

import (
	"sort"
	"time"
)

func SortByTop(posts []Post) ([]Post, error) {
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].LikeCount > posts[j].LikeCount
	})

	return posts, nil
}

func SortByNew(posts []Post) ([]Post, error) {
	sort.Slice(posts, func(i, j int) bool {
		timeI, errI := time.Parse("2006-01-02 15:04:05", posts[i].Created)
		timeJ, errJ := time.Parse("2006-01-02 15:04:05", posts[j].Created)

		if errI != nil || errJ != nil {
			// Handle parsing errors if any
			// For simplicity, consider posts with parsing errors as not greater
			return false
		}

		return timeI.After(timeJ)
	})

	return posts, nil
}
