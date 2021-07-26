package main

// cric module
//
// Copyright (c) 2020 - Valentin Kuznetsov <vkuznet@gmail.com>
//

import (
	"log"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dmwm/cmsauth"
)

// CricRecords list to hold CMS CRIC entries
var CricRecords cmsauth.CricRecords

// cmsRecords holds map of CricRecords for CMS users
var cmsRecords cmsauth.CricRecords

// cmsRecordsLock keeps lock for cmsRecords updates
var cmsRecordsLock sync.RWMutex

// int pattern
var intPattern = regexp.MustCompile(`^\d+$`)

// UpdateCricRecords periodically updates cric records
// should be run as goroutine
func UpdateCricRecords(key, cricFile, cricURL string, cricUpdateInterval int64, cricVerbose int) {
	log.Println("update cric records with", key, "as a key")
	var err error
	cricRecords := make(cmsauth.CricRecords)
	verbose := false
	if cricVerbose > 0 {
		verbose = true
	}
	// if cric file is given read it first, then if we have
	// cric url we'll update it from there
	if cricFile != "" {
		if key == "id" {
			cricRecords, err = cmsauth.ParseCricByKey(cricFile, "id", verbose)
		} else {
			cricRecords, err = cmsauth.ParseCric(cricFile, verbose)
		}
		log.Printf("obtain CRIC records from %s, %v", cricFile, err)
		if err != nil {
			log.Printf("Unable to update CRIC records: %v", err)
		} else {
			CricRecords = cricRecords
			keys := reflect.ValueOf(CricRecords).MapKeys()
			log.Println("Updated CRIC records", len(keys))
			if key == "id" {
				cmsRecords = cricRecords
			} else {
				updateCMSRecords(cricRecords)
			}
			log.Println("Updated cms records", len(cmsRecords))
		}
	}
	for {
		interval := cricUpdateInterval
		if interval == 0 {
			interval = 3600
		}
		// parse cric records
		if cricURL != "" {
			if key == "id" {
				cricRecords, err = cmsauth.GetCricDataByKey(cricURL, "id", verbose)
			} else {
				cricRecords, err = cmsauth.GetCricData(cricURL, verbose)
			}
			log.Printf("obtain CRIC records from %s, %v", cricURL, err)
		} else if cricFile != "" {
			if key == "id" {
				cricRecords, err = cmsauth.ParseCricByKey(cricFile, "id", verbose)
			} else {
				cricRecords, err = cmsauth.ParseCric(cricFile, verbose)
			}
			log.Printf("obtain CRIC records from %s, %v", cricFile, err)
		} else {
			log.Println("Untable to get CRIC records no file or no url was provided")
		}
		if err != nil {
			log.Printf("Unable to update CRIC records: %v", err)
		} else {
			CricRecords = cricRecords
			keys := reflect.ValueOf(CricRecords).MapKeys()
			log.Println("Updated CRIC records", len(keys))
			if key == "id" {
				cmsRecords = cricRecords
			} else {
				updateCMSRecords(cricRecords)
			}
			log.Println("Updated cms records", len(cmsRecords))
			if cricVerbose > 2 {
				for k, v := range cmsRecords {
					log.Printf("key=%s value=%s record=%+v\n", key, k, v)
				}
			} else if cricVerbose > 0 {
				for k, v := range cmsRecords {
					log.Printf("key=%s value=%s record=%+v\n", key, k, v)
					break // break to avoid lots of CRIC record printous
				}
			}
		}
		d := time.Duration(interval) * time.Second
		time.Sleep(d) // sleep for next iteration
	}
}

// helper function to create cmsRecords
func updateCMSRecords(cricRecords cmsauth.CricRecords) {
	cmsRecordsLock.Lock()
	defer cmsRecordsLock.Unlock()
	cmsRecords = make(cmsauth.CricRecords)
	for _, r := range cricRecords {
		for _, dn := range r.DNs {
			for _, v := range strings.Split(dn, "/CN=") {
				if !strings.HasPrefix(v, "/") {
					if matched := intPattern.MatchString(v); !matched {
						cmsRecords[v] = r
					}
				}
			}
		}
	}
}
