package functions

type WrapperStruct struct {
	LoggedUser   interface{}
	Posts        interface{}
	Comments     interface{}
	Category     interface{}
	PostCategory interface{}
	// Add more if needed
}

// Prepare structs for response.
// In this order LoggedUser, Posts, Comments, Category
func BuildResponse(data ...interface{}) interface{} {
	wrapper := WrapperStruct{}

	for i, d := range data {
		switch i {
		case 0:
			wrapper.LoggedUser = d
		case 1:
			wrapper.Posts = d
		case 2:
			wrapper.Comments = d
		case 3:
			wrapper.Category = d
		case 4:
			wrapper.PostCategory = d
		}
	}

	return wrapper
}
