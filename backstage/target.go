package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/tsuru/tsuru/fs"
	"gopkg.in/v1/yaml"
)

var (
	TargetFileName       = joinHomePath(".backstage_targets")
	ErrLabelExists       = errors.New("The label provided exists already.")
	ErrLabelNotFound     = errors.New("Label not found.")
	ErrBadFormattedFile  = errors.New("Bad formatted file. Please open an issue on or Github page: backstage/backstage")
	ErrCommandCancelled  = errors.New("Command Cancelled.")
	ErrFailedWritingFile = errors.New("Failed trying to write the target file.")
)

var fsystem fs.Fs

func filesystem() fs.Fs {
	if fsystem == nil {
		fsystem = fs.OsFs{}
	}
	return fsystem
}

type Target struct {
	Current string
	Options map[string]string
}

func (t *Target) GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:        "target-add",
			Usage:       "target-add <label> <endpoint>",
			Description: "Adds a new target in the list of targets.",
			Action: func(c *cli.Context) {
				defer RecoverStrategy("target-add")()
				targets, err := LoadTargets()
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				args := c.Args()
				label, endpoint := args[0], args[1]
				err = targets.add(label, endpoint)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Println("Target added successfully!")
			},
		},
		{
			Name:        "target-list",
			Usage:       "",
			Description: "Adds a new target in the list of targets.",
			Action: func(c *cli.Context) {
				targets, err := LoadTargets()
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Println(targets.list())
			},
		},
		{
			Name:        "target-remove",
			Usage:       "target-remove <label>",
			Description: "Remove a target from the list of targets.",
			Before: func(c *cli.Context) error {
				if c.Args().First() == "" {
					return ErrCommandCancelled
				}
				context := &Context{Stdout: os.Stdout, Stdin: os.Stdin}
				if Confirm(context, "Are you sure you want to remove this target?") != true {
					return ErrCommandCancelled
				}
				return nil
			},
			Action: func(c *cli.Context) {
				defer RecoverStrategy("target-remove")()
				targets, err := LoadTargets()
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				label := c.Args()[1]
				err = targets.remove(label)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Println("Target removed successfully!")
			},
		},
		{
			Name:        "target-set",
			Usage:       "target-set <label>",
			Description: "Set a target as default to be used.",
			Action: func(c *cli.Context) {
				defer RecoverStrategy("target-set")()
				targets, err := LoadTargets()
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				label := c.Args().First()
				err = targets.setDefault(label)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Println("You have a new target as default!")
			},
		},
	}
}

func (t *Target) add(label string, endpoint string) error {
	if _, ok := t.Options[label]; ok {
		return ErrLabelExists
	}
	t.Options[label] = endpoint
	return t.save()
}

func (t *Target) list() string {
	var targetList bytes.Buffer
	for label, endpoint := range t.Options {
		if t.Current == label {
			targetList.WriteString("* ")
		}
		targetList.WriteString(label + " - " + endpoint + "\n")
	}
	return targetList.String()
}

func (t *Target) remove(label string) error {
	if _, ok := t.Options[label]; !ok {
		return ErrLabelNotFound
	}
	if t.Current == label {
		t.Current = ""
	}
	delete(t.Options, label)
	return t.save()
}

func (t *Target) setDefault(label string) error {
	if _, ok := t.Options[label]; !ok {
		return ErrLabelNotFound
	}
	t.Current = label
	return t.save()
}

func (t *Target) save() error {
	d, err := yaml.Marshal(&t)
	if err != nil {
		return err
	}
	targetsFile, err := filesystem().OpenFile(TargetFileName, syscall.O_RDWR|syscall.O_CREAT, 0600)
	if err != nil {
		return err
	}
	n, err := targetsFile.WriteString(string(d))
	if n != len(string(d)) || err != nil {
		return ErrFailedWritingFile
	}
	return nil
}

func LoadTargets() (*Target, error) {
	targetsFile, err := filesystem().OpenFile(TargetFileName, syscall.O_RDWR|syscall.O_CREAT, 0600)
	if err != nil {
		return nil, err
	}
	defer targetsFile.Close()
	data, err := ioutil.ReadAll(targetsFile)
	if err == nil {
		var t Target
		err = yaml.Unmarshal([]byte(data), &t)
		if err != nil {
			return nil, ErrBadFormattedFile
		}
		if t.Options == nil {
			t.Options = map[string]string{}
		}
		return &t, nil
	}
	return nil, err
}
