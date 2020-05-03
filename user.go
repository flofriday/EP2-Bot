package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"sync"
)

var (
	// users is map only because go has no good set implementation
	users = make(map[int64]bool, 0)

	userMutex = sync.Mutex{}
	userFile  = path.Join("data", "users.json")
)

// Load the users from the disk.
// This function should be called once
func loadUsers() error {
	// Lock the mutex to ensure only one is modifying the data
	userMutex.Lock()
	defer userMutex.Unlock()

	// Read the file
	byteValue, err := ioutil.ReadFile(userFile)
	if err != nil {
		log.Println("The user file does not exist")
		return nil
	}

	// Parse the file
	err = json.Unmarshal(byteValue, &users)
	if err != nil {
		return err
	}

	return nil
}

// Save the current userlist to the disk
// Note: the caller must lock the userMutex to avoid race conditions
func saveUsers() error {
	// Create the content
	byteValue, err := json.Marshal(users)
	if err != nil {
		return err
	}

	// Write the file
	return ioutil.WriteFile(userFile, byteValue, 0777)
}

func getUsers() []int64 {
	userMutex.Lock()
	defer userMutex.Unlock()

	result := make([]int64, 0, len(users))
	for k, _ := range users {
		result = append(result, k)
	}
	return result
}

// Returns true if the specified userID is in the list of subscribed users
func isUser(user int64) bool {
	userMutex.Lock()
	defer userMutex.Unlock()

	_, ok := users[user]
	return ok
}

// Add the specified user to the list of subscribed users
func addUser(user int64) error {
	userMutex.Lock()
	defer userMutex.Unlock()

	users[user] = true
	return saveUsers()
}

// Remove the specified user from the list
func removeUser(user int64) error {
	userMutex.Lock()
	defer userMutex.Unlock()

	delete(users, user)
	return saveUsers()
}
