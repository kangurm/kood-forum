package main

import (
	"fmt"
	"forum/functions"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	functions.NoCacheHeaders(w)
	//This block parsing the catogory from the URL, if exist then handling CategoryHandler
	parts := strings.Split(r.URL.Path, "/")
	categoryURL := parts[1]
	bCategoryExists := functions.DoesCategoryExist(categoryURL)
	if bCategoryExists {
		CategoryHandler(w, r, categoryURL)
		return
	}

	if r.URL.Path != "/" {
		ErrorHandler(w, "Page not found", http.StatusNotFound)
		return
	}
	//enforce that only GET request are allowed.
	if r.Method != "GET" {
		ErrorHandler(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	// Posts sorting logic: retrieves posts from db and checks for a sorting
	//parameters in the URL query string and sorts the posts accordingly
	posts, err := functions.GetPostsFromDb()
	if err != nil {
		fmt.Println(err)
	}
	//r.URL.Query method returns the first value associated with the given key "sort"
	action := r.URL.Query().Get("sort")
	switch action {
	//sorting by the most liked comments
	case "top":
		posts, err = functions.SortByTop(posts)
		if err != nil {
			fmt.Println("Error sorting")
		}
	case "new":
		posts, err = functions.SortByNew(posts)
		if err != nil {
			fmt.Println("Error sorting")
		}
	default:
	}

	// Get categories for posts to display them on postbar.
	for i, post := range posts {
		category_ids, err := functions.GetAllCategoryIDsForPost(post.Post_id)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			tpl.ExecuteTemplate(w, "index.html", nil)
			return
		}
		categoryNames, err := functions.GetCategoryNamesForPost(category_ids)
		if err != nil {
			w.Header().Set("Content-Type", "text/html")
			tpl.ExecuteTemplate(w, "index.html", nil)
			return
		}
		posts[i].Categories = categoryNames
	}

	loggedUser, err := functions.AuthenticateUser(w, r)
	if err != nil {
		fmt.Println("Not logged in")
		loggedUser.IsLoggedIn = false
	}

	categories, err := functions.GetAllCategoriesFromDb()
	if err != nil {
		fmt.Println("Error getting categories: ", err)
	}
	//this block check if there are posts or no
	//it is needed because to display a text "No posts in this category"
	var comments struct{}
	currentCategory := functions.Category{}
	if len(posts) == 0 {
		currentCategory.NoPosts = true
		data := functions.BuildResponse(loggedUser, posts, comments, categories, currentCategory)
		tpl.ExecuteTemplate(w, "index.html", data)
		return
	}
	currentCategory.NoPosts = false
	data := functions.BuildResponse(loggedUser, posts, comments, categories, currentCategory)

	w.Header().Set("Content-Type", "text/html")
	tpl.ExecuteTemplate(w, "index.html", data)
}

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	//This block is needed to redirect to index if user is already logged in
	if r.Method == "GET" {
		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil {
			loggedUser.IsLoggedIn = false
			data := functions.BuildResponse(loggedUser)
			tpl.ExecuteTemplate(w, "register.html", data)
			return
		}
		fmt.Println("User is already logged in, redirecting to index.")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return

	}
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing the form", http.StatusInternalServerError)
			return
		}

		username := r.FormValue("username")
		firstname := r.FormValue("firstname")
		lastname := r.FormValue("lastname")
		password := r.FormValue("password")
		email := r.FormValue("email")

		exists, err := functions.UserExists(username, email)
		if err != nil {
			http.Error(w, "Error checking user existence", http.StatusInternalServerError)
			return
		}
		if exists {
			w.Header().Set("Content-Type", "text/html")
			var loggedUser functions.LoggedUser
			loggedUser.UserExists = "Username or Email already in use"
			loggedUser.IsLoggedIn = false
			data := functions.BuildResponse(loggedUser)
			tpl.ExecuteTemplate(w, "register.html", data)
			return
		}
		passwordHash, _ := functions.HashPassword(password)

		functions.RegisterUserToDb(username, firstname, lastname, passwordHash, email)
		w.Header().Set("Content-Type", "text/html")
		var loggedUser functions.LoggedUser
		loggedUser.IsLoggedIn = false
		loggedUser.WelcomeMessage = "Welcome, you are registered, please login in!"
		data := functions.BuildResponse(loggedUser)
		tpl.ExecuteTemplate(w, "login.html", data)
	}
}

