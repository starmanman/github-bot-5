//Filename: Main.go
//Author: Nyah Check
//Purpose: GitHub Bot to increase following and print following list.
//Licence: GNU PL 2017


package main

import (

	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/oauth2"

	"github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
)

const (
	// LOGGER is what is printed for help/info output
	LOGGER = "github-bot - %s\n"
	// VERSION is the binary version.
	VERSION = "v1.0"
)

var (
	token    string
	interval string

	lastChecked time.Time

	debug   bool
	version bool
)

type UserDetails struct {
    Login string `json:"login"`
    Id int64 `json:"id"`
    Url int64 `json:"url"`
    HtmlUrl string `json:"html_url"`
}

func init() {
	// parse flags
	flag.StringVar(&token, "token", "", "GitHub API token")
	flag.StringVar(&interval, "interval", "30s", "check interval (ex. 5ms, 10s, 1m, 3h)")

	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(LOGGER, VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	if version {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if token == "" {
		usageAndExit("GitHub token cannot be empty.", 1)
	}
}


func main() {
	var ticker *time.Ticker
	var kmd, usr string
	ctx := context.Background()
	
	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			ticker.Stop()
			logrus.Infof("Received %s, exiting.", sig.String())
			os.Exit(0)
		}
	}()

	// Create the http client.
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	// Create the github client.
	client := github.NewClient(tc)
	

	// Get the authenticated user, the empty string being passed let's the GitHub
	// API know we want ourself.
	user, _, err := client.Users.Get(ctx,"")
	if err != nil {
		logrus.Fatal(err)
	}
	username := *user.Login

	// parse the duration
	dur, err := time.ParseDuration(interval)
	if err != nil {
		logrus.Fatalf("parsing %s as duration failed: %v", interval, err)
	}
	ticker = time.NewTicker(dur)

	logrus.Infof("Bot started for user %s.", username)
	logrus.Infof("Commands: G - Get Followers, I - Follow Users , F - Get Following, U - Unfollow Users, Q - Quit")
	
	fmt.Printf("\nEnter Command: ")
	_, e := fmt.Scanf(" %c", &kmd)
	if err != nil {
	    logrus.Fatal(e)
	}
	
	switch(kmd) {

        case "G": case "g":
        for range ticker.C {
		    page := 1
		    perPage := 30
		    if err = getFollowers(client, username, page, perPage); err != nil {
		        logrus.Fatal(err)
		    }
		
	    }
	    break
	    
	    case "C": case "c":
	    for range ticker.C {
		    page := 1
		    perPage := 30
		    if err = getFollowing(client, username, page, perPage); err != nil {
		        logrus.Fatal(err)
		    }
		
	    }
	    break
    
        case "I": case "i":
        fmt.Printf("\nEnter username(Whose following you wish to follow): ")
	    _, e := fmt.Scanf(" %c", &usr)
	    if e != nil {
	        logrus.Fatal(e)
	    }
	    for range ticker.C {
		    page := 1
		    perPage := 30
		    if e = followUsers(client, usr, page, perPage); e != nil {
		        logrus.Fatal(err)
		    }
		
	    }
	    break
	    
	    case "U": case "u":
        for range ticker.C {
            page := 1
            perPage := 30
            if err := unFollow(client, username, page, perPage); err != nil {
                logrus.Fatal(err)
            }
        }
        break
        
        case "Q": case "q":
            fmt.Printf("\n Exit successful.\n")
            os.Exit(1)
        break
        
        default:
            fmt.Printf("\nInvalid Command.\n")
            os.Exit(1)
    }
}

// getFollowers iterates over all followers received for user.
func getFollowers(client *github.Client, username string, page, perPage int) error {
    ctx := context.Background()
	opt := &github.ListOptions{
			Page:    page,
			PerPage: perPage,
	}
	

    followers, resp, err := client.Users.ListFollowers(ctx, username, opt)
	if err != nil {
		return err
	}

	for _, flwr := range followers {
        
        //writes user details to file.
        saveData("results/followers.txt", flwr)
        fmt.Printf("%+v", flwr)
	}

	// Return early if we are on the last page.
	if page == resp.LastPage || resp.NextPage == 0 {
		return nil
	}

	page = resp.NextPage
	return getFollowers(client, username, page, perPage)
}


//saveData to file.
func saveData(file string, data *github.User) (error) {
     in, err := os.Open(file)
     if err != nil {
           return err
     }
     defer in.Close()

     _, err = fmt.Fprintf(in, "%v", &data)
     return err
}

// getFollowing iterates over the list of following and writes to file.
func getFollowing(client *github.Client, username string, page, perPage int) error {
    ctx := context.Background()
	opt := &github.ListOptions{
			Page:    page,
			PerPage: perPage,
	}
	

    following, resp, err := client.Users.ListFollowing(ctx, username, opt)//to test properly whether to parse resp instead inloop
	if err != nil {
		return err
	}

	for _, flwg := range following {
        
        //writes user details to file.
        saveData("results/following.txt", flwg)
        fmt.Printf("%+v", flwg)
	}

	// Return early if we are on the last page.
	if page == resp.LastPage || resp.NextPage == 0 {
		return nil
	}

	page = resp.NextPage
	return getFollowing(client, username, page, perPage)
}

// followUsers, gets the list of followers for a particular user and followers them on GitHub.
// This requires authentication with the API.
func followUsers(client *github.Client, username string, page, perPage int) error {
    ctx := context.Background()
	opt := &github.ListOptions{
			Page:    page,
			PerPage: perPage,
	}
	
    usrs, resp, err := client.Users.ListFollowing(ctx, username, opt) //to test properly whether to parse resp instead inloop
	if err != nil {
		return err
	}

	for _, usr := range usrs {
		//Follow user
		res, e := client.Users.Follow(ctx, *usr.Login)
        if err != nil {
            panic(e.Error())
        }
        
        fmt.Printf("%+v", res)
	}

	// Return early if we are on the last page.
	if page == resp.LastPage || resp.NextPage == 0 {
		return nil
	}

	page = resp.NextPage
	return followUsers(client, username, page, perPage)
}


// Unfollow all GitHub users on one's follower list.
func unFollow(client *github.Client, username string, page, perPage int) error {
    ctx := context.Background()
	opt := &github.ListOptions{
			Page:    page,
			PerPage: perPage,
	}
	
    usrs, resp, err := client.Users.ListFollowing(ctx, username, opt)//to test properly whether to parse resp instead inloop
	if err != nil {
		return err
	}

	for _, usr := range usrs {
		//Follow user
		res, e := client.Users.Unfollow(ctx, *usr.Login)
        if err != nil {
            panic(e.Error())
        }
        
        fmt.Printf("%+v", res)
	}

	// Return early if we are on the last page.
	if page == resp.LastPage || resp.NextPage == 0 {
		return nil
	}

	page = resp.NextPage
	return unFollow(client, username, page, perPage)
}


func usageAndExit(message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}
