# forum

## Author

### Rain Praks (rpraks) and Marcus Kangur (mkangur)
#### <a href="https://01.kood.tech/git/mkangur/forum">My Gitea Repo</a>

## Info

This web forum project enables user communication through posts and comments, supports categorization of posts, and allows users to like or dislike both posts and comments. It features filtering options to navigate through content easily.

## Accessing the Forum

Pre-saved users for login: email: username@username.ee, password: username  
Example: username: mati, email: mati@mati.ee, password: mati.

New Users: Feel free to register a new account directly on the forum.


## See [audit requirements]: 
<a href="https://github.com/01-edu/public/tree/master/subjects/forum/audit">Audit Repo</a>

## How to view forum webpage

1. **Clone the project:** 
` git clone https://01.kood.tech/git/mkangur/forum` 


<a href=https://01.kood.tech/git/mkangur/forum></a>

2. **Install Docker. Skip steps 2-4 if you already have docker installed on your machine.**

Download Docker Desktop and follow instruction from webpage.

<a href="https://docs.docker.com/desktop/">Docker Desktop</a>


3. **Ensure that docker works with following command.**

` docker run hello-world `

Output should look like this: <br>
<sub> 
Hello from Docker!                                                         <br>
This message shows that your installation appears to be working correctly. <br>
...                                                                        <br>
</sub>

4. **Ensure youre in the forum folder.**

5. **Build the image with following command. PS! Ending with a dot. forum can be named to anything you want**

` docker image build -f Dockerfile -t forum . `

Output should look like this: <br>
<sub>
 ...                                           <br>
 Successfully built 72fac098e532               <br>
 Successfully tagged forum:latest              <br>
</sub>


6. **Start the container using the image just created**

` docker container run -p 8080:8080 -d --name testcontainer forum `


7. **Stop the container from running**

` docker stop testcontainer `

8. **Delete the created newly created container and image.**

` docker rmi -f forum      `                <br>

` docker rm -f testcontainer `              <br>

9. **You can run dockertest.sh file what autamatically run all steps and starts the server.**

` chmod +x dockertest.sh `

` ./dockertest.sh `
