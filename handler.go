package main

import (
	"fmt"
	"forum/functions"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type TemplateHandler struct {
	Tpl *template.Template
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	functions.NoCacheHeaders(w)
	//This block parsing the catogory from the URL, if exist then handling CategoryHandler
	parts := strings.Split(r.URL.Path, "/")
	categoryURL := parts[1]
	fmt.Println(categoryURL)
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
	//if there is no such key, it returns an empty string
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
		fmt.Println("No sort done.")
	}

	// Get categories for posts to display them.
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

	var comments struct{}
	currentCategory := functions.Category{}
	if len(posts) == 0 {
		currentCategory.NoPosts = true
		data := functions.BuildResponse(loggedUser, posts, comments, categories, currentCategory)
		fmt.Println(data)
		tpl.ExecuteTemplate(w, "index.html", data)
		return
	}
	currentCategory.NoPosts = false
	data := functions.BuildResponse(loggedUser, posts, comments, categories, currentCategory)

	w.Header().Set("Content-Type", "text/html")
	tpl.ExecuteTemplate(w, "index.html", data)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received (/register) a request with method:", r.Method)
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
		fmt.Println("Form data:", username, firstname, lastname, email)

		functions.RegisterUserToDb(username, firstname, lastname, passwordHash, email)
		w.Header().Set("Content-Type", "text/html")
		var loggedUser functions.LoggedUser
		loggedUser.IsLoggedIn = false
		loggedUser.WelcomeMessage = "Welcome, you are registered, please login in!"
		data := functions.BuildResponse(loggedUser)
		tpl.ExecuteTemplate(w, "login.html", data)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
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
			fmt.Println("OI EI", err)
		}

		cookieName := "forum"

		functions.NewCookie(w, cookieName, sessionID)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	functions.NoCacheHeaders(w)
	fmt.Println("ERROR in logoutHandler 223")
	loggedUser, err := functions.AuthenticateUser(w, r)
	if err != nil || loggedUser.Id == 0 {
		fmt.Println(err)
		fmt.Println("Error in logoutHandler 237")
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		return
	}

	fmt.Println("Error in logoutHandler 242", loggedUser.Id)
	err = functions.DeleteSessionFromDb(loggedUser.Id)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		return
	}
	functions.RemoveCookieFromClient(w)
	fmt.Printf("Deleted %v's session", loggedUser.Id)
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	w.WriteHeader(301)
}

func CreateAPostHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil || loggedUser.Id == 0 {
			fmt.Println("User is not logged in. Redirecting to login.")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
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

		// TODO: Add categories to html and use them here.

		postTitle := r.FormValue("userPostTitle")
		postBody := r.FormValue("userPostBodyText")
		categories := r.Form["categories"]
		fmt.Println(categories)

		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil || loggedUser.Id == 0 {
			fmt.Println("User is not logged in. Redirecting to login.")
			// TODO: Replace with message to login
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		}

		username, err := functions.GetUserByID(loggedUser.Id)
		if err != nil {
			fmt.Println("Error getting username")
		}
		functions.RegisterPostToDb(loggedUser.Id, postTitle, postBody, username)
		post_id := functions.GetPostByContent(loggedUser.Id, postTitle, postBody)
		fmt.Println(post_id)
		functions.RegisterPostCategoriesToDb(post_id, categories)
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	}
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

func PostHandler(w http.ResponseWriter, r *http.Request) {
	functions.NoCacheHeaders(w)
	parts := strings.Split(r.URL.Path, "/")
	postID := parts[2]
	fmt.Printf("PostID:%s\n", postID)
	post_id, err := strconv.Atoi(postID)
	if err != nil {
		fmt.Println("error in string conversion to int. 326")
	}
	postExists, err := functions.CheckIfPostExists(post_id)
	if err != nil {
		fmt.Println("error checking postID from database")
		return
	}
	if postExists {
		fmt.Println("Post exists")
	} else {
		fmt.Println("Post do not exist")
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

func CreateACommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		loggedUser, err := functions.AuthenticateUser(w, r)
		if err != nil || loggedUser.Id == 0 {
			fmt.Println("User is not logged in. To make a comment, the user must be logged in.")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			//http.Redirect(w, r, "/login?next="+r.URL.RequestURI(), http.StatusSeeOther)
			return
		}

		err = r.ParseForm()
		if err != nil {
			ErrorHandler(w, "Error parsing the form", http.StatusInternalServerError)
			return
		}

		commentBody := r.FormValue("comment")
		fmt.Println(commentBody)

		postIDStr := r.URL.Query().Get("post_id")
		fmt.Println("postIDStr:", postIDStr)
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			ErrorHandler(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		username, err := functions.GetUserByID(loggedUser.Id)
		if err != nil {
			fmt.Println("Error getting username")
		}

		functions.RegisterCommentToDb(loggedUser.Id, postID, commentBody, username)
		http.Redirect(w, r, "/post/"+postIDStr, http.StatusMovedPermanently)
	}
}

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
		fmt.Println("Post_id from url: ", post_id)
		if err != nil {
			return
		}

		commentIDStr := r.URL.Query().Get("comment_id")
		comment_id, err := strconv.Atoi(commentIDStr)
		if err != nil {
			fmt.Println("error converting comment_id")
		}
		fmt.Println("Comment id from url: ", comment_id)

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

func CategoryHandler(w http.ResponseWriter, r *http.Request, categoryURL string) {
	currentCategory, err := functions.GetCurrentCategory(categoryURL)
	if err != nil {
		fmt.Println("Error getting current category.")
		return
	}
	fmt.Println("CurrentCategory: ", currentCategory)
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
		fmt.Println("No sort done.")
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
		fmt.Println(data)
		tpl.ExecuteTemplate(w, "subforum.html", data)
		return

	}

	data := functions.BuildResponse(loggedUser, posts, comments, categories, currentCategory)
	tpl.ExecuteTemplate(w, "subforum.html", data)
}
