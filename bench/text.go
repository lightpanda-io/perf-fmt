package bench

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lightpanda-io/perf-fmt/git"
)

var (
	ErrBadData = errors.New("bad data format")
	ErrBadName = errors.New("bad name format")
)

func ParseTxtData(filename string, data []byte) (OutResult, error) {
	datetime, commit, err := parseTxtName(filename)
	if err != nil {
		return OutResult{}, fmt.Errorf("parse filename: %s", err)
	}

	res := OutResult{
		Hash: commit,
		Time: datetime,
	}

	var (
		i    = 0
		rest = data
		ok   bool
		v    []byte
	)
LOOP:
	for {
		v, rest, ok = bytes.Cut(rest, []byte("\n"))
		if !ok {
			return OutResult{}, ErrBadData
		}

		switch i {
		case 7:
			item, err := parseLine(v)
			if err != nil {
				return OutResult{}, fmt.Errorf("bad data format: %w", err)
			}
			res.Data.WithIsolate = item
		case 9:
			item, err := parseLine(v)
			if err != nil {
				return OutResult{}, fmt.Errorf("bad data format: %w", err)
			}
			res.Data.WithoutIsolate = item
			break LOOP
		}
		i += 1
	}

	return res, nil
}

var linerxp = regexp.MustCompile(`^ *\|[^|]+\| +(\d+)us[^|]+\| +(\d+)[^|]+\|(| +(\d+)[^|]+\|)? +(\d+)kb[^|]+\|$`)

func parseLine(data []byte) (OutItem, error) {
	b := linerxp.FindSubmatch(data)
	if len(b) != 6 {
		return OutItem{}, ErrBadData
	}

	duration, err := strconv.Atoi(string(b[1]))
	if err != nil {
		return OutItem{}, fmt.Errorf("bad data format: %w", err)
	}

	reallocnb := 0
	if len(b[4]) > 0 {
		reallocnb, err = strconv.Atoi(string(b[4]))
		if err != nil {
			return OutItem{}, fmt.Errorf("bad data format: %w", err)
		}
	}

	allocnb, err := strconv.Atoi(string(b[2]))
	if err != nil {
		return OutItem{}, fmt.Errorf("bad data format: %w", err)
	}

	allocsize, err := strconv.Atoi(string(b[5]))
	if err != nil {
		return OutItem{}, fmt.Errorf("bad data format: %w", err)
	}

	return OutItem{
		Duration:  duration,
		ReallocNb: reallocnb,
		AllocNb:   allocnb,
		AllocSize: allocsize,
	}, nil
}

func parseTxtName(filename string) (time.Time, git.CommitHash, error) {
	strdate, rest, ok := strings.Cut(filename, "_")
	if !ok {
		return time.Time{}, "", ErrBadName
	}

	strtime, rest, ok := strings.Cut(rest, "_")
	if !ok {
		return time.Time{}, "", ErrBadName
	}

	commit, rest, ok := strings.Cut(rest, "_")
	if !ok {
		return time.Time{}, "", ErrBadName
	}

	// parse time
	date, err := time.Parse("2006-01-02 15-04", strdate+" "+strtime)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("bad date format: %w", err)
	}

	return date, git.CommitHash(commit), nil
}
