package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"strings"

	pb "github.com/penguinpoweredapps/todo/proto"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTodoServiceServer
	db *sql.DB
}

func (s *server) CreateTodo(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	id := uuid.New().String()
	t := req.GetTodo()

	_, err := s.db.Exec(`INSERT INTO todos (id, membership_id, description, completed, date_added, date_due, category) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, t.MembershipId, t.Description, t.Completed, t.DateAdded, t.DateDue, t.Category)

	if err != nil {
		return nil, err
	}
	return &pb.CreateResponse{Id: id}, nil
}

func (s *server) ReadTodo(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	row := s.db.QueryRow("SELECT id, membership_id, description, completed, date_added, date_due, category FROM todos WHERE id = ?", req.GetId())

	var t pb.Todo
	err := row.Scan(&t.Id, &t.MembershipId, &t.Description, &t.Completed, &t.DateAdded, &t.DateDue, &t.Category)
	if err != nil {
		return nil, err
	}
	return &pb.ReadResponse{Todo: &t}, nil
}

func (s *server) UpdateTodo(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	t := req.GetTodo()
	_, err := s.db.Exec(`UPDATE todos SET description=?, completed=?, date_due=?, category=? WHERE id=?`,
		t.Description, t.Completed, t.DateDue, t.Category, t.Id)
	return &pb.UpdateResponse{Success: err == nil}, err
}

func (s *server) DeleteTodo(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	_, err := s.db.Exec("DELETE FROM todos WHERE id=?", req.GetId())
	return &pb.DeleteResponse{Success: err == nil}, err
}

func (s *server) ReadAllTodos(ctx context.Context, req *pb.ReadAllRequest) (*pb.ReadAllResponse, error) {
	limit := req.GetLimit()
	if limit <= 0 {
		limit = 10
	}

	offset := req.GetOffset()
	if offset < 0 {
		offset = 0
	}

	var conditions []string
	var args []interface{}

	if req.GetSearchTerm() != "" {
		conditions = append(conditions, "description LIKE ?")
		args = append(args, "%"+req.GetSearchTerm()+"%")
	}
	if req.Completed != nil {
		conditions = append(conditions, "completed = ?")
		args = append(args, req.GetCompleted())
	}
	if req.Category != nil {
		conditions = append(conditions, "category = ?")
		args = append(args, req.GetCategory())
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCount int32
	countQuery := "SELECT COUNT(*) FROM todos" + whereClause
	if err := s.db.QueryRow(countQuery, args...).Scan(&totalCount); err != nil {
		return nil, err
	}

	orderByColumn := "date_added"
	if req.GetSortBy() == pb.SortBy_DATE_DUE {
		orderByColumn = "date_due"
	}

	sortDir := "DESC"
	if req.GetSortDirection() == pb.SortDirection_ASC {
		sortDir = "ASC"
	}

	orderByClause := " ORDER BY " + orderByColumn + " " + sortDir

	selectQuery := `
		SELECT id, membership_id, description, completed, date_added, date_due, category 
		FROM todos` + whereClause + orderByClause + ` LIMIT ? OFFSET ?`

	// Create a new slice to avoid mutating the original args slice's backing array
	selectArgs := append(append([]interface{}(nil), args...), limit, offset)

	rows, err := s.db.Query(selectQuery, selectArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []*pb.Todo
	for rows.Next() {
		var t pb.Todo
		err := rows.Scan(&t.Id, &t.MembershipId, &t.Description, &t.Completed, &t.DateAdded, &t.DateDue, &t.Category)
		if err != nil {
			return nil, err
		}
		todos = append(todos, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &pb.ReadAllResponse{Todos: todos, TotalCount: totalCount}, nil
}

func main() {
	db, err := sql.Open("sqlite3", "./todos.db")
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS todos (
		id TEXT PRIMARY KEY,
		membership_id TEXT,
		description TEXT,
		completed BOOLEAN,
		date_added TEXT,
		date_due TEXT,
		category INTEGER
	)`)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTodoServiceServer(s, &server{db: db})
	log.Printf("Server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
