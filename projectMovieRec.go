//300148490
//Dera Ramiliarijaona

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Threshold for “liked” movies
const iLiked float64 = 3.5

// Minimal number of users who must like a movie
const K = 10

// Number of top recommendations to keep
const N = 20

// Recommendation holds data for each recommended movie
type Recommendation struct {
	userID     int
	movieID    int
	movieTitle string
	score      float32
	nUsers     int
}

// Probability-like measure if needed
func (r Recommendation) getProbLike() float32 {
	if r.nUsers == 0 {
		return 0
	}
	return r.score / float32(r.nUsers)
}

// User with ID and lists of liked / notLiked
type User struct {
	userID   int
	liked    []int
	notLiked []int
}

// Read ratings into a map of userID -> *User
func readRatingsCSV(fileName string) (map[int]*User, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	// Skip header
	if _, err = reader.Read(); err != nil {
		return nil, err
	}

	users := make(map[int]*User)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, row := range records {
		if len(row) != 4 {
			return nil, fmt.Errorf("invalid row length: %v", row)
		}
		uID, _ := strconv.Atoi(row[0])
		mID, _ := strconv.Atoi(row[1])
		rating, _ := strconv.ParseFloat(row[2], 64)

		u, ok := users[uID]
		if !ok {
			u = &User{userID: uID}
			users[uID] = u
		}
		if rating >= iLiked {
			u.liked = append(u.liked, mID)
		} else {
			u.notLiked = append(u.notLiked, mID)
		}
	}
	return users, nil
}

// Read movies into a map of movieID -> title
func readMoviesCSV(fileName string) (map[int]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	// Skip header
	if _, err = reader.Read(); err != nil {
		return nil, err
	}

	movies := make(map[int]string)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	for _, row := range records {
		if len(row) != 3 {
			return nil, fmt.Errorf("invalid row length: %v", row)
		}
		mID, _ := strconv.Atoi(row[0])
		movies[mID] = row[1] // ignoring row[2] if it's genres
	}
	return movies, nil
}

// Build indices for quick lookups
func buildLikeCountIndex(users map[int]*User) map[int]int {
	likeCount := make(map[int]int)
	for _, u := range users {
		for _, mID := range u.liked {
			likeCount[mID]++
		}
	}
	return likeCount
}
func buildMovieToUsersIndex(users map[int]*User) map[int][]int {
	mIndex := make(map[int][]int)
	for uid, u := range users {
		for _, mID := range u.liked {
			mIndex[mID] = append(mIndex[mID], uid)
		}
	}
	return mIndex
}

// Helpers for Jaccard
func intersection(a, b []int) []int {
	setB := make(map[int]bool)
	for _, x := range b {
		setB[x] = true
	}
	var out []int
	for _, x := range a {
		if setB[x] {
			out = append(out, x)
		}
	}
	return out
}
func union(slices ...[]int) []int {
	set := make(map[int]bool)
	for _, s := range slices {
		for _, x := range s {
			set[x] = true
		}
	}
	var out []int
	for x := range set {
		out = append(out, x)
	}
	return out
}
func computeJaccard(u1, u2 *User) float32 {
	interLiked := intersection(u1.liked, u2.liked)
	interNotLiked := intersection(u1.notLiked, u2.notLiked)
	uni := union(u1.liked, u1.notLiked, u2.liked, u2.notLiked)
	if len(uni) == 0 {
		return 0
	}
	return float32(len(interLiked)+len(interNotLiked)) / float32(len(uni))
}

// Pipeline Stage 1: Generate all recommendations
func generateMovieRec(wg *sync.WaitGroup, stop <-chan bool, userID int, titles map[int]string) <-chan Recommendation {
	out := make(chan Recommendation)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		for mID, title := range titles {
			select {
			case <-stop:
				return
			case out <- Recommendation{userID, mID, title, 0, 0}:
			}
		}
	}()
	return out
}

