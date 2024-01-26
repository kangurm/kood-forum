package functions

import (
	"sort"
)

func SortByTop() ([]Post, error) {
	posts, err := GetPostsFromDb()
	if err != nil {
		return nil, err
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].LikeCount > posts[j].LikeCount
	})

	return posts, nil
}

func SortByNew() {

}

func SortByHot() {

}
