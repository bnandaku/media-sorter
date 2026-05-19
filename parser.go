package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// TV show regex patterns
var (
	tvPatternStandard = regexp.MustCompile(`(?i)s(\d{1,2})e(\d{1,2})`)
	tvPatternSpelled  = regexp.MustCompile(`(?i)season\s+(\d{1,2})\s*-\s*(\d{1,3})`)
	tvPatternSimple   = regexp.MustCompile(`\s-\s+(\d{1,3})(?:\D|$)`)
	bracketsPattern   = regexp.MustCompile(`\[[^\]]*\]`)
	yearPattern       = regexp.MustCompile(`\s*\((?:19|20)\d{2}\)`)
)

// TVShowInfo holds parsed TV show episode data.
type TVShowInfo struct {
	ShowName     string
	Season       string
	Episode      string
	OriginalName string
	QualityInfo  string
	Extension    string
	HasSeasonInfo bool
	StandardName string // ShowName.SXXEXX.quality.ext
}

// MovieInfo holds parsed movie data.
type MovieInfo struct {
	Title        string
	Year         string
	Quality      string
	OriginalName string
	Extension    string
	HasMovieInfo bool
}

func parseTVShowInfo(filename string) TVShowInfo {
	info := TVShowInfo{OriginalName: filename}

	var matches []int
	var showNameRaw, seasonNum, episodeNum string

	if m := tvPatternStandard.FindStringSubmatchIndex(filename); m != nil {
		matches = m
		showNameRaw = filename[:m[0]]
		seasonNum = filename[m[2]:m[3]]
		episodeNum = filename[m[4]:m[5]]
	} else if m := tvPatternSpelled.FindStringSubmatchIndex(filename); m != nil {
		matches = m
		showNameRaw = filename[:m[0]]
		seasonNum = filename[m[2]:m[3]]
		episodeNum = filename[m[4]:m[5]]
	} else if m := tvPatternSimple.FindStringSubmatchIndex(filename); m != nil {
		matches = m
		showNameRaw = filename[:m[0]]
		seasonNum = "01"
		episodeNum = filename[m[2]:m[3]]
	} else {
		return info
	}

	showName := bracketsPattern.ReplaceAllString(showNameRaw, "")
	showName = yearPattern.ReplaceAllString(showName, "")
	showName = strings.ReplaceAll(showName, ".", "_")
	showName = strings.ReplaceAll(showName, " ", "_")
	showName = strings.Trim(showName, "_-")
	for strings.Contains(showName, "__") {
		showName = strings.ReplaceAll(showName, "__", "_")
	}
	showName = toTitleCase(showName)

	if len(seasonNum) == 1 {
		seasonNum = "0" + seasonNum
	}
	if len(episodeNum) == 1 {
		episodeNum = "0" + episodeNum
	}

	extension := filepath.Ext(filename)
	filenameNoExt := strings.TrimSuffix(filename, extension)

	qualityInfo := ""
	if matches != nil && len(matches) > 1 {
		after := filenameNoExt[matches[1]:]
		after = strings.TrimLeft(after, ".- ")
		after = strings.Trim(after, ".")
		qualityInfo = after
	}

	standardName := showName + ".S" + seasonNum + "E" + episodeNum
	if qualityInfo != "" {
		standardName += "." + qualityInfo
	}
	standardName += extension

	info.ShowName = showName
	info.Season = seasonNum
	info.Episode = episodeNum
	info.QualityInfo = qualityInfo
	info.Extension = extension
	info.StandardName = standardName
	info.HasSeasonInfo = true
	return info
}

// buildTVShowPath creates ShowName/Season_XX/ under basePath and returns the full dir path.
func buildTVShowPath(basePath string, info TVShowInfo) (string, error) {
	if !info.HasSeasonInfo {
		return basePath, nil
	}
	seasonDir := filepath.Join(basePath, info.ShowName, fmt.Sprintf("Season_%s", info.Season))
	if err := os.MkdirAll(seasonDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", seasonDir, err)
	}
	return seasonDir, nil
}

func parseMovieInfo(filename string) MovieInfo {
	info := MovieInfo{OriginalName: filename}

	extension := filepath.Ext(filename)
	info.Extension = extension
	nameNoExt := strings.TrimSuffix(filename, extension)

	yearRe := regexp.MustCompile(`(?:^|\.)(\d{4})(?:\.|$)`)
	yearMatches := yearRe.FindStringSubmatch(nameNoExt)

	var titlePart, qualityPart, year string

	if len(yearMatches) >= 2 {
		y, _ := strconv.Atoi(yearMatches[1])
		if y >= 1900 && y <= 2099 {
			year = yearMatches[1]
			parts := strings.SplitN(nameNoExt, year, 2)
			titlePart = strings.Trim(parts[0], ".")
			if len(parts) > 1 {
				qualityPart = strings.Trim(parts[1], ".")
			}
		}
	}

	if year == "" {
		qualityRe := regexp.MustCompile(`(?i)\.(720p|1080p|2160p|4K|BluRay|BrRip|WEB-DL|WEBRip|HDTV|x264|x265|HEVC|10bit).*$`)
		if loc := qualityRe.FindStringIndex(nameNoExt); loc != nil {
			titlePart = nameNoExt[:loc[0]]
			qualityPart = nameNoExt[loc[0]+1:]
		} else {
			titlePart = nameNoExt
		}
	}

	title := bracketsPattern.ReplaceAllString(titlePart, "")
	title = strings.ReplaceAll(title, ".", " ")
	title = strings.ReplaceAll(title, "_", " ")
	title = strings.TrimSpace(title)

	quality := strings.ReplaceAll(qualityPart, ".", " ")
	quality = strings.TrimSpace(quality)

	info.Title = title
	info.Year = year
	info.Quality = quality
	info.HasMovieInfo = title != ""
	return info
}

// buildMovieFilename creates a standardised movie filename from parsed info.
func buildMovieFilename(info MovieInfo) string {
	name := strings.ReplaceAll(info.Title, " ", ".")
	if info.Year != "" {
		name += "." + info.Year
	}
	if info.Quality != "" {
		name += "." + strings.ReplaceAll(info.Quality, " ", ".")
	}
	return name + info.Extension
}

func toTitleCase(s string) string {
	words := strings.Split(s, "_")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, "_")
}