// Pipeline Stage 2: Filter out movies the user has already seen
func filterAlreadySeen(wg *sync.WaitGroup, stop <-chan bool, in <-chan Recommendation, user *User) <-chan Recommendation {
	out := make(chan Recommendation)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		for rec := range in {
			select {
			case <-stop:
				return
			default:
				seenLiked := false
				for _, lm := range user.liked {
					if lm == rec.movieID {
						seenLiked = true
						break
					}
				}
				seenNotLiked := false
				for _, lm := range user.notLiked {
					if lm == rec.movieID {
						seenNotLiked = true
						break
					}
				}
				if !seenLiked && !seenNotLiked {
					out <- rec
				}
			}
		}
	}()
	return out
}

// Pipeline Stage 3: Filter out movies liked by fewer than K users
func filterByK(wg *sync.WaitGroup, stop <-chan bool, in <-chan Recommendation, likeCount map[int]int) <-chan Recommendation {
	out := make(chan Recommendation)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		for rec := range in {
			select {
			case <-stop:
				return
			default:
				if likeCount[rec.movieID] >= K {
					rec.nUsers = likeCount[rec.movieID]
					out <- rec
				}
			}
		}
	}()
	return out
}

// Pipeline Stage 4: Compute score (Jaccard) in parallel
func computeScoreStage(wg *sync.WaitGroup, stop <-chan bool, in <-chan Recommendation, out chan<- Recommendation, users map[int]*User, movieToUsers map[int][]int, target *User) {
	defer wg.Done()
	for rec := range in {
		select {
		case <-stop:
			return
		default:
			var sum float32
			whoLike := movieToUsers[rec.movieID]
			for _, uid := range whoLike {
				if uid != target.userID {
					sum += computeJaccard(target, users[uid])
				}
			}
			if len(whoLike) > 0 {
				rec.score = sum / float32(len(whoLike))
			}
			out <- rec
		}
	}
}

// Collect results, keep top N
func collectRecommendations(in <-chan Recommendation) []Recommendation {
	var recs []Recommendation
	for r := range in {
		recs = append(recs, r)
	}
	sort.Slice(recs, func(i, j int) bool {
		return recs[i].score > recs[j].score
	})
	if len(recs) > N {
		return recs[:N]
	}
	return recs
}

func main() {
	fmt.Println("Number of CPUs:", runtime.NumCPU())

	var currentUserID int
	fmt.Print("Recommendations for which user? ")
	fmt.Scanf("%d", &currentUserID)

	titles, err := readMoviesCSV("movies.csv")
	if err != nil {
		log.Fatal(err)
	}
	users, err := readRatingsCSV("ratings.csv")
	if err != nil {
		log.Fatal(err)
	}
	user := users[currentUserID]
	if user == nil {
		fmt.Printf("User %d not found!\n", currentUserID)
		return
	}

	likeCount := buildLikeCountIndex(users)
	movieToUsers := buildMovieToUsersIndex(users)

	stop := make(chan bool)
	var wg sync.WaitGroup

	start := time.Now()

	// Stage 1
	recCh := generateMovieRec(&wg, stop, currentUserID, titles)

	// Stage 2
	filt1Ch := filterAlreadySeen(&wg, stop, recCh, user)

	// Stage 3
	filt2Ch := filterByK(&wg, stop, filt1Ch, likeCount)

	// Stage 4: two parallel goroutines
	scoredCh := make(chan Recommendation)
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go computeScoreStage(&wg, stop, filt2Ch, scoredCh, users, movieToUsers, user)
	}

	// Collect
	var finalRecs []Recommendation
	collectorDone := make(chan bool)
	go func() {
		finalRecs = collectRecommendations(scoredCh)
		close(collectorDone)
	}()

	wg.Wait()
	close(scoredCh)
	<-collectorDone

	end := time.Now()
	close(stop)

	fmt.Printf("\nRecommendations for user #%d:\n", currentUserID)
	for _, rec := range finalRecs {
		// format: Title at 0.1075 [ 10]
		fmt.Printf("%s at %.4f [ %2d]\n", rec.movieTitle, rec.score, rec.nUsers)
	}
	fmt.Printf("\nExecution time: %v\n", end.Sub(start))
}