// LoginHandler handles user login. It authenticates, checks password
// invalidates any existing sessions, generates a new session and sets a new session cookie
func LoginHandler(w http.ResponseWriter, r *http.Request) {

	//This block is needed to redirect to index if user is already logged in
	if r.Method == "GET" {
		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil {
			loggedUser.IsLoggedIn = false
			tpl.ExecuteTemplate(w, "login.html", nil)
			return
		}
		fmt.Println("User is already logged in, redirecting to index.")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if r.Method == "POST" {
		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil {
			loggedUser.IsLoggedIn = false
		}

		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing the form (login)", http.StatusInternalServerError)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		//is needed to check the existing user from db, message on index and error msg on login
		user, err := functions.GetUserByEmail(email)
		if err != nil {
			log.Printf("Error retrieving user: %v\n", err)
			loggedUser.ErrorMessage = "Invalid email or password"
			data := functions.BuildResponse(loggedUser)
			tpl.ExecuteTemplate(w, "login.html", data)
			return
		}

		match := functions.CheckPasswordHash(password, user.Password)
		if !match {
			loggedUser.ErrorMessage = "Invalid email or password"
			data := functions.BuildResponse(loggedUser)
			tpl.ExecuteTemplate(w, "login.html", data)
			return
		}
		//this is done to ensure that the client uses the new session ID
		err = functions.DeleteSessionFromDb(user.Id)
		if err != nil {
			fmt.Println("Failed to delete session from database after user logged in")
		}

		functions.RemoveCookieFromClient(w)

		sessionID, err := functions.GenerateSessionID(user.Password)
		if err != nil {
			ErrorHandler(w, "Error generating session ID", http.StatusInternalServerError)
			return
		}

		err = functions.StoreSessionInDb(sessionID, *user)
		if err != nil {
			fmt.Println("Error in login handler line 227", err)
		}

		cookieName := "forum"

		functions.NewCookie(w, cookieName, sessionID)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

// LogoutHandler authenticates the user, deletes their session from db
// removes the session cookie from the browser
func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	functions.NoCacheHeaders(w)
	//this block activates if there are no usher loged in and redirect to index
	loggedUser, err := functions.AuthenticateUser(w, r)
	if err != nil || loggedUser.Id == 0 {
		fmt.Println(err)
		fmt.Println("Error in logoutHandler 237")
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		return
	}

	err = functions.DeleteSessionFromDb(loggedUser.Id)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		return
	}
	functions.RemoveCookieFromClient(w)
	fmt.Printf("Deleted %v's session \n", loggedUser.Id)
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	//its needed because otherwise would be statuscode 201 what will
	//store cookie and logout would be possible only by deleting manually browser cache.
	//status code 301 indicates a permanent redirect
	w.WriteHeader(301)
}

// CreateAPostHandler handles the process of creating a new post.
func CreateAPostHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil || loggedUser.Id == 0 {
			fmt.Println("User is not logged in. Redirecting to login.")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		//this block retrieves all categories form db,
		//builds a response data struckture containg all the data below,
		//renders a template
		var posts struct{}
		var comments struct{}
		categories, err := functions.GetAllCategoriesFromDb()
		if err != nil {
			fmt.Println("Error getting categories")
		}
		data := functions.BuildResponse(loggedUser, posts, comments, categories)
		tpl.ExecuteTemplate(w, "create-a-post.html", data)
		return
	}
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			ErrorHandler(w, "Error parsing the form", http.StatusInternalServerError)
		}

		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil || loggedUser.Id == 0 {
			fmt.Println("User is not logged in. Redirecting to login.")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		}

		postTitle := r.FormValue("userPostTitle")
		postBody := r.FormValue("userPostBodyText")
		categories := r.Form["categories"]

		postTitle, err = functions.FormatString(postTitle)
		if err != nil {
			loggedUser.ErrorMessage = "Post title cannot be empty."
			var posts struct{}
			var comments struct{}
			categories, err := functions.GetAllCategoriesFromDb()
			if err != nil {
				fmt.Println("Error getting categories")
			}
			data := functions.BuildResponse(loggedUser, posts, comments, categories)
			tpl.ExecuteTemplate(w, "create-a-post.html", data)
			return
		}

		postBody, err = functions.FormatString(postBody)
		if err != nil {
			loggedUser.ErrorMessage = "Post body cannot be empty."
			var posts struct{}
			var comments struct{}
			categories, err := functions.GetAllCategoriesFromDb()
			if err != nil {
				fmt.Println("Error getting categories")
			}
			data := functions.BuildResponse(loggedUser, posts, comments, categories)
			tpl.ExecuteTemplate(w, "create-a-post.html", data)
			return
		}

		username, err := functions.GetUserByID(loggedUser.Id)
		if err != nil {
			fmt.Println("Error getting username")
		}
		functions.RegisterPostToDb(loggedUser.Id, postTitle, postBody, username)
		post_id := functions.GetPostByContent(loggedUser.Id, postTitle, postBody)
		functions.RegisterPostCategoriesToDb(post_id, categories)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

