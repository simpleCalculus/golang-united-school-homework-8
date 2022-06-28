package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type Users []User

type Arguments map[string]string

var (
	ErrOperation = errors.New("-operation flag has to be specified")
	ErrItem      = errors.New("-item flag has to be specified")
	ErrFlagName  = errors.New("-fileName flag has to be specified")
	ErrId        = errors.New("-id flag has to be specified")
)

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

func Perform(args Arguments, writer io.Writer) error {
	operation := args["operation"]
	if operation == "" {
		return ErrOperation
	}

	fileName := args["fileName"]
	if fileName == "" {
		return ErrFlagName
	}

	switch operation {
	case "list":
		err := getList(fileName, writer)
		return err
	case "add":
		item := args["item"]
		if item == "" {
			return ErrItem
		}
		err := addUser(item, fileName, writer)
		return err
	case "findById":
		id := args["id"]
		if id == "" {
			return ErrId
		}
		err := findUserById(id, fileName, writer)
		return err
	case "remove":
		id := args["id"]
		if id == "" {
			return ErrId
		}
		err := removeUser(id, fileName, writer)
		return err
	default:
		return errors.New(fmt.Sprintf("Operation %s not allowed!", args["operation"]))
	}
}

func parseArgs() Arguments {
	operation := flag.String("operation", "", "type of operation")
	id := flag.String("id", "", "user id")
	fileName := flag.String("fileName", "", "name of file")
	item := flag.String("item", "", "item")

	flag.Parse()

	return Arguments{
		"operation": *operation,
		"id":        *id,
		"item":      *item,
		"fileName":  *fileName,
	}
}

func getList(fileName string, writer io.Writer) error {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	_, err = writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func getUsersArray(fileName string) (Users, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	users := Users{}
	json.Unmarshal(bytes, &users)

	return users, nil
}

func addUser(item string, fileName string, writer io.Writer) error {
	var newUser User
	if err := json.Unmarshal([]byte(item), &newUser); err != nil {
		return fmt.Errorf("can't unmarshal json, error = %w", err)
	}

	users, err := getUsersArray(fileName)
	if err != nil {
		return fmt.Errorf("can't get users array from json file, error = %w", err)
	}

	for _, u := range users {
		if u.Id == newUser.Id {
			writer.Write([]byte(fmt.Sprintf("Item with id %s already exists", newUser.Id)))
			return nil
		}
	}

	users = append(users, newUser)

	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("can't open file, error = %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("can't marshal users slice, error = %w", err)
	}

	file.Write(data)

	return nil
}

func findUserById(id string, fileName string, writer io.Writer) error {
	users, err := getUsersArray(fileName)
	if err != nil {
		return fmt.Errorf("can't get users array from json file, error = %w", err)
	}

	var foundUser *User

	for _, u := range users {
		if u.Id == id {
			foundUser = &u
			break
		}
	}

	if foundUser == nil {
		return nil
	}

	data, err := json.Marshal(foundUser)
	if err != nil {
		return fmt.Errorf("can't marshal found user, error = %w", err)
	}

	writer.Write(data)

	return nil
}

func removeUser(id string, fileName string, writer io.Writer) error {
	users, err := getUsersArray(fileName)
	if err != nil {
		return fmt.Errorf("can't get users array from json file, error = %w", err)
	}

	var userForRemove *User
	var index int

	for i, u := range users {
		if u.Id == id {
			userForRemove = &u
			index = i
			break
		}
	}

	if userForRemove == nil {
		writer.Write([]byte(fmt.Sprintf("Item with id %s not found", id)))
		return nil
	}

	users = append(users[:index], users[index+1:]...)
	os.Remove(fileName)

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("can't open or create file, error = %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("can't marshal users slice, error = %w", err)
	}

	file.Write(data)

	return nil
}
