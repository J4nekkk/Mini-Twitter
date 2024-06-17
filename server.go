package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	pb "github.com/J4nekkk/Mini-Twitterek"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTwitterServer
	db *sql.DB
}

func (s *server) SendTweet(ctx context.Context, in *pb.Tweet) (*pb.Empty, error) {
	_, err := s.db.Exec("INSERT INTO tweets (text) VALUES (?)", in.Text)
	if err != nil {
		log.Printf("Błąd podczas dodawania tweeta do bazy danych: %v", err)
		return nil, err
	}
	log.Printf("Dodano nowy tweet do bazy danych: %s\n", in.Text)
	return &pb.Empty{}, nil
}

func (s *server) GetTweet(in *pb.TweetLicz, stream pb.Twitter_GetTweetServer) error {
	rows, err := s.db.Query("SELECT text FROM tweets ORDER BY id DESC LIMIT ?", in.Liczba)
	if err != nil {
		log.Printf("Błąd podczas pobierania tweetów z bazy danych: %v", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			log.Printf("Błąd podczas skanowania wiersza z bazy danych: %v", err)
			return err
		}
		if err := stream.Send(&pb.Tweet{Text: text}); err != nil {
			log.Printf("Błąd podczas wysyłania tweeta przez strumień: %v", err)
			return err
		}
	}

	return nil
}

func main() {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/Tweety")
	if err != nil {
		log.Fatalf("Nie można połączyć się z bazą danych: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Nie można połączyć się z bazą danych: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTwitterServer(s, &server{db: db})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Nie można nasłuchiwać na porcie 50051: %v", err)
	}

	fmt.Println("Serwer Mini Twitterka uruchomiony na porcie 50051.")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Nie udało się obsłużyć: %v", err)
	}
}