// PostHandler handles the process of retrieving a specifi post and its comments.
func PostHandler(w http.ResponseWriter, r *http.Request) {
	functions.NoCacheHeaders(w)
	//this block takes postid from url and converts it into int
	//its needed because of to make post request from db
	parts := strings.Split(r.URL.Path, "/")
	postID := parts[2]
	post_id, err := strconv.Atoi(postID)
	if err != nil {
		fmt.Println("error in string conversion to int. 326")
	}
	postExists, err := functions.CheckIfPostExists(post_id)
	if err != nil {
		fmt.Println("error checking postID from database")
		return
	}
	if !postExists {
		ErrorHandler(w, "Page not found", http.StatusNotFound)
		return
	}

	currentPost, err := functions.GetPostById(post_id)
	if err != nil {
		fmt.Println("Error getting post info from database")
		ErrorHandler(w, "Server internal error", http.StatusInternalServerError)

	}

	loggedUser, err := functions.AuthenticateUser(w, r)
	if err != nil || loggedUser.Id == 0 {
		loggedUser.IsLoggedIn = false
	}

	currentComments, err := functions.GetCommentsByPostId(post_id)
	if err != nil {
		fmt.Println("Error getting comment info from database")
	}

	data := functions.BuildResponse(loggedUser, currentPost, currentComments)

	tpl.ExecuteTemplate(w, "post.html", data)
}

// CreateACommentHandler handles POST requests to create a comment on a post
// It authendicates the user, parses the form data, validates the data, reggister into database
func CreateACommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil || loggedUser.Id == 0 {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			//http.Redirect(w, r, "/login?next="+r.URL.RequestURI(), http.StatusSeeOther)
			return
		}
		//parses the form data from http request
		err = r.ParseForm()
		if err != nil {
			ErrorHandler(w, "Error parsing the form", http.StatusInternalServerError)
			return
		}

		//this block takes post id from url to connect comment id with post id
		postIDStr := r.URL.Query().Get("post_id")
		post_id, err := strconv.Atoi(postIDStr)
		if err != nil {
			ErrorHandler(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		currentPost, _ := functions.GetPostById(post_id)
		//retrieves the data from comment field
		commentBody := r.FormValue("comment")
		//Check if there are no more than 2 whitespace.
		commentBody, err = functions.FormatString(commentBody)
		if err != nil {
			loggedUser.ErrorMessage = "Comment cannot be empty."

			currentComments, err := functions.GetCommentsByPostId(post_id)
			if err != nil {
				fmt.Println("Error getting comment info from database")
			}

			data := functions.BuildResponse(loggedUser, currentPost, currentComments)
			tpl.ExecuteTemplate(w, "post.html", data)
			return
		}

		username, err := functions.GetUserByID(loggedUser.Id)
		if err != nil {
			fmt.Println("Error getting username")
		}

		functions.RegisterCommentToDb(loggedUser.Id, post_id, commentBody, username)
		http.Redirect(w, r, "/post/"+postIDStr, http.StatusMovedPermanently)
	}
}

