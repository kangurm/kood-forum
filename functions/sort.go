package functions

import (
	"sort"
	"time"
)

// SortByTop function sorting posts by the amount of likes/dislikes on post.
func SortByTop(posts []Post) ([]Post, error) {
	//sorting function. takes two argument. slice to be sorted and less function.
	//wheter the element with [i] should sort before element with index [j]
	sort.Slice(posts, func(i, j int) bool {
		//returns the true if likecount i > likecount j
		return posts[i].LikeCount > posts[j].LikeCount
	})
	//error is always nil because sorting operation cannot fail
	return posts, nil
}

// Sort by date, newest first.
func SortByNew(posts []Post) ([]Post, error) {
	sort.Slice(posts, func(i, j int) bool {
		timeI, errI := time.Parse("2006-01-02 15:04:05", posts[i].Created)
		timeJ, errJ := time.Parse("2006-01-02 15:04:05", posts[j].Created)

		if errI != nil || errJ != nil {
			return false
		}

		return timeI.After(timeJ)
	})

	return posts, nil
}
