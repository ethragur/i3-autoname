package main

import (
	"fmt"
	"os"
	"strings"
	"github.com/ethragur/i3ipc-go"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"flag"
)

func main() {
	db, err := sql.Open("sqlite3", "./gorename.db")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error could not create database: " + err.Error())
		return
	}
	defer db.Close()

	err = createWindowDB(db)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error could not create table: " + err.Error())
		return
	}

	insertPtr := flag.Bool("i", false, "Insert Icon Class Into DB")
	keyPtr	  := flag.String("class", "", "The Class of the Window")
	iconPtr  := flag.String("icon", "", "The Icon of the Window")

	flag.Parse()

	if *insertPtr {
		if *keyPtr != "" && *iconPtr != "" {
			err := insertIconPair(db, *keyPtr, *iconPtr)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error inserting Win/Icon pair: " + err.Error())
			} else {
				fmt.Println("Succesfully inserted Win/Icon pair: " + *keyPtr + " " + *iconPtr)
			}
		} else {
			fmt.Fprintln(os.Stderr, "Please enter Class and Icon with -class=... & -icon=...")
		}
		return
	}

	window_icons, err := getWindowIcons(db)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error could not read table: " + err.Error())
		return
	}

	i3ipc.StartEventListener()
	windowEvents, err := i3ipc.Subscribe(i3ipc.I3WindowEvent)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not subscribe to i3ipc: " + err.Error())
		return
	}

	ipcSocket, err := i3ipc.GetIPCSocket()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot get i3ipc Socket: ", err)
		return
	}

	for {

		event := <-windowEvents

		if event.Change == "close" || event.Change == "new" || event.Change == "move" {

			tree, err := ipcSocket.GetTree()

			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			//workspace, err := ipcSocket.GetWorkspaces();
			workspaces := GetWorkspaces(tree.Nodes...)
			for _, workspace := range workspaces {
				// give it a capacity of 4
				newWsName := fmt.Sprintf("%d:  ", workspace.Num)
				for _ , window := range GetWindows(workspace) {
					icon := window_icons[strings.ToLower(window.Window_Properties.Class)]
					if icon == "" {
						icon = "*"
					}
					newWsName += icon + "    "

				}
				ipcSocket.Command(fmt.Sprintf("rename workspace \"%s\" to \"%s  \"", workspace.Name, newWsName))
			}
		}
	}
}

func GetWorkspaces(Nodes ...i3ipc.I3Node) (workspaces []i3ipc.I3Node) {
	if len(Nodes) == 0 {
		return
	}
	for _, Node := range Nodes {
		//get All of type workspace execpt the internal __i3_scratch space
		if Node.Type == "workspace" && Node.Num != -1 {
			workspaces = append(workspaces, Node)
		} else {
			workspaces = append(workspaces, GetWorkspaces(Node.Nodes...)...)
		}
	}
	return
}

func GetWindows(Nodes ...i3ipc.I3Node) (windows []i3ipc.I3Node) {
	if len(Nodes) == 0 {
		return
	}
	for _, Node := range Nodes {
		//get All of type workspace execpt the internal __i3_scratch space
		if (Node.Type == "con" || Node.Type == "floating_con") && Node.Window > 0 {
			windows = append(windows, Node)
		} else {
			windows = append(windows, GetWindows(Node.Nodes...)...)
		}
	}
	return
}

func createWindowDB(db *sql.DB) (err error) {
	sqlStmt := "CREATE TABLE IF NOT EXISTS window_db(window_class TEXT, window_icon TEXT);" 
	_, err = db.Exec(sqlStmt)
	return err
}

func insertIconPair(db *sql.DB, class string, icon string) (err error) {
	sqlStmt := fmt.Sprintf("INSERT INTO window_db(window_class, window_icon) VALUES(\"%s\", \"%s\");", class, icon)
	_, err = db.Exec(sqlStmt)
	return err
}

func getWindowIcons(db *sql.DB) (window_infos map[string]string, err error) {
	window_infos = make(map[string]string)
	sqlStmt := "SELECT * FROM window_db;"
	rows, err := db.Query(sqlStmt)
	if err != nil {
		return window_infos, err
	}
	defer rows.Close()

	for rows.Next() {
		var winClass string
		var winIcon string
		err = rows.Scan(&winClass, &winIcon)
		if err != nil {
			return window_infos, err
		}
		window_infos[strings.ToLower(winClass)] = winIcon
	}
	return window_infos, rows.Err()
}

