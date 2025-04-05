# Movie Recommendation System

A Go-based movie recommendation system that uses collaborative filtering to suggest movies to users based on their preferences and similar users' ratings.

## Overview

This project implements a movie recommendation system using the Jaccard similarity measure and collaborative filtering techniques. It processes movie ratings data to generate personalized movie recommendations for users.

## Features

- Collaborative filtering using Jaccard similarity
- Parallel processing using Go's concurrency features
- Multi-stage recommendation pipeline
- Filters out already seen movies
- Considers both liked and disliked movies
- Minimum threshold for movie popularity (K users)
- Configurable recommendation parameters

## Prerequisites

- Go 1.x or higher
- CSV files containing movie and rating data:
  - `movies.csv`: Contains movie IDs and titles
  - `ratings.csv`: Contains user ratings for movies

## Project Structure

```
.
├── projectMovieRec.go    # Main application code
├── movies.csv            # Movie database
├── ratings.csv           # User ratings data
├── User*.txt            # User-specific data files
└── printGrapth/         # Directory for graph visualization
```

## Configuration

The system uses several configurable constants:
- `iLiked`: Threshold for "liked" movies (default: 3.5)
- `K`: Minimum number of users who must like a movie (default: 10)
- `N`: Number of top recommendations to keep (default: 20)

## Usage

1. Ensure you have the required CSV files in the project directory
2. Run the program:
   ```bash
   go run projectMovieRec.go
   ```

## Implementation Details

The recommendation system uses a multi-stage pipeline:
1. Generate all possible movie recommendations
2. Filter out movies the user has already seen
3. Filter out movies liked by fewer than K users
4. Compute recommendation scores using Jaccard similarity
5. Collect and sort the top N recommendations


## License

This project is part of CSI2520 Programming Paradigms course work. 