package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// buildCache builds up the cache.
// p is the path that we need to walk through
// cache is the caching instance that will be filled
// while we walk through p recursively.
//
// p should point to a folder structure that contains
// games. This should look like
// games
// ├── Season_1
// │   ├── 2019-02-15
// │   │   ├── game.json
// │   │   └── overview.json
// │   ├── 2019-02-18
// │   │   ├── game.json
// │   │   └── overview.json
// │   ├── 2019-02-19
// │   │   ├── game.json
// │   │   └── overview.json
// │   └── overview.json
// └── Season_2
//     └── overview.json
//     ...
func buildCache(p string, cache *cache) error {
	err := filepath.Walk(p, visit(cache))
	if err != nil {
		return err
	}

	return nil
}

// visit is representing a single step while walking
// through a folder structure via filepath.Walk.
// cache is the instance where we write the file content into.
//
// We are only interested in overview|game.json files.
func visit(cache *cache) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".json" {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}

			// Attention. Now we enter a quite hacky sequence.
			// I suggest good body protection. Stay safe!

			// We check if the path contains only two slashes
			// Two slashes look like a path like
			// "games/Season_1/overview.json"
			// Means we are reading a season folder.
			//
			// We assume a lot here. But hey, it works on my machine.
			if n := strings.Count(path, string(os.PathSeparator)); n == 2 {
				if strings.HasSuffix(path, "overview.json") {
					cache.SeasonCache[path] = data
				}
			}

			// We check if the path contains now three slashes
			// Three slashes mean we are one level deeper (Yep, this is an Inception pun -> https://9gag.com/gag/amBAZwv)
			// right into a single game like "games/Season_1/2019-02-19/overview.json"
			//
			// PS: Do you like Inception? Check out this -> https://9gag.com/gag/a3KOxqe/michael-caine-explains-the-ending-of-inception-after-8-years
			if n := strings.Count(path, string(os.PathSeparator)); n == 3 {
				cache.GameCache[path] = data
			}
		}
		return nil
	}
}
