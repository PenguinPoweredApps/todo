package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/penguinpoweredapps/todo/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func parseCategory(cat string) pb.Category {
	if strings.ToUpper(cat) == "PERSONAL" {
		return pb.Category_PERSONAL
	}
	return pb.Category_WORK
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Expected 'create', 'read', 'update', 'delete', or 'list' subcommands")
		os.Exit(1)
	}

	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewTodoServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createMem := createCmd.String("membership", "MEMB-001", "Membership ID")
	createDesc := createCmd.String("desc", "", "Todo description (Required)")
	createDue := createCmd.String("due", time.Now().AddDate(0, 0, 1).Format(time.RFC3339), "Due date")
	createCat := createCmd.String("category", "WORK", "Category: WORK or PERSONAL")

	readCmd := flag.NewFlagSet("read", flag.ExitOnError)
	readId := readCmd.String("id", "", "Todo ID (Required)")

	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
	updateId := updateCmd.String("id", "", "Todo ID (Required)")
	updateDesc := updateCmd.String("desc", "", "New description")
	updateComp := updateCmd.Bool("completed", false, "Mark as completed")
	updateDue := updateCmd.String("due", "", "New due date")
	updateCat := updateCmd.String("category", "", "New category: WORK or PERSONAL")

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteId := deleteCmd.String("id", "", "Todo ID (Required)")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listLimit := listCmd.Int("limit", 10, "Number of items to return")
	listOffset := listCmd.Int("offset", 0, "Number of items to skip")
	listSearch := listCmd.String("search", "", "Filter by description text")
	listCompleted := listCmd.String("completed", "", "Filter by status: true or false")
	listCategory := listCmd.String("category", "", "Filter by category: WORK or PERSONAL")
	listSortBy := listCmd.String("sort", "added", "Sort by: 'added' or 'due'")
	listAsc := listCmd.Bool("asc", false, "Sort in ascending order (default is descending)")

	switch os.Args[1] {
	case "create":
		createCmd.Parse(os.Args[2:])
		if *createDesc == "" {
			log.Fatal("Description is required to create a Todo")
		}

		res, err := c.CreateTodo(ctx, &pb.CreateRequest{
			Todo: &pb.Todo{
				MembershipId: *createMem,
				Description:  *createDesc,
				Completed:    false,
				DateAdded:    time.Now().Format(time.RFC3339),
				DateDue:      *createDue,
				Category:     parseCategory(*createCat),
			},
		})
		if err != nil {
			log.Fatalf("Error creating todo: %v", err)
		}
		fmt.Printf("✅ Successfully created Todo with ID: %s\n", res.GetId())

	case "read":
		readCmd.Parse(os.Args[2:])
		if *readId == "" {
			log.Fatal("ID is required to read a Todo")
		}

		res, err := c.ReadTodo(ctx, &pb.ReadRequest{Id: *readId})
		if err != nil {
			log.Fatalf("Error reading todo: %v", err)
		}
		t := res.GetTodo()
		fmt.Printf("Todo ID: %s\nMembership: %s\nDescription: %s\nCompleted: %t\nAdded: %s\nDue: %s\nCategory: %s\n",
			t.Id, t.MembershipId, t.Description, t.Completed, t.DateAdded, t.DateDue, t.Category.String())

	case "update":
		updateCmd.Parse(os.Args[2:])
		if *updateId == "" {
			log.Fatal("ID is required to update a Todo")
		}

		readRes, err := c.ReadTodo(ctx, &pb.ReadRequest{Id: *updateId})
		if err != nil {
			log.Fatalf("Error finding existing todo: %v", err)
		}
		existingTodo := readRes.GetTodo()

		if *updateDesc != "" {
			existingTodo.Description = *updateDesc
		}
		if *updateDue != "" {
			existingTodo.DateDue = *updateDue
		}
		if *updateCat != "" {
			existingTodo.Category = parseCategory(*updateCat)
		}
		existingTodo.Completed = *updateComp

		_, err = c.UpdateTodo(ctx, &pb.UpdateRequest{Todo: existingTodo})
		if err != nil {
			log.Fatalf("Error updating todo: %v", err)
		}
		fmt.Printf("✅ Successfully updated Todo ID: %s\n", *updateId)

	case "delete":
		deleteCmd.Parse(os.Args[2:])
		if *deleteId == "" {
			log.Fatal("ID is required to delete a Todo")
		}

		_, err = c.DeleteTodo(ctx, &pb.DeleteRequest{Id: *deleteId})
		if err != nil {
			log.Fatalf("Error deleting todo: %v", err)
		}
		fmt.Printf("✅ Successfully deleted Todo ID: %s\n", *deleteId)

	case "list":
		listCmd.Parse(os.Args[2:])

		sortBy := pb.SortBy_DATE_ADDED
		if strings.ToLower(*listSortBy) == "due" {
			sortBy = pb.SortBy_DATE_DUE
		}
		sortDir := pb.SortDirection_DESC
		if *listAsc {
			sortDir = pb.SortDirection_ASC
		}

		req := &pb.ReadAllRequest{
			Limit:         int32(*listLimit),
			Offset:        int32(*listOffset),
			SearchTerm:    *listSearch,
			SortBy:        sortBy,
			SortDirection: sortDir,
		}

		if *listCompleted != "" {
			isCompleted := strings.ToLower(*listCompleted) == "true"
			req.Completed = &isCompleted
		}
		if *listCategory != "" {
			cat := parseCategory(*listCategory)
			req.Category = &cat
		}

		res, err := c.ReadAllTodos(ctx, req)
		if err != nil {
			log.Fatalf("Error reading todos: %v", err)
		}

		todos := res.GetTodos()

		fmt.Printf("📋 Showing %d of %d Total Todos (Limit: %d, Offset: %d):\n",
			len(todos), res.GetTotalCount(), *listLimit, *listOffset)
		fmt.Println(strings.Repeat("=", 60))

		if len(todos) == 0 {
			fmt.Println("No todos found matching your criteria.")
			return
		}

		const (
			colorReset = "\033[0m"
			colorRed   = "\033[31m"
			colorGreen = "\033[32m"
			colorBlue  = "\033[34m"
		)

		var workTodos, personalTodos []*pb.Todo
		for _, t := range todos {
			if t.Category == pb.Category_WORK {
				workTodos = append(workTodos, t)
			} else {
				personalTodos = append(personalTodos, t)
			}
		}

		printTodos := func(list []*pb.Todo) {
			now := time.Now()
			for _, t := range list {
				status := " "
				if t.Completed {
					status = "x"
				}

				dueStr := t.DateDue
				dueDate, err := time.Parse(time.RFC3339, t.DateDue)
				if err == nil && !t.Completed && dueDate.Before(now) {
					dueStr = colorRed + t.DateDue + colorReset
				}

				fmt.Printf("[%s] ID: %s\n    %s (Due: %s)\n", status, t.Id, t.Description, dueStr)
			}
		}

		if len(workTodos) > 0 {
			fmt.Printf("\n%s👔 WORK TASKS%s\n", colorBlue, colorReset)
			fmt.Println(strings.Repeat("-", 20))
			printTodos(workTodos)
		}
		if len(personalTodos) > 0 {
			fmt.Printf("\n%s🏠 PERSONAL TASKS%s\n", colorGreen, colorReset)
			fmt.Println(strings.Repeat("-", 20))
			printTodos(personalTodos)
		}
		fmt.Println()

	default:
		fmt.Println("Expected 'create', 'read', 'update', 'delete', or 'list' subcommands")
		os.Exit(1)
	}
}
