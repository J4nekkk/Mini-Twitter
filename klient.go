package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	pb "github.com/J4nekkk/Mini-Twitterek"
	"google.golang.org/grpc"
)

const (
	maxTweets = 10
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Nie można się połączyć: %v", err)
	}
	defer conn.Close()
	c := pb.NewTwitterClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reader := bufio.NewReader(os.Stdin)
MainLoop:
	for {
		fmt.Println("Mini Twitter")
		fmt.Println("1. Wyślij Tweeta")
		fmt.Println("2. Pobierz N liczbe ostatnich Tweetów")
		fmt.Println("3. Wyjście")
		fmt.Print("Wybierz: ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Błąd podczas wybierania wejścia:", err)
			break
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Println("Wprowadź wiadomość (maksymalnie 80 znaków):")
			message, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Błąd podczas wybierania wejścia:", err)
				break
			}
			message = strings.TrimSpace(message)
			if len(message) > 80 {
				log.Println("Wiadomość przekracza 80 znaków. Proszę napisać krótszą wiadomość.")
				continue
			}

			_, err = c.SendTweet(ctx, &pb.Tweet{Text: message})
			if err != nil {
				log.Println("Błąd podczas wysyłania Tweeta:", err)
			} else {
				fmt.Println("Tweet został wysłany!")
			}
		case "2":
			numStr := getNumOfTweets(reader)
			num, err := strconv.Atoi(numStr)
			if err != nil || num < 1 {
				log.Println("Błąd: Nieprawidłowa liczba tweetów.")
				continue
			}

			tweets, err := getTweetsFromDB(c, ctx, num)
			if err != nil {
				log.Println("Błąd podczas pobierania tweetów:", err)
				continue
			}

			displayTweets(tweets)
		case "3":
			cancel()
			break MainLoop
		default:
			fmt.Println("Nieprawidłowy wybór.")
		}
	}

	fmt.Println("Zamykanie klienta.")
}

func getNumOfTweets(reader *bufio.Reader) string {
	fmt.Println("Podaj liczbę Tweetów do pobrania (maksymalnie", maxTweets, "):")
	numStr, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Błąd podczas wybierania wejścia:", err)
		return "0"
	}
	return strings.TrimSpace(numStr)
}

func getTweetsFromDB(c pb.TwitterClient, ctx context.Context, num int) ([]string, error) {
	stream, err := c.GetTweet(ctx, &pb.TweetLicz{Liczba: int32(num)})
	if err != nil {
		return nil, err
	}

	var tweets []string
	for {
		tweet, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		tweets = append(tweets, tweet.Text)
	}

	return tweets, nil
}

func displayTweets(tweets []string) {
	if len(tweets) == 0 {
		fmt.Println("Brak Tweetów.")
		return
	}

	fmt.Println("Tweety:")
	for _, tweet := range tweets {
		fmt.Println("-", tweet)
	}
}