// ReactionHandler handles reaction on a post or a comment. It authendicates the user
// retrieves the post id, comment id and action from url query parameters, add reaction to db
func ReactionHandler(w http.ResponseWriter, r *http.Request) {
	functions.NoCacheHeaders(w)
	if r.Method == "GET" {
		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil || loggedUser.Id == 0 {
			fmt.Println("Log in to add reaction")
			loggedUser.IsLoggedIn = false
			loggedUser.ErrorMessage = "Log in to add reaction"
			data := functions.BuildResponse(loggedUser)
			tpl.ExecuteTemplate(w, "login.html", data)
			return
		}

		postIDStr := r.URL.Query().Get("post_id")
		post_id, err := strconv.Atoi(postIDStr)
		if err != nil {
			return
		}

		var comment_id int
		commentIDStr := r.URL.Query().Get("comment_id")
		if !(commentIDStr == "") {
			comment_id, err = strconv.Atoi(commentIDStr)
			if err != nil {
				fmt.Println("error converting comment_id")
			}
		}

		action := r.URL.Query().Get("action")
		var like bool
		switch action {
		case "like":
			like = true
		case "dislike":
			like = false
		}

		functions.AddReaction(post_id, comment_id, loggedUser.Id, like)

		http.Redirect(w, r, "/post/"+postIDStr, http.StatusTemporaryRedirect)
	}
}

// CategoryHandler handles HTTP requests related to a specific category.
// Retrieves the category, its associated posts, all categories, sorts the posts
// and executes subforum.htm
func CategoryHandler(w http.ResponseWriter, r *http.Request, categoryURL string) {
	currentCategory, err := functions.GetCurrentCategory(categoryURL)
	if err != nil {
		fmt.Println("Error getting current category.")
		return
	}
	postIDs, err := functions.GetAllPostIDsByCategory(currentCategory.ID)
	if err != nil {
		fmt.Println("Error getting post ids by category.")
		return
	}
	posts, err := functions.GetAllPostsByPostIDs(postIDs)
	if err != nil {
		fmt.Println("Error getting posts structs for category")
		return
	}

	action := r.URL.Query().Get("sort")
	switch action {
	case "top":
		posts, err = functions.SortByTop(posts)
		if err != nil {
			fmt.Println("Error sorting")
		}
	case "new":
		posts, err = functions.SortByNew(posts)
		if err != nil {
			fmt.Println("Error sorting")
		}
	default:

	}

	loggedUser, err := functions.AuthenticateUser(w, r)
	if err != nil {
		fmt.Println("Error authenticating user")
		loggedUser.IsLoggedIn = false
	}

	categories, err := functions.GetAllCategoriesFromDb()
	if err != nil {
		fmt.Println("Error getting categories")
		return
	}

	var comments struct{}

	if len(posts) == 0 {
		currentCategory.NoPosts = true
		data := functions.BuildResponse(loggedUser, posts, comments, categories, currentCategory)
		tpl.ExecuteTemplate(w, "subforum.html", data)
		return

	}

	data := functions.BuildResponse(loggedUser, posts, comments, categories, currentCategory)
	tpl.ExecuteTemplate(w, "subforum.html", data)
}

func ErrorHandler(w http.ResponseWriter, s string, i int) {

	data := struct {
		StatusCode int
		Message    string
	}{
		StatusCode: i,
		Message:    s,
	}

	w.WriteHeader(i)
	tpl.ExecuteTemplate(w, "error.html", data)
}
